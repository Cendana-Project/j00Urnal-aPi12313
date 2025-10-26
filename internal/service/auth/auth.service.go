package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/api-monolith-template/internal/model/response"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	hosprepo "github.com/api-monolith-template/internal/repository/hospital"
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
	hosp  *hosprepo.Repository
	email EmailSender

	loc        *time.Location
	pinTTL     time.Duration
	accessTTL  time.Duration
	refreshTTL time.Duration
	jwtSecret  []byte
}

func NewService(users *userrepo.Repository, roles *rolerepo.Repository, rdb *redis.Client, sender EmailSender, hosp *hosprepo.Repository) *Service {
	loc, _ := time.LoadLocation("Asia/Jakarta")

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
		hosp:       hosp,
		loc:        loc,
		pinTTL:     10 * time.Minute,
		accessTTL:  time.Duration(accessMin) * time.Minute,
		refreshTTL: time.Duration(refreshDays) * 24 * time.Hour,
		jwtSecret:  []byte(config.Env.JWT.Secret),
	}
}

/* ==================== Helpers ==================== */

var (
	errUserNotFound = constant.ErrUserNotFound     // 404
	errValidation   = constant.ErrValidationFailed // 400
)

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

func (s *Service) issueJWT(userID string) (jwtPair, string, time.Time, time.Time, error) { // <=== changed
	now := time.Now().In(s.loc)
	accessExp := now.Add(s.accessTTL)
	refreshExp := now.Add(s.refreshTTL)

	// access
	acc := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID, "typ": "access", "iat": now.Unix(), "exp": accessExp.Unix(),
	})
	at, err := acc.SignedString(s.jwtSecret)
	if err != nil {
		return jwtPair{}, "", time.Time{}, time.Time{}, err
	}

	// refresh (with jti)
	jti := base64.RawURLEncoding.EncodeToString(randBytes(16))
	ref := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID, "typ": "refresh", "jti": jti, "iat": now.Unix(), "exp": refreshExp.Unix(),
	})
	rt, err := ref.SignedString(s.jwtSecret)
	if err != nil {
		return jwtPair{}, "", time.Time{}, time.Time{}, err
	}

	// store refresh jti in Redis for rotation
	if err := s.redis.Set(context.Background(), "refresh:"+jti, userID, s.refreshTTL).Err(); err != nil {
		return jwtPair{}, "", time.Time{}, time.Time{}, err
	}
	return jwtPair{AccessToken: at, RefreshToken: rt}, jti, accessExp.UTC(), refreshExp.UTC(), nil
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

