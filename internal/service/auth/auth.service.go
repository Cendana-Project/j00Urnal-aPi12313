package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/scrypt"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/contract/repository"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
)

type Service struct {
	userRepo    repository.UserRepository
	roleRepo    repository.RoleRepository
	redis       *redis.Client
	emailSender EmailSender
	pinTTL      time.Duration
	timeLoc     *time.Location
}

type EmailSender interface {
	Send(to, subject, htmlBody string) error
}

func NewService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	redisClient *redis.Client,
	emailSender EmailSender,
) *Service {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return &Service{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		redis:       redisClient,
		emailSender: emailSender,
		pinTTL:      120 * time.Minute, // TTL PIN verifikasi
		timeLoc:     loc,
	}
}

func (s *Service) Register(ctx context.Context, req *request.RegisterRequest) (*response.RegisterResponse, error) {
	// Validasi dasar
	if !hasLetter(req.Password) || !hasDigit(req.Password) || len(req.Password) < 8 {
		return nil, errors.New("password harus minimal 8 karakter dan kombinasi huruf+angka")
	}
	if req.NIK != nil && !isNIK16(*req.NIK) {
		return nil, errors.New("NIK harus 16 digit numerik")
	}
	if !hasLetter(req.Password) || !hasDigit(req.Password) || len(req.Password) < 8 {
		return nil, constant.ErrInvalidPassword
	}
	if req.NIK != nil && !isNIK16(*req.NIK) {
		return nil, constant.ErrValidationFailed
	}
	// Normalisasi email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	existing, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}

	// === CASE: user sudah ada ===
	if existing != nil {
		switch existing.Status {
		case "pending":
			// resend PIN & 200 OK (tidak return error)
			pin := sixDigitPIN()
			key := "verify:pin:" + existing.Email
			if err := s.redis.Set(ctx, key, pin, s.pinTTL).Err(); err != nil {
				return nil, constant.ErrInternalServerError
			}
			if s.emailSender != nil {
				html := buildVerificationEmailHTML(existing.FirstName, pin, s.timeLoc, int(s.pinTTL.Minutes()))
				if err := s.emailSender.Send(existing.Email, "PIN Verifikasi Akun MedikaOne", html); err != nil {
					_ = s.redis.Del(ctx, key).Err()
					return nil, constant.ErrEmailSendFailed
				}
			}
			return &response.RegisterResponse{
				UserID: existing.ID, Email: existing.Email, Status: "pending",
			}, nil

		case "active":
			return nil, constant.ErrEmailAlreadyActive

		default:
			return nil, constant.ErrInternalServerError
		}
	}

	// === CASE: email belum ada → buat akun baru ===
	salt := randBytes(16)
	key, err := scrypt.Key([]byte(req.Password), salt, 1<<15, 8, 1, 64)
	if err != nil {
		return nil, err
	}
	hash := base64.StdEncoding.EncodeToString(key) + ":" + base64.StdEncoding.EncodeToString(salt)

	var dobPtr *time.Time
	if req.DOB != nil && *req.DOB != "" {
		tm, err := time.ParseInLocation("2006-01-02", *req.DOB, s.timeLoc)
		if err != nil {
			return nil, errors.New("format DOB harus YYYY-MM-DD")
		}
		if tm.After(time.Now().In(s.timeLoc)) {
			return nil, errors.New("DOB tidak boleh di masa depan")
		}
		dobPtr = &tm
	}

	u := &entity.User{
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		DOB:          dobPtr,
		Address:      req.Address,
		Gender:       req.Gender, // L | P
		NIK:          req.NIK,    // 16 digit
		PasswordHash: hash,
		Status:       "pending",
	}
	if err := s.userRepo.Create(u); err != nil {
		return nil, err
	}

	role, err := s.roleRepo.FindBySlug(req.AccountRole)
	if err != nil || role == nil {
		return nil, constant.ErrAccountRoleNotFound
	}
	if err := s.roleRepo.Assign(u.ID, role.ID); err != nil {
		return nil, err
	}

	// Buat PIN + kirim email
	pin := sixDigitPIN()
	keyRedis := "verify:pin:" + u.Email
	if err := s.redis.Set(ctx, keyRedis, pin, s.pinTTL).Err(); err != nil {
		return nil, constant.ErrInternalServerError
	}
	if s.emailSender != nil {
		html := buildVerificationEmailHTML(u.FirstName, pin, s.timeLoc, int(s.pinTTL.Minutes()))
		if err := s.emailSender.Send(u.Email, "PIN Verifikasi Akun MedikaOne", html); err != nil {
			_ = s.redis.Del(ctx, keyRedis).Err()
			return nil, constant.ErrEmailSendFailed
		}
	}

	return &response.RegisterResponse{
		UserID: u.ID, Email: u.Email, Status: "pending",
	}, nil
}

