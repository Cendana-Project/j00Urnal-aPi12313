package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv" // <=== added
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/email"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	rolerepo "github.com/api-monolith-template/internal/repository/role"
	userrepo "github.com/api-monolith-template/internal/repository/user"
	"github.com/api-monolith-template/internal/service/manuscript"
	ulog "github.com/api-monolith-template/internal/util"
)

/* ==================== Types ==================== */

type EmailSender interface {
	Send(to, subject, htmlBody string) error
}

// TokenPair diexpose supaya metode exported tidak mengembalikan tipe unexported.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
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

	manuscriptService *manuscript.Service
}

func NewService(users *userrepo.Repository, roles *rolerepo.Repository, rdb *redis.Client, sender EmailSender, ms *manuscript.Service) *Service {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	acc := config.Env.Token.AccessTokenDuration
	ref := config.Env.Token.RefreshTokenDuration

	return &Service{
		users:             users,
		roles:             roles,
		redis:             rdb,
		email:             sender,
		loc:               loc,
		pinTTL:            10 * time.Minute,
		accessTTL:         acc,
		refreshTTL:        ref,
		jwtSecret:         []byte(config.Env.Token.AccessTokenSecret),
		manuscriptService: ms,
	}
}

/* ==================== Helpers ==================== */

var (
	errUserNotFound = constant.ErrUserNotFound
	errValidation   = constant.ErrValidationFailed
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
	return strconv6(n)
}
func strconv6(n int) string {
	d := [6]byte{'0', '0', '0', '0', '0', '0'}
	for i := 5; i >= 0; i-- {
		d[i] = byte('0' + (n % 10))
		n /= 10
	}
	return string(d[:])
}

// ====== Key helpers (Redis) ======
func keyRefresh(jti string) string         { return "refresh:" + jti }          // existing pattern
func keyUserRefreshSet(uid string) string  { return "user:refreshes:" + uid }   // <=== added
func keyAccessBlacklist(jti string) string { return "access:blacklist:" + jti } // <=== added

// ====== JWT helpers ======
func parseJWTHS256(tokenStr string, secret []byte) (jwt.MapClaims, error) { // <=== added
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
	if err != nil || !tok.Valid {
		return nil, constant.ErrInvalidToken
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, constant.ErrInvalidToken
	}
	return claims, nil
}

// Konversi aman ke int64 dari berbagai tipe (float64, string, json.Number, int, dll). // <=== added
func toInt64(v any) int64 { // <=== added
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	case json.Number:
		if n, err := t.Int64(); err == nil {
			return n
		}
	case string:
		if t == "" {
			return 0
		}
		if n, err := strconv.ParseInt(t, 10, 64); err == nil {
			return n
		}
	}
	return 0
}

/* ==================== JWT Issue & Rotate ==================== */

func (s *Service) issueJWT(userID string) (TokenPair, string, time.Time, time.Time, error) {
	now := time.Now().In(s.loc)
	accessExp := now.Add(s.accessTTL)
	refreshExp := now.Add(s.refreshTTL)

	// --- Access token: tambahkan jti ---
	accessJTI := base64.RawURLEncoding.EncodeToString(randBytes(16)) // <=== added
	acc := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID, "typ": "access", "jti": accessJTI, "iat": now.Unix(), "exp": accessExp.Unix(), // <=== changed
	})
	at, err := acc.SignedString(s.jwtSecret)
	if err != nil {
		return TokenPair{}, "", time.Time{}, time.Time{}, err
	}

	// --- Refresh token: sudah pakai jti sebelumnya ---
	refreshJTI := base64.RawURLEncoding.EncodeToString(randBytes(16))
	ref := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID, "typ": "refresh", "jti": refreshJTI, "iat": now.Unix(), "exp": refreshExp.Unix(),
	})
	rt, err := ref.SignedString(s.jwtSecret)
	if err != nil {
		return TokenPair{}, "", time.Time{}, time.Time{}, err
	}

	// Simpan refresh:<jti> -> userID (EX=RefreshTTL)
	if err := s.redis.Set(context.Background(), keyRefresh(refreshJTI), userID, s.refreshTTL).Err(); err != nil {
		return TokenPair{}, "", time.Time{}, time.Time{}, err
	}
	// Index semua refresh jti milik user (untuk logout-all)
	_ = s.redis.SAdd(context.Background(), keyUserRefreshSet(userID), refreshJTI).Err() // <=== added

	return TokenPair{AccessToken: at, RefreshToken: rt}, refreshJTI, accessExp.UTC(), refreshExp.UTC(), nil
}

