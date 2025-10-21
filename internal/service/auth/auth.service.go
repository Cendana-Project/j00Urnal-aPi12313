package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/scrypt"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	rolerepo "github.com/api-monolith-template/internal/repository/role"
	userrepo "github.com/api-monolith-template/internal/repository/user"
)

type EmailSender interface {
	Send(to, subject, htmlBody string) error
}

type Service struct {
	users *userrepo.Repository
	roles *rolerepo.Repository
	redis *redis.Client
	email EmailSender

	loc        *time.Location
	pinTTL     time.Duration
	accessTTL  time.Duration
	refreshTTL time.Duration
	jwtSecret  []byte
}

func NewService(users *userrepo.Repository, roles *rolerepo.Repository, rdb *redis.Client, sender EmailSender) *Service {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	// Ambil dari config.yml (lihat patch config.go di bawah)
	accessMin := config.Env.JWT.AccessTTLMinutes
	if accessMin <= 0 {
		accessMin = 15
	}
	refreshDays := config.Env.JWT.RefreshTTLDays
	if refreshDays <= 0 {
		refreshDays = 30
	}

	return &Service{
		users:      users,
		roles:      roles,
		redis:      rdb,
		email:      sender,
		loc:        loc,
		pinTTL:     10 * time.Minute,
		accessTTL:  time.Duration(accessMin) * time.Minute,
		refreshTTL: time.Duration(refreshDays) * 24 * time.Hour,
		jwtSecret:  []byte(config.Env.JWT.Secret),
	}
}

/* ==================== Helpers ==================== */

func hasLetter(s string) bool {
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			return true
		}
	}
	return false
}
func hasDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}
func isNIK16(s string) bool {
	if len(s) != 16 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func randBytes(n int) []byte { b := make([]byte, n); _, _ = rand.Read(b); return b }
func sixDigitPIN() string {
	var b [3]byte
	_, _ = rand.Read(b[:])
	n := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	n = n % 1000000
	return fmt.Sprintf("%06d", n)
}

type jwtPair struct{ AccessToken, RefreshToken string }

func (s *Service) issueJWT(userID string) (jwtPair, string, error) {
	now := time.Now().In(s.loc)

	// access
	acc := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"typ": "access",
		"iat": now.Unix(),
		"exp": now.Add(s.accessTTL).Unix(),
	})
	at, err := acc.SignedString(s.jwtSecret)
	if err != nil {
		return jwtPair{}, "", err
	}

	// refresh (with jti)
	jti := base64.RawURLEncoding.EncodeToString(randBytes(16))
	ref := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"typ": "refresh",
		"jti": jti,
		"iat": now.Unix(),
		"exp": now.Add(s.refreshTTL).Unix(),
	})
	rt, err := ref.SignedString(s.jwtSecret)
	if err != nil {
		return jwtPair{}, "", err
	}

	// store refresh jti in Redis for rotation
	if err := s.redis.Set(context.Background(), "refresh:"+jti, userID, s.refreshTTL).Err(); err != nil {
		return jwtPair{}, "", err
	}
	return jwtPair{AccessToken: at, RefreshToken: rt}, jti, nil
}

func (s *Service) rotateRefresh(oldJTI string) {
	_ = s.redis.Del(context.Background(), "refresh:"+oldJTI).Err()
}

/* ==================== Core Flows ==================== */