func (s *Service) VerifyPIN(ctx context.Context, email, pin string) (jwtPair, time.Time, time.Time, error) { // <=== changed
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.users.FindByEmail(email)
	if err != nil {
		return jwtPair{}, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	if u == nil {
		return jwtPair{}, time.Time{}, time.Time{}, constant.ErrUserNotFound
	}

	val, err := s.redis.Get(ctx, "verify:pin:"+email).Result()
	if err != nil || val != pin {
		return jwtPair{}, time.Time{}, time.Time{}, constant.ErrInvalidOTP
	}

	// activate
	if err := s.users.MarkVerified(u.ID); err != nil {
		return jwtPair{}, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	_ = s.redis.Del(ctx, "verify:pin:"+email).Err()

	// issue tokens + expiry
	pair, _, aexp, rexp, err := s.issueJWT(u.ID)
	if err != nil {
		return jwtPair{}, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	return pair, aexp, rexp, nil
}

// Login: verifikasi credential, issue token, + kembalikan roles global
func (s *Service) Login(ctx context.Context, identity, password string) (pair struct {
	AccessToken  string
	RefreshToken string
}, roles []response.RoleBrief, accessExp, refreshExp time.Time, err error) { // <=== changed
	identity = strings.ToLower(strings.TrimSpace(identity))

	var u *entity.User
	if strings.Contains(identity, "@") {
		u, err = s.users.FindByEmail(identity)
	} else {
		u, err = s.users.FindByUsername(identity)
	}
	if err != nil {
		return pair, nil, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	if u == nil {
		return pair, nil, time.Time{}, time.Time{}, constant.ErrUserNotFound
	}
	if strings.ToLower(u.Status) != "active" {
		return pair, nil, time.Time{}, time.Time{}, constant.ErrEmailNotVerified
	}

	// verify password (scrypt "key:salt")
	parts := strings.Split(u.PasswordHash, ":")
	if len(parts) != 2 {
		return pair, nil, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	keyEnc, saltEnc := parts[0], parts[1]
	salt, derr := base64.StdEncoding.DecodeString(saltEnc)
	if derr != nil {
		return pair, nil, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	derived, derr := scrypt.Key([]byte(password), salt, 1<<15, 8, 1, 64)
	if derr != nil {
		return pair, nil, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	if base64.StdEncoding.EncodeToString(derived) != keyEnc {
		return pair, nil, time.Time{}, time.Time{}, constant.ErrPasswordNotMatch
	}

	// tokens + expiry
	j, _, aexp, rexp, jerr := s.issueJWT(u.ID)
	if jerr != nil {
		return pair, nil, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	pair.AccessToken, pair.RefreshToken = j.AccessToken, j.RefreshToken
	accessExp, refreshExp = aexp, rexp

	// roles global
	rows, rerr := s.roles.ListRolesByUser(ctx, u.ID)
	if rerr != nil {
		return pair, nil, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	out := make([]response.RoleBrief, 0, len(rows))
	for _, r := range rows {
		out = append(out, response.RoleBrief{ID: r.ID, Slug: r.Slug, Name: r.Name})
	}

	return pair, out, accessExp, refreshExp, nil
}

// LoginHospital: untuk sekarang login standar + TODO cek hospital & membership.
// Param hospitalID/hospitalCode opsional; salah satu harus terisi (sudah diverifikasi di controller).
func (s *Service) LoginHospital(ctx context.Context, identifier, password, hospitalHint string) (*LoginHospitalResult, error) {
	// resolve hospital, user, membership, verify password (kode aslimu) …
	hID, err := s.hosp.ResolveHospitalID(ctx, hospitalHint)
	if err != nil || hID == "" {
		return nil, constant.ErrHospitalNotFound
	}

	id := strings.ToLower(strings.TrimSpace(identifier))
	var u *entity.User
	if strings.Contains(id, "@") {
		u, err = s.users.FindByEmail(id)
	} else {
		u, err = s.users.FindByUsername(id)
	}
	if err != nil {
		return nil, constant.ErrInternalServerError
	}
	if u == nil {
		return nil, constant.ErrUserNotFound
	}
	if strings.ToLower(u.Status) != "active" {
		return nil, constant.ErrForbidden
	}

	ok, lerr := s.hosp.IsUserLinkedToHospital(ctx, u.ID, hID)
	if lerr != nil {
		return nil, constant.ErrInternalServerError
	}
	if !ok {
		return nil, constant.ErrUserNotLinkedToHospital
	}

	// issue tokens + expiry
	j, _, aexp, rexp, jerr := s.issueJWT(u.ID)
	if jerr != nil {
		return nil, constant.ErrInternalServerError
	}

	// roles tenant
	rs, rerr := s.roles.ListHospitalRolesByUser(ctx, hID, u.ID)
	if rerr != nil {
		return nil, constant.ErrInternalServerError
	}
	rbrief := make([]response.RoleBrief, 0, len(rs))
	for _, r := range rs {
		rbrief = append(rbrief, response.RoleBrief{ID: r.ID, Slug: r.Slug, Name: r.Name})
	}

	return &LoginHospitalResult{
		AccessToken: j.AccessToken, RefreshToken: j.RefreshToken,
		ExpiresIn: int64(s.accessTTL / time.Second), TokenType: "Bearer",
		HospitalID: hID, Roles: rbrief,
		AccessExp: aexp, RefreshExp: rexp, // <=== added
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (jwtPair, time.Time, time.Time, error) { // <=== changed
	tok, err := jwt.Parse(refreshToken, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("bad sign")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !tok.Valid {
		return jwtPair{}, time.Time{}, time.Time{}, constant.ErrInvalidToken
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || claims["typ"] != "refresh" {
		return jwtPair{}, time.Time{}, time.Time{}, constant.ErrInvalidToken
	}

	jti, _ := claims["jti"].(string)
	sub, _ := claims["sub"].(string)
	if jti == "" || sub == "" {
		return jwtPair{}, time.Time{}, time.Time{}, constant.ErrInvalidToken
	}

	// check jti in redis (rotation)
	val, err := s.redis.Get(ctx, "refresh:"+jti).Result()
	if err != nil || val != sub {
		return jwtPair{}, time.Time{}, time.Time{}, constant.ErrTokenExpired
	}

	// rotate & issue baru + expiry
	s.rotateRefresh(jti)
	pair, _, aexp, rexp, err := s.issueJWT(sub)
	if err != nil {
		return jwtPair{}, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	return pair, aexp, rexp, nil
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

// ===== Helpers: identity -> user, and password verification =====
func (s *Service) findUserByIdentity(ctx context.Context, identity string) (*entity.User, error) {
	id := strings.TrimSpace(identity)
	if id == "" {
		return nil, errValidation
	}
	// coba email dulu
	if u, err := s.users.FindByEmail(strings.ToLower(id)); err != nil {
		return nil, constant.ErrInternalServerError
	} else if u != nil {
		return u, nil
	}
	// lalu username
	if u, err := s.users.FindByUsername(strings.ToLower(id)); err != nil {
		return nil, constant.ErrInternalServerError
	} else if u != nil {
		return u, nil
	}
	return nil, errUserNotFound
}

func (s *Service) verifyPassword(stored string, password string) error {
	s = s // silence staticcheck if needed

	// --- BCRYPT ---
	if strings.HasPrefix(stored, "$2a$") ||
		strings.HasPrefix(stored, "$2b$") ||
		strings.HasPrefix(stored, "$2y$") {
		if err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)); err != nil {
			return constant.ErrPasswordNotMatch
		}
		return nil
	}

	// --- SCRYPT ---
	parts := strings.Split(stored, ":")
	if len(parts) != 2 {
		return constant.ErrInternalServerError
	}
	keyEnc, saltEnc := parts[0], parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltEnc)
	if err != nil {
		return constant.ErrInternalServerError
	}
	derived, err := scrypt.Key([]byte(password), salt, 1<<15, 8, 1, 64)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if base64.StdEncoding.EncodeToString(derived) != keyEnc {
		return constant.ErrPasswordNotMatch
	}
	return nil
}

type LoginHospitalResult struct {
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	ExpiresIn    int64                `json:"expires_in"`
	TokenType    string               `json:"token_type"`
	HospitalID   string               `json:"hospital_id"`
	Roles        []response.RoleBrief `json:"roles"`
	AccessExp    time.Time            `json:"-"` // <=== added (internal use for controller)
	RefreshExp   time.Time            `json:"-"` // <=== added
}

func (s *Service) hashPasswordScrypt(password string) (string, error) {
	salt := randBytes(16)
	key, err := scrypt.Key([]byte(password), salt, 1<<15, 8, 1, 64)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key) + ":" + base64.StdEncoding.EncodeToString(salt), nil
}

// verifyAndMigratePassword:
// - Jika stored bcrypt, verifikasi bcrypt -> kalau OK, rehash ke scrypt & update DB.
// - Jika stored scrypt, verifikasi scrypt seperti biasa.
func (s *Service) verifyAndMigratePassword(ctx context.Context, u *entity.User, plain string) error {
	stored := u.PasswordHash

	// BCRYPT
	if strings.HasPrefix(stored, "$2a$") ||
		strings.HasPrefix(stored, "$2b$") ||
		strings.HasPrefix(stored, "$2y$") {
		if err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)); err != nil {
			return constant.ErrPasswordNotMatch
		}
		// Rehash -> scrypt & update DB (best-effort)
		if newHash, err := s.hashPasswordScrypt(plain); err == nil && newHash != "" {
			_ = s.users.UpdateByID(u.ID, map[string]any{"password_hash": newHash})
			u.PasswordHash = newHash // keep in-memory consistent
		}
		return nil
	}

	// SCRYPT
	parts := strings.Split(stored, ":")
	if len(parts) != 2 {
		return constant.ErrInternalServerError
	}
	keyEnc, saltEnc := parts[0], parts[1]
	salt, err := base64.StdEncoding.DecodeString(saltEnc)
	if err != nil {
		return constant.ErrInternalServerError
	}
	derived, err := scrypt.Key([]byte(plain), salt, 1<<15, 8, 1, 64)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if base64.StdEncoding.EncodeToString(derived) != keyEnc {
		return constant.ErrPasswordNotMatch
	}
	return nil
}

// ===== Email template khusus reset password
func buildResetEmailHTML(firstName, pin string, ttlMin int) string {
	if strings.TrimSpace(firstName) == "" {
		firstName = "Pengguna"
	}
	return fmt.Sprintf(`
<!doctype html><html><body style="font-family:Arial,Helvetica,sans-serif;background:#f6f9fc;padding:24px">
  <div style="max-width:560px;margin:0 auto;background:#fff;border:1px solid #e6ecf1;border-radius:12px;padding:24px">
    <h2 style="margin:0 0 8px 0;color:#111">MedikaOne</h2>
    <p style="color:#555">Halo %s, berikut PIN reset password (berlaku %d menit):</p>
    <div style="text-align:center;margin:16px 0">
      <span style="display:inline-block;font-size:28px;letter-spacing:6px;font-weight:700;background:#0ea5e9;color:#fff;padding:12px 16px;border-radius:10px">%s</span>
    </div>
    <p style="color:#667">Jika kamu tidak meminta reset ini, abaikan email ini.</p>
  </div>
  <div style="text-align:center;color:#99a; font-size:12px;margin-top:10px">© %d MedikaOne</div>
</body></html>`, firstName, ttlMin, pin, time.Now().Year())
}

// ====== 2.1 Lupa Password — kirim PIN via email ======
func (s *Service) PasswordForgot(ctx context.Context, req *request.PasswordForgotRequest) error {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	u, err := s.users.FindByEmail(email)
	if err != nil {
		return constant.ErrInternalServerError
	}
	// Demi keamanan, jangan bocorkan ada/tidaknya user → tapi kita tetap kirim 200
	// Namun jika ingin strict:
	if u == nil {
		return constant.ErrUserNotFound
	}
	if u.Status != "active" {
		// Bisa juga return email-not-verified; untuk UX lebih aman kita "ok" saja
		return constant.ErrEmailNotVerified
		//return nil
	}

	pin := sixDigitPIN()
	key := "pwd:pin:" + email
	if err := s.redis.Set(ctx, key, pin, s.pinTTL).Err(); err != nil {
		return constant.ErrInternalServerError
	}
	if s.email != nil {
		html := buildResetEmailHTML(u.FirstName, pin, int(s.pinTTL.Minutes()))
		if err := s.email.Send(email, "PIN Reset Password MedikaOne", html); err != nil {
			_ = s.redis.Del(ctx, key).Err()
			return constant.ErrEmailSendFailed
		}
	}
	return nil
}

// ====== 2.2 Reset Password — verifikasi PIN & set password baru (scrypt) ======
func (s *Service) PasswordReset(ctx context.Context, req *request.PasswordResetRequest) error {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	newPass := strings.TrimSpace(req.NewPassword)
	pin := strings.TrimSpace(req.PIN)

	// validation policy
	if !hasLetter(newPass) || !hasDigit(newPass) || len(newPass) < 8 {
		return constant.ErrInvalidPassword
	}

	u, err := s.users.FindByEmail(email)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if u == nil {
		return constant.ErrUserNotFound
	}
	if u.Status != "active" {
		return constant.ErrEmailNotVerified
	}

	val, err := s.redis.Get(ctx, "pwd:pin:"+email).Result()
	if err != nil || val != pin {
		return constant.ErrInvalidOTP
	}

	// hash scrypt
	newHash, err := s.hashPasswordScrypt(newPass)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if err := s.users.UpdateByID(u.ID, map[string]any{"password_hash": newHash}); err != nil {
		return constant.ErrInternalServerError
	}

	// hapus PIN
	_ = s.redis.Del(ctx, "pwd:pin:"+email).Err()

	// (opsional) revoke refresh token: jika kamu simpan daftar jti per user, hapus di sini
	return nil
}

// ====== 2.3 Ubah Password (authenticated) ======
func (s *Service) PasswordChange(ctx context.Context, userID string, req *request.PasswordChangeRequest) error {
	oldPass := strings.TrimSpace(req.OldPassword)
	newPass := strings.TrimSpace(req.NewPassword)

	if !hasLetter(newPass) || !hasDigit(newPass) || len(newPass) < 8 {
		return constant.ErrInvalidPassword
	}
	if strings.EqualFold(oldPass, newPass) {
		return constant.ErrNewPasswordSame
	}

	// ambil user
	u, err := s.users.GetByID(userID)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if u == nil {
		return constant.ErrUserNotFound
	}

	// verifikasi old password (hybrid + migrate jika bcrypt)
	if err := s.verifyAndMigratePassword(ctx, u, oldPass); err != nil {
		return err // ini sudah typed (ErrPasswordNotMatch / Internal)
	}

	// set new scrypt
	newHash, err := s.hashPasswordScrypt(newPass)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if err := s.users.UpdateByID(userID, map[string]any{"password_hash": newHash}); err != nil {
		return constant.ErrInternalServerError
	}

	// (opsional) revoke refresh token di sini
	return nil
}