// VerifyPIN memvalidasi PIN 6 digit yang dikirim via email, lalu aktivasi akun.
func (s *Service) VerifyPIN(ctx context.Context, email, pin string) (*response.VerifyEmailResponse, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || len(pin) != 6 {
		return nil, constant.ErrValidationFailed
	}
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, constant.ErrInternalServerError
	}
	if u == nil {
		return nil, constant.ErrUserNotFound
	}
	if u.Status == "active" {
		return &response.VerifyEmailResponse{Email: email, Status: "active"}, nil
	}
	val, err := s.redis.Get(ctx, "verify:pin:"+email).Result()
	if err != nil {
		return nil, constant.ErrInvalidOTP
	} // pakai ini untuk “invalid or expired OTP”
	if val != pin {
		return nil, constant.ErrInvalidOTP
	}
	if err := s.userRepo.MarkVerified(u.ID); err != nil {
		return nil, constant.ErrInternalServerError
	}
	_ = s.redis.Del(ctx, "verify:pin:"+email).Err()
	return &response.VerifyEmailResponse{Email: email, Status: "active"}, nil
}

// ===== helpers =====

func sixDigitPIN() string {
	// aman & uniform: ambil 3 byte random → 0..16777215 → mod 1e6
	var b [3]byte
	_, _ = rand.Read(b[:])
	n := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	n = n % 1000000
	return fmt.Sprintf("%06d", n)
}

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

// Email HTML sederhana & rapi (inline CSS, aman di banyak klien)
var emailTpl = template.Must(template.New("verify").Parse(`
<!doctype html>
<html>
<head>
<meta charset="utf-8"/>
<title>PIN Verifikasi MedikaOne</title>
</head>
<body style="margin:0;padding:0;background:#f6f9fc;font-family:Arial,Helvetica,sans-serif;color:#223;">
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#f6f9fc;padding:20px 0;">
    <tr><td align="center">
      <table role="presentation" width="560" cellspacing="0" cellpadding="0" style="background:#fff;border-radius:12px;padding:24px;border:1px solid #e6ecf1;">
        <tr>
          <td style="text-align:center;">
            <div style="font-size:20px;font-weight:700;margin-bottom:8px;">MedikaOne</div>
            <div style="font-size:14px;color:#5a6b7b;">Verifikasi Email Akun</div>
          </td>
        </tr>
        <tr><td style="height:16px"></td></tr>
        <tr>
          <td style="font-size:14px;line-height:22px;">
            <p>Halo {{.Name}},</p>
            <p>Gunakan PIN di bawah ini untuk memverifikasi akun kamu. PIN berlaku selama <b>{{.TTL}} menit</b>.</p>
          </td>
        </tr>
        <tr><td style="height:12px"></td></tr>
        <tr>
          <td align="center">
            <div style="display:inline-block;font-size:28px;letter-spacing:6px;font-weight:700;background:#0ea5e9;color:#fff;padding:12px 16px;border-radius:10px;">
              {{.PIN}}
            </div>
          </td>
        </tr>
        <tr><td style="height:12px"></td></tr>
        <tr>
          <td style="font-size:12px;color:#5a6b7b;line-height:20px;">
            <p>Jika kamu tidak meminta verifikasi ini, abaikan email ini.</p>
            <p>Terima kasih,<br/>Tim MedikaOne</p>
          </td>
        </tr>
      </table>
      <div style="font-size:11px;color:#8a99a8;margin-top:12px;">© {{.Year}} MedikaOne</div>
    </td></tr>
  </table>
</body>
</html>
`))

func buildVerificationEmailHTML(firstName, pin string, loc *time.Location, ttlMin int) string {
	if strings.TrimSpace(firstName) == "" {
		firstName = "Pengguna"
	}
	var sb strings.Builder
	_ = emailTpl.Execute(&sb, map[string]any{
		"Name": firstName,
		"PIN":  pin,
		"TTL":  ttlMin,
		"Year": time.Now().In(loc).Year(),
	})
	return sb.String()
}