func (s *Service) RegisterLite(ctx context.Context, req *request.RegisterLiteRequest) (*entity.User, error) {
	// validate
	if !hasLetter(req.Password) || !hasDigit(req.Password) || len(req.Password) < 8 {
		return nil, constant.ErrInvalidPassword
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	uname := strings.ToLower(strings.TrimSpace(req.Username))
	phone := strings.TrimSpace(req.Phone)
	if len(uname) < 3 {
		return nil, constant.ErrValidationFailed
	}

	// unique
	if u, _ := s.users.FindByEmail(email); u != nil {
		if u.Status == "pending" {
			// resend PIN
			pin := sixDigitPIN()
			key := "verify:pin:" + email
			if err := s.redis.Set(ctx, key, pin, s.pinTTL).Err(); err != nil {
				return nil, constant.ErrInternalServerError
			}
			if s.email != nil {
				html := buildEmailHTML("", pin, int(s.pinTTL.Minutes()))
				if err := s.email.Send(email, "PIN Verifikasi Akun MedikaOne", html); err != nil {
					_ = s.redis.Del(ctx, key).Err()
					return nil, constant.ErrEmailSendFailed
				}
			}
			return u, nil
		}
		return nil, constant.ErrEmailAlreadyActive
	}
	if ok, _ := s.users.ExistsUsername(uname); ok {
		return nil, constant.ErrDuplicateUsernameOrEmail
	}

	// create user
	salt := randBytes(16)
	key, err := scrypt.Key([]byte(req.Password), salt, 1<<15, 8, 1, 64)
	if err != nil {
		return nil, constant.ErrInternalServerError
	}
	hash := base64.StdEncoding.EncodeToString(key) + ":" + base64.StdEncoding.EncodeToString(salt)

	u := &entity.User{
		Email:        email,
		Username:     &uname,
		Phone:        &phone,
		PasswordHash: hash,
		Status:       "pending",
	}
	if err := s.users.Create(u); err != nil {
		return nil, constant.ErrInternalServerError
	}

	// PIN
	pin := sixDigitPIN()
	keyRedis := "verify:pin:" + email
	if err := s.redis.Set(ctx, keyRedis, pin, s.pinTTL).Err(); err != nil {
		return nil, constant.ErrInternalServerError
	}
	if s.email != nil {
		html := buildEmailHTML("", pin, int(s.pinTTL.Minutes()))
		if err := s.email.Send(email, "PIN Verifikasi Akun MedikaOne", html); err != nil {
			_ = s.redis.Del(ctx, keyRedis).Err()
			return nil, constant.ErrEmailSendFailed
		}
	}
	return u, nil
}

func (s *Service) ResendPIN(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.users.FindByEmail(email)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if u == nil {
		return constant.ErrUserNotFound
	}
	if u.Status == "active" {
		return constant.ErrEmailAlreadyActive
	}

	pin := sixDigitPIN()
	key := "verify:pin:" + email
	if err := s.redis.Set(ctx, key, pin, s.pinTTL).Err(); err != nil {
		return constant.ErrInternalServerError
	}
	if s.email != nil {
		html := buildEmailHTML(u.FirstName, pin, int(s.pinTTL.Minutes()))
		if err := s.email.Send(email, "PIN Verifikasi Akun MedikaOne", html); err != nil {
			_ = s.redis.Del(ctx, key).Err()
			return constant.ErrEmailSendFailed
		}
	}
	return nil
}

func (s *Service) VerifyPIN(ctx context.Context, email, pin string) (jwtPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.users.FindByEmail(email)
	if err != nil {
		return jwtPair{}, constant.ErrInternalServerError
	}
	if u == nil {
		return jwtPair{}, constant.ErrUserNotFound
	}
	val, err := s.redis.Get(ctx, "verify:pin:"+email).Result()
	if err != nil || val != pin {
		return jwtPair{}, constant.ErrInvalidOTP
	}

	// activate
	if err := s.users.MarkVerified(u.ID); err != nil {
		return jwtPair{}, constant.ErrInternalServerError
	}
	_ = s.redis.Del(ctx, "verify:pin:"+email).Err()

	// issue tokens
	pair, _, err := s.issueJWT(u.ID)
	if err != nil {
		return jwtPair{}, constant.ErrInternalServerError
	}
	return pair, nil
}

func (s *Service) Login(ctx context.Context, identity, password string) (jwtPair, error) {
	identity = strings.ToLower(strings.TrimSpace(identity))
	var u *entity.User
	var err error
	if strings.Contains(identity, "@") {
		u, err = s.users.FindByEmail(identity)
	} else {
		u, err = s.users.FindByUsername(identity)
	}
	if err != nil {
		return jwtPair{}, constant.ErrInternalServerError
	}
	if u == nil {
		return jwtPair{}, constant.ErrUserNotFound
	}
	if u.Status != "active" {
		return jwtPair{}, constant.ErrEmailNotVerified
	}

	// verify password
	parts := strings.Split(u.PasswordHash, ":")
	if len(parts) != 2 {
		return jwtPair{}, constant.ErrInternalServerError
	}
	keyEnc, saltEnc := parts[0], parts[1]
	salt, err := base64.StdEncoding.DecodeString(saltEnc)
	if err != nil {
		return jwtPair{}, constant.ErrInternalServerError
	}
	derived, err := scrypt.Key([]byte(password), salt, 1<<15, 8, 1, 64)
	if err != nil {
		return jwtPair{}, constant.ErrInternalServerError
	}
	if base64.StdEncoding.EncodeToString(derived) != keyEnc {
		return jwtPair{}, constant.ErrPasswordNotMatch
	}

	pair, _, err := s.issueJWT(u.ID)
	if err != nil {
		return jwtPair{}, constant.ErrInternalServerError
	}
	return pair, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (jwtPair, error) {
	tok, err := jwt.Parse(refreshToken, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("bad sign")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !tok.Valid {
		return jwtPair{}, constant.ErrInvalidToken
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || claims["typ"] != "refresh" {
		return jwtPair{}, constant.ErrInvalidToken
	}

	jti, _ := claims["jti"].(string)
	sub, _ := claims["sub"].(string)
	if jti == "" || sub == "" {
		return jwtPair{}, constant.ErrInvalidToken
	}

	// check jti in redis (rotation)
	val, err := s.redis.Get(ctx, "refresh:"+jti).Result()
	if err != nil || val != sub {
		return jwtPair{}, constant.ErrTokenExpired
	}

	// rotate
	s.rotateRefresh(jti)
	pair, _, err := s.issueJWT(sub)
	if err != nil {
		return jwtPair{}, constant.ErrInternalServerError
	}
	return pair, nil
}

/* ==================== Role & Profiles ==================== */

func (s *Service) ChooseRole(ctx context.Context, userID, roleSlug string) error {
	r, err := s.roles.FindBySlug(roleSlug)
	if err != nil || r == nil {
		return constant.ErrAccountRoleNotFound
	}
	if err := s.roles.Assign(userID, r.ID); err != nil {
		return constant.ErrInternalServerError
	}
	return nil
}

func (s *Service) CompletePatientProfile(ctx context.Context, userID string, req *request.PatientProfileRequest) error {
	updates := map[string]any{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"address":    req.Address,
		"gender":     req.Gender,
		"nik":        req.NIK,
	}
	if req.NIK != nil && !isNIK16(*req.NIK) {
		return constant.ErrValidationFailed
	}
	if req.DOB != nil && *req.DOB != "" {
		tm, err := time.ParseInLocation("2006-01-02", *req.DOB, s.loc)
		if err != nil || tm.After(time.Now().In(s.loc)) {
			return constant.ErrValidationFailed
		}
		updates["dob"] = tm
	}
	if err := s.users.UpdateByID(userID, updates); err != nil {
		return constant.ErrInternalServerError
	}

	prof := map[string]any{
		"user_id":      userID,
		"height_cm":    req.HeightCM,
		"weight_kg":    req.WeightKG,
		"allergies":    req.Allergies,
		"medical_hist": req.History,
	}
	if err := s.users.UpsertPatientProfile(prof); err != nil {
		return constant.ErrInternalServerError
	}
	return nil
}

func (s *Service) CompleteDoctorProfile(ctx context.Context, userID string, req *request.DoctorProfileRequest) error {
	updates := map[string]any{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"address":    req.Address,
		"gender":     req.Gender,
	}
	if err := s.users.UpdateByID(userID, updates); err != nil {
		return constant.ErrInternalServerError
	}

	prof := map[string]any{
		"user_id":    userID,
		"sip_number": req.SIP,
		"specialty":  req.Specialty,
	}
	if err := s.users.UpsertDoctorProfile(prof); err != nil {
		return constant.ErrInternalServerError
	}
	return nil
}

/* ==================== Email HTML ==================== */
func buildEmailHTML(firstName, pin string, ttlMin int) string {
	if strings.TrimSpace(firstName) == "" {
		firstName = "Pengguna"
	}
	return fmt.Sprintf(`
<!doctype html><html><body style="font-family:Arial,Helvetica,sans-serif;background:#f6f9fc;padding:24px">
  <div style="max-width:560px;margin:0 auto;background:#fff;border:1px solid #e6ecf1;border-radius:12px;padding:24px">
    <h2 style="margin:0 0 8px 0;color:#111">MedikaOne</h2>
    <p style="color:#555">Halo %s, gunakan PIN berikut untuk verifikasi email (berlaku %d menit):</p>
    <div style="text-align:center;margin:16px 0">
      <span style="display:inline-block;font-size:28px;letter-spacing:6px;font-weight:700;background:#0ea5e9;color:#fff;padding:12px 16px;border-radius:10px">%s</span>
    </div>
    <p style="color:#667">Jika kamu tidak meminta verifikasi ini, abaikan email ini.</p>
  </div>
  <div style="text-align:center;color:#99a; font-size:12px;margin-top:10px">© %d MedikaOne</div>
</body></html>`, firstName, ttlMin, pin, time.Now().Year())
}