func (s *Service) rotateRefresh(oldJTI string) {
	_ = s.redis.Del(context.Background(), keyRefresh(oldJTI)).Err()
	// Tidak tahu userID di sini; SREM dilakukan saat Refresh ketika kita tahu sub. // <=== added (note)
}

// assignDefaultAuthorRole grants the global AUTHOR role (idempotent via ON CONFLICT DO NOTHING).
func (s *Service) assignDefaultAuthorRole(ctx context.Context, userID string) error {
	if userID == "" {
		return constant.ErrInternalServerError
	}
	role, err := s.roles.FindBySlug(constant.RoleAuthor)
	if err != nil {
		return err
	}
	if role == nil || role.ID == "" {
		ulog.Errorf(ctx, "role AUTHOR not found in database")
		return constant.ErrInternalServerError
	}
	if err := s.roles.Assign(userID, role.ID); err != nil {
		return err
	}
	return nil
}

/* ==================== Core Flows ==================== */

func (s *Service) Register(ctx context.Context, req *request.RegisterRequest) (*entity.User, error) {
	ulog.Infof(ctx, "register attempt email=%s", req.Email)

	if !hasLetter(req.Password) || !hasDigit(req.Password) || len(req.Password) < 8 {
		return nil, constant.ErrInvalidPassword
	}
	emailAddr := strings.ToLower(strings.TrimSpace(req.Email))
	uname := strings.ToLower(strings.TrimSpace(req.Username))
	phone := strings.TrimSpace(req.Phone)
	if len(uname) < 3 {
		return nil, constant.ErrValidationFailed
	}

	if u, _ := s.users.FindByEmail(emailAddr); u != nil {
		if u.Status == "pending" {
			// If email service is disabled, auto-activate existing pending user
			if s.email == nil {
				if err := s.users.UpdateByID(u.ID, map[string]any{"status": "active"}); err != nil {
					return nil, constant.ErrInternalServerError
				}
				u.Status = "active"
				ulog.Infof(ctx, "Email service disabled - auto-activated existing pending user: %s", emailAddr)
				if err := s.assignDefaultAuthorRole(ctx, u.ID); err != nil {
					return nil, err
				}
				return u, nil
			}

			// Send verification PIN if email service is available
			pin := sixDigitPIN()
			key := "verify:pin:" + emailAddr
			if err := s.redis.Set(ctx, key, pin, s.pinTTL).Err(); err != nil {
				return nil, constant.ErrInternalServerError
			}
			html := email.RenderVerifyPIN("", pin, int(s.pinTTL.Minutes()))
			if err := s.email.Send(emailAddr, "PIN Verifikasi Akun MedikaOne", html); err != nil {
				ulog.Errorf(ctx, "smtp send failed (register_resend_pin): %v", err)
				_ = s.redis.Del(ctx, key).Err()
				return nil, constant.ErrEmailSendFailed
			}
			ulog.Infof(ctx, "register resend pin email=%s", emailAddr)
			return u, nil
		}
		return nil, constant.ErrEmailAlreadyActive
	}
	if ok, _ := s.users.ExistsUsername(uname); ok {
		return nil, constant.ErrDuplicateUsernameOrEmail
	}

	salt := randBytes(16)
	key, err := scrypt.Key([]byte(req.Password), salt, 1<<15, 8, 1, 64)
	if err != nil {
		return nil, constant.ErrInternalServerError
	}
	hash := base64.StdEncoding.EncodeToString(key) + ":" + base64.StdEncoding.EncodeToString(salt)

	// Determine initial status based on email availability
	// If email service is disabled, auto-activate for development/testing
	initialStatus := "pending"
	if s.email == nil {
		initialStatus = "active" // Auto-activate if no email service
		ulog.Infof(ctx, "Email service disabled - auto-activating user")
	}

	u := &entity.User{
		Email:        emailAddr,
		Username:     uname, // Fixed: now string
		Phone:        &phone,
		PasswordHash: hash,
		Status:       initialStatus,
		Affiliation:  &req.Affiliation,
	}
	if err := s.users.Create(u); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) ||
			strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
			// Try to distinguish if it's email or username if possible, or just generic duplicate
			if strings.Contains(strings.ToLower(err.Error()), "username") {
				return nil, constant.ErrDuplicateUsernameOrEmail
			}
			return nil, constant.ErrEmailAlreadyActive
		}
		return nil, constant.ErrInternalServerError
	}

	if err := s.assignDefaultAuthorRole(ctx, u.ID); err != nil {
		return nil, err
	}

	// Only send verification email if email service is available
	if s.email != nil {
		pin := sixDigitPIN()
		keyRedis := "verify:pin:" + emailAddr
		if err := s.redis.Set(ctx, keyRedis, pin, s.pinTTL).Err(); err != nil {
			return nil, constant.ErrInternalServerError
		}
		html := email.RenderVerifyPIN("", pin, int(s.pinTTL.Minutes()))
		if err := s.email.Send(emailAddr, "PIN Verifikasi Akun MedikaOne", html); err != nil {
			ulog.Errorf(ctx, "smtp send failed (register_send_pin): %v", err)
			_ = s.redis.Del(ctx, keyRedis).Err()
			return nil, constant.ErrEmailSendFailed
		}
		ulog.Infof(ctx, "register success - verification email sent to %s", emailAddr)
	} else {
		ulog.Infof(ctx, "register success - user auto-activated (no email service) email=%s", emailAddr)
	}
	return u, nil
}

