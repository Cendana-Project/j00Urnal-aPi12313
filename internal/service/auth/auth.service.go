package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/email"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	hosprepo "github.com/api-monolith-template/internal/repository/hospital"
	rolerepo "github.com/api-monolith-template/internal/repository/role"
	userrepo "github.com/api-monolith-template/internal/repository/user"
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

type LoginHospitalResult struct {
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	ExpiresIn    int64                `json:"expires_in"`
	TokenType    string               `json:"token_type"`
	HospitalID   string               `json:"hospital_id"`
	Roles        []response.RoleBrief `json:"roles"`
	AccessExp    time.Time            `json:"-"`
	RefreshExp   time.Time            `json:"-"`
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
	acc, ref := config.Env.JWT.ParseDurations()

	return &Service{
		users:      users,
		roles:      roles,
		redis:      rdb,
		email:      sender,
		hosp:       hosp,
		loc:        loc,
		pinTTL:     10 * time.Minute,
		accessTTL:  acc,
		refreshTTL: ref,
		jwtSecret:  []byte(config.Env.JWT.Secret),
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

func (s *Service) issueJWT(userID string) (TokenPair, string, time.Time, time.Time, error) {
	now := time.Now().In(s.loc)
	accessExp := now.Add(s.accessTTL)
	refreshExp := now.Add(s.refreshTTL)

	acc := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID, "typ": "access", "iat": now.Unix(), "exp": accessExp.Unix(),
	})
	at, err := acc.SignedString(s.jwtSecret)
	if err != nil {
		return TokenPair{}, "", time.Time{}, time.Time{}, err
	}

	jti := base64.RawURLEncoding.EncodeToString(randBytes(16))
	ref := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID, "typ": "refresh", "jti": jti, "iat": now.Unix(), "exp": refreshExp.Unix(),
	})
	rt, err := ref.SignedString(s.jwtSecret)
	if err != nil {
		return TokenPair{}, "", time.Time{}, time.Time{}, err
	}

	if err := s.redis.Set(context.Background(), "refresh:"+jti, userID, s.refreshTTL).Err(); err != nil {
		return TokenPair{}, "", time.Time{}, time.Time{}, err
	}
	return TokenPair{AccessToken: at, RefreshToken: rt}, jti, accessExp.UTC(), refreshExp.UTC(), nil
}

func (s *Service) rotateRefresh(oldJTI string) {
	_ = s.redis.Del(context.Background(), "refresh:"+oldJTI).Err()
}

/* ==================== Core Flows ==================== */