func (s *Service) ResendPIN(ctx context.Context, emailAddr string) error {
	emailAddr = strings.ToLower(strings.TrimSpace(emailAddr))
	u, err := s.users.FindByEmail(emailAddr)
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
	key := "verify:pin:" + emailAddr
	if err := s.redis.Set(ctx, key, pin, s.pinTTL).Err(); err != nil {
		return constant.ErrInternalServerError
	}
	if s.email != nil {
		name := ""
		if u.FirstName != nil {
			name = *u.FirstName
		}
		html := email.RenderVerifyPIN(name, pin, int(s.pinTTL.Minutes()))
		if err := s.email.Send(emailAddr, "PIN Verifikasi Akun MedikaOne", html); err != nil {
			_ = s.redis.Del(ctx, key).Err()
			return constant.ErrEmailSendFailed
		}
	}
	ulog.Infof(ctx, "resend pin email=%s", emailAddr)
	return nil
}

func (s *Service) VerifyPIN(ctx context.Context, emailAddr, pin string) (TokenPair, time.Time, time.Time, error) {
	emailAddr = strings.ToLower(strings.TrimSpace(emailAddr))
	u, err := s.users.FindByEmail(emailAddr)
	if err != nil {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	if u == nil {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrUserNotFound
	}

	val, err := s.redis.Get(ctx, "verify:pin:"+emailAddr).Result()
	if err != nil || val != pin {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrInvalidOTP
	}

	if err := s.users.MarkVerified(u.ID); err != nil {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	_ = s.redis.Del(ctx, "verify:pin:"+emailAddr).Err()

	if err := s.assignDefaultAuthorRole(ctx, u.ID); err != nil {
		return TokenPair{}, time.Time{}, time.Time{}, err
	}

	pair, _, aexp, rexp, err := s.issueJWT(u.ID)
	if err != nil {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}

	// Update LastLogin on verification success? Or maybe on Login only.
	// For now, let's leave it as is.

	ulog.Infof(ctx, "verify pin success email=%s", emailAddr)
	return pair, aexp, rexp, nil
}

func (s *Service) Login(ctx context.Context, identity, password string) (pair struct {
	AccessToken  string
	RefreshToken string
}, roles []response.RoleBrief, accessExp, refreshExp time.Time, err error) {
	identity = strings.ToLower(strings.TrimSpace(identity))
	ulog.Infof(ctx, "login attempt identity=%s", identity)

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
		ulog.Errorf(ctx, "login fail: user not found")
		return pair, nil, time.Time{}, time.Time{}, constant.ErrPasswordNotMatch
	}
	if strings.ToLower(u.Status) != "active" {
		ulog.Errorf(ctx, "login fail: email not verified")
		return pair, nil, time.Time{}, time.Time{}, constant.ErrEmailNotVerified
	}

	// verify password via hybrid (bcrypt → migrate ke scrypt jika perlu)
	if err := s.verifyAndMigratePassword(ctx, u, password); err != nil {
		ulog.Errorf(ctx, "login fail: password not match")
		return pair, nil, time.Time{}, time.Time{}, err
	}

	j, _, aexp, rexp, jerr := s.issueJWT(u.ID)
	if jerr != nil {
		return pair, nil, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	pair.AccessToken, pair.RefreshToken = j.AccessToken, j.RefreshToken
	accessExp, refreshExp = aexp, rexp

	rows, rerr := s.roles.ListRolesByUser(ctx, u.ID)
	if rerr != nil {
		return pair, nil, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	out := make([]response.RoleBrief, 0, len(rows))
	for _, r := range rows {
		out = append(out, response.RoleBrief{ID: r.ID, Slug: r.Slug, Name: r.Name})
	}

	ulog.Infof(ctx, "login success user_id=%s", u.ID)
	return pair, out, accessExp, refreshExp, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, time.Time, time.Time, error) {
	tok, err := jwt.Parse(refreshToken, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("bad sign")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !tok.Valid {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrInvalidToken
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || claims["typ"] != "refresh" {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrInvalidToken
	}

	jti, _ := claims["jti"].(string)
	sub, _ := claims["sub"].(string)
	if jti == "" || sub == "" {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrInvalidToken
	}

	val, err := s.redis.Get(ctx, keyRefresh(jti)).Result()
	if err != nil || val != sub {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrTokenExpired
	}

	// Hapus refresh lama & indeks user
	s.rotateRefresh(jti)
	_ = s.redis.SRem(ctx, keyUserRefreshSet(sub), jti).Err() // <=== added

	pair, _, aexp, rexp, err := s.issueJWT(sub)
	if err != nil {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	ulog.Infof(ctx, "refresh success user_id=%s", sub)
	return pair, aexp, rexp, nil
}

/* ==================== Role ==================== */

func (s *Service) ChooseRole(ctx context.Context, userID, roleSlug string) error {
	roleID, err := s.roles.GetRoleIDBySlug(ctx, roleSlug)
	if err != nil {
		return err
	}
	return s.roles.Assign(userID, roleID)
}

/* ==================== Password helpers ==================== */

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
		if newHash, err := s.hashPasswordScrypt(plain); err == nil && newHash != "" {
			_ = s.users.UpdateByID(u.ID, map[string]any{"password_hash": newHash})
			u.PasswordHash = newHash
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

/* ==================== Forgot/Reset/Change Password ==================== */

func (s *Service) PasswordForgot(ctx context.Context, req *request.PasswordForgotRequest) error {
	emailAddr := strings.ToLower(strings.TrimSpace(req.Email))

	u, err := s.users.FindByEmail(emailAddr)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if u == nil {
		return constant.ErrUserNotFound
	}
	if u.Status != "active" {
		return constant.ErrEmailNotVerified
	}

	pin := sixDigitPIN()
	key := "pwd:pin:" + emailAddr
	if err := s.redis.Set(ctx, key, pin, s.pinTTL).Err(); err != nil {
		return constant.ErrInternalServerError
	}
	if s.email != nil {
		name := ""
		if u.FirstName != nil {
			name = *u.FirstName
		}
		html := email.RenderResetPIN(name, pin, int(s.pinTTL.Minutes()))
		if err := s.email.Send(emailAddr, "PIN Reset Password MedikaOne", html); err != nil {
			ulog.Errorf(ctx, "smtp send failed (forgot_send_pin): %v", err)
			_ = s.redis.Del(ctx, key).Err()
			return constant.ErrEmailSendFailed
		}
	}
	ulog.Infof(ctx, "password forgot pin sent email=%s", emailAddr)
	return nil
}

func (s *Service) PasswordReset(ctx context.Context, req *request.PasswordResetRequest) error {
	emailAddr := strings.ToLower(strings.TrimSpace(req.Email))
	newPass := strings.TrimSpace(req.NewPassword)
	pin := strings.TrimSpace(req.PIN)

	if !hasLetter(newPass) || !hasDigit(newPass) || len(newPass) < 8 {
		return constant.ErrInvalidPassword
	}

	u, err := s.users.FindByEmail(emailAddr)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if u == nil {
		return constant.ErrUserNotFound
	}
	if u.Status != "active" {
		return constant.ErrEmailNotVerified
	}

	val, err := s.redis.Get(ctx, "pwd:pin:"+emailAddr).Result()
	if err != nil || val != pin {
		return constant.ErrInvalidOTP
	}

	newHash, err := s.hashPasswordScrypt(newPass)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if err := s.users.UpdateByID(u.ID, map[string]any{"password_hash": newHash}); err != nil {
		return constant.ErrInternalServerError
	}

	_ = s.redis.Del(ctx, "pwd:pin:"+emailAddr).Err()
	ulog.Infof(ctx, "password reset success user_id=%s", u.ID)
	return nil
}

func (s *Service) PasswordChange(ctx context.Context, userID string, req *request.PasswordChangeRequest) error {
	oldPass := strings.TrimSpace(req.OldPassword)
	newPass := strings.TrimSpace(req.NewPassword)

	if !hasLetter(newPass) || !hasDigit(newPass) || len(newPass) < 8 {
		return constant.ErrInvalidPassword
	}
	if strings.EqualFold(oldPass, newPass) {
		return constant.ErrNewPasswordSame
	}

	u, err := s.users.GetByID(userID)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if u == nil {
		return constant.ErrUserNotFound
	}

	if err := s.verifyAndMigratePassword(ctx, u, oldPass); err != nil {
		return err
	}

	newHash, err := s.hashPasswordScrypt(newPass)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if err := s.users.UpdateByID(userID, map[string]any{"password_hash": newHash}); err != nil {
		return constant.ErrInternalServerError
	}

	ulog.Infof(ctx, "password change success user_id=%s", userID)
	return nil
}

/* ==================== Logout APIs ==================== */

func (s *Service) Logout(ctx context.Context, accessToken string, refreshToken string) error { // <=== unchanged signature
	if strings.TrimSpace(accessToken) == "" {
		return constant.ErrUnauthorized
	}
	claims, err := parseJWTHS256(accessToken, s.jwtSecret)
	if err != nil {
		return err
	}
	if typ, _ := claims["typ"].(string); typ != "access" {
		return constant.ErrInvalidToken
	}
	acJTI, _ := claims["jti"].(string)
	sub, _ := claims["sub"].(string)
	// expUnix, _ := ulog.GetInt64(claims["exp"])                              // <=== removed
	expUnix := toInt64(claims["exp"]) // <=== changed
	if acJTI == "" || sub == "" || expUnix == 0 {
		return constant.ErrInvalidToken
	}
	ttl := time.Until(time.Unix(expUnix, 0))
	if ttl <= 0 {
		ttl = time.Second
	}
	// Blacklist access token
	if err := s.redis.Set(ctx, keyAccessBlacklist(acJTI), "1", ttl).Err(); err != nil {
		return constant.ErrInternalServerError
	}

	// Revoke refresh token jika dikirim
	if strings.TrimSpace(refreshToken) != "" {
		rfClaims, err := parseJWTHS256(refreshToken, s.jwtSecret)
		if err == nil && rfClaims != nil {
			if rfTyp, _ := rfClaims["typ"].(string); rfTyp == "refresh" {
				rfJTI, _ := rfClaims["jti"].(string)
				rfSub, _ := rfClaims["sub"].(string)
				if rfJTI != "" {
					_ = s.redis.Del(ctx, keyRefresh(rfJTI)).Err()
					if rfSub != "" {
						_ = s.redis.SRem(ctx, keyUserRefreshSet(rfSub), rfJTI).Err()
					} else {
						_ = s.redis.SRem(ctx, keyUserRefreshSet(sub), rfJTI).Err()
					}
				}
			}
		}
	}

	ulog.Infof(ctx, "logout success user_id=%s", sub)
	return nil
}

func (s *Service) LogoutAll(ctx context.Context, accessToken string) error { // <=== unchanged signature
	if strings.TrimSpace(accessToken) == "" {
		return constant.ErrUnauthorized
	}
	claims, err := parseJWTHS256(accessToken, s.jwtSecret)
	if err != nil {
		return err
	}
	if typ, _ := claims["typ"].(string); typ != "access" {
		return constant.ErrInvalidToken
	}
	acJTI, _ := claims["jti"].(string)
	sub, _ := claims["sub"].(string)
	// expUnix, _ := ulog.GetInt64(claims["exp"])                              // <=== removed
	expUnix := toInt64(claims["exp"]) // <=== changed
	if acJTI == "" || sub == "" || expUnix == 0 {
		return constant.ErrInvalidToken
	}
	ttl := time.Until(time.Unix(expUnix, 0))
	if ttl <= 0 {
		ttl = time.Second
	}
	// Blacklist access token sekarang
	if err := s.redis.Set(ctx, keyAccessBlacklist(acJTI), "1", ttl).Err(); err != nil {
		return constant.ErrInternalServerError
	}
	// Revoke semua refresh milik user
	setKey := keyUserRefreshSet(sub)
	members, err := s.redis.SMembers(ctx, setKey).Result()
	if err == nil {
		for _, jti := range members {
			_ = s.redis.Del(ctx, keyRefresh(jti)).Err()
		}
		_ = s.redis.Del(ctx, setKey).Err()
	}

	ulog.Infof(ctx, "logout-all success user_id=%s", sub)
	return nil
}

func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	// 1. Check if user exists
	u, err := s.users.GetByID(userID)
	if err != nil {
		return err
	}
	if u == nil {
		return constant.ErrUserNotFound
	}

	// 2. Cascade Delete Manuscripts (where user is Main Author)
	if err := s.manuscriptService.DeleteByAuthor(ctx, userID); err != nil {
		return err
	}

	// 3. Delete User Record
	if err := s.users.Delete(userID); err != nil {
		return err
	}

	// 4. Cleanup Redis (Auth tokens, etc)
	_ = s.redis.Del(ctx, keyUserRefreshSet(userID)).Err()
	// We might want to blacklist tokens, but simpler to just delete refresh tokens.

	ulog.Infof(ctx, "delete user success user_id=%s", userID)
	return nil
}