func (s *Service) RegisterLite(ctx context.Context, req *request.RegisterLiteRequest) (*entity.User, error) {
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
			pin := sixDigitPIN()
			key := "verify:pin:" + emailAddr
			if err := s.redis.Set(ctx, key, pin, s.pinTTL).Err(); err != nil {
				return nil, constant.ErrInternalServerError
			}
			if s.email != nil {
				html := email.RenderVerifyPIN("", pin, int(s.pinTTL.Minutes()))
				if err := s.email.Send(emailAddr, "PIN Verifikasi Akun MedikaOne", html); err != nil {
					_ = s.redis.Del(ctx, key).Err()
					return nil, constant.ErrEmailSendFailed
				}
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

	u := &entity.User{
		Email:        emailAddr,
		Username:     &uname,
		Phone:        &phone,
		PasswordHash: hash,
		Status:       "pending",
	}
	if err := s.users.Create(u); err != nil {
		return nil, constant.ErrInternalServerError
	}

	pin := sixDigitPIN()
	keyRedis := "verify:pin:" + emailAddr
	if err := s.redis.Set(ctx, keyRedis, pin, s.pinTTL).Err(); err != nil {
		return nil, constant.ErrInternalServerError
	}
	if s.email != nil {
		html := email.RenderVerifyPIN("", pin, int(s.pinTTL.Minutes()))
		if err := s.email.Send(emailAddr, "PIN Verifikasi Akun MedikaOne", html); err != nil {
			_ = s.redis.Del(ctx, keyRedis).Err()
			return nil, constant.ErrEmailSendFailed
		}
	}
	ulog.Infof(ctx, "register success email=%s", emailAddr)
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
		html := email.RenderVerifyPIN(u.FirstName, pin, int(s.pinTTL.Minutes()))
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

	pair, _, aexp, rexp, err := s.issueJWT(u.ID)
	if err != nil {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
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
		return pair, nil, time.Time{}, time.Time{}, constant.ErrUserNotFound
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

func (s *Service) LoginHospital(ctx context.Context, identifier, password, hospitalHint string) (*LoginHospitalResult, error) {
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

	// password verify & migrate jika perlu
	if err := s.verifyAndMigratePassword(ctx, u, password); err != nil {
		return nil, err
	}

	ok, lerr := s.hosp.IsUserLinkedToHospital(ctx, u.ID, hID)
	if lerr != nil {
		return nil, constant.ErrInternalServerError
	}
	if !ok {
		return nil, constant.ErrUserNotLinkedToHospital
	}

	j, _, aexp, rexp, jerr := s.issueJWT(u.ID)
	if jerr != nil {
		return nil, constant.ErrInternalServerError
	}

	rs, rerr := s.roles.ListHospitalRolesByUser(ctx, hID, u.ID)
	if rerr != nil {
		return nil, constant.ErrInternalServerError
	}
	rbrief := make([]response.RoleBrief, 0, len(rs))
	for _, r := range rs {
		rbrief = append(rbrief, response.RoleBrief{ID: r.ID, Slug: r.Slug, Name: r.Name})
	}

	ulog.Infof(ctx, "login hospital success user_id=%s hospital_id=%s", u.ID, hID)
	return &LoginHospitalResult{
		AccessToken: j.AccessToken, RefreshToken: j.RefreshToken,
		ExpiresIn: int64(s.accessTTL / time.Second), TokenType: "Bearer",
		HospitalID: hID, Roles: rbrief,
		AccessExp: aexp, RefreshExp: rexp,
	}, nil
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

	val, err := s.redis.Get(ctx, "refresh:"+jti).Result()
	if err != nil || val != sub {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrTokenExpired
	}

	s.rotateRefresh(jti)
	pair, _, aexp, rexp, err := s.issueJWT(sub)
	if err != nil {
		return TokenPair{}, time.Time{}, time.Time{}, constant.ErrInternalServerError
	}
	ulog.Infof(ctx, "refresh success user_id=%s", sub)
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
	ulog.Infof(ctx, "choose role user_id=%s role=%s", userID, roleSlug)
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
		"medical_hist": req.MedicalHistory,
	}
	if err := s.users.UpsertPatientProfile(prof); err != nil {
		return constant.ErrInternalServerError
	}
	ulog.Infof(ctx, "patient profile updated user_id=%s", userID)
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
		"sip_number": req.SIPNumber,
		"specialty":  req.Specialty,
	}
	if err := s.users.UpsertDoctorProfile(prof); err != nil {
		return constant.ErrInternalServerError
	}
	ulog.Infof(ctx, "doctor profile updated user_id=%s", userID)
	return nil
}

func (s *Service) SetProfile(ctx context.Context, userID, roleSlugUpper string, rawProfile *json.RawMessage) (*response.SetProfileResponse, error) {
	if rawProfile == nil {
		return nil, constant.ErrValidationError
	}
	role := strings.ToUpper(strings.TrimSpace(roleSlugUpper))
	switch role {
	case constant.RolePatient, constant.RoleDoctor:
	default:
		return nil, constant.ErrValidationError
	}

	if role == constant.RolePatient {
		ok, err := s.users.ExistsPatientProfile(userID)
		if err != nil {
			return nil, constant.ErrInternalServerError
		}
		if ok {
			return nil, constant.ErrProfileAlreadySet
		}
	} else {
		ok, err := s.users.ExistsDoctorProfile(userID)
		if err != nil {
			return nil, constant.ErrInternalServerError
		}
		if ok {
			return nil, constant.ErrProfileAlreadySet
		}
	}

	r, err := s.roles.FindBySlug(role)
	if err != nil || r == nil || !r.Active {
		return nil, constant.ErrAccountRoleNotFound
	}
	if err := s.roles.Assign(userID, r.ID); err != nil {
		return nil, constant.ErrInternalServerError
	}

	if role == constant.RolePatient {
		var req request.PatientProfileRequest
		if err := json.Unmarshal(*rawProfile, &req); err != nil {
			return nil, constant.ErrValidationError
		}
		if err := s.CompletePatientProfile(ctx, userID, &req); err != nil {
			return nil, err
		}
	} else {
		var req request.DoctorProfileRequest
		if err := json.Unmarshal(*rawProfile, &req); err != nil {
			return nil, constant.ErrValidationError
		}
		if err := s.CompleteDoctorProfile(ctx, userID, &req); err != nil {
			return nil, err
		}
	}

	u, err := s.users.GetByID(userID)
	if err != nil || u == nil {
		return nil, constant.ErrInternalServerError
	}
	var dobStr *string
	if u.DOB != nil {
		tmp := u.DOB.In(s.loc).Format("2006-01-02")
		dobStr = &tmp
	}
	prof := response.UserProfile{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		DOB:       dobStr,
		Address:   u.Address,
		Gender:    u.Gender,
		NIK:       u.NIK,
	}
	if role == constant.RolePatient {
		h, w, a, m, err := s.users.GetPatientProfileByUserID(userID)
		if err != nil {
			return nil, constant.ErrInternalServerError
		}
		prof.HeightCM, prof.WeightKG, prof.Allergies, prof.MedicalHistory = h, w, a, m
	} else {
		sip, spc, err := s.users.GetDoctorProfileByUserID(userID)
		if err != nil {
			return nil, constant.ErrInternalServerError
		}
		prof.SIPNumber, prof.Specialty = sip, spc
	}

	ulog.Infof(ctx, "set-profile success user_id=%s role=%s", userID, role)
	return &response.SetProfileResponse{
		Role:    role,
		Profile: prof,
	}, nil
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
		html := email.RenderResetPIN(u.FirstName, pin, int(s.pinTTL.Minutes()))
		if err := s.email.Send(emailAddr, "PIN Reset Password MedikaOne", html); err != nil {
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
