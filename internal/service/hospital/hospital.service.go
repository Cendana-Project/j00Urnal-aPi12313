package hospital

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/scrypt"
	"gorm.io/datatypes"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	hrepo "github.com/api-monolith-template/internal/repository/hospital"
	rrepo "github.com/api-monolith-template/internal/repository/role"
	urepo "github.com/api-monolith-template/internal/repository/user"
)

type Service struct {
	userRepo     *urepo.Repository
	roleRepo     *rrepo.Repository
	hospitalRepo *hrepo.Repository
	cache        *redis.Client
}

func NewService(u *urepo.Repository, r *rrepo.Repository, h *hrepo.Repository, cache *redis.Client) *Service {
	return &Service{userRepo: u, roleRepo: r, hospitalRepo: h, cache: cache}
}

/* ================= Helpers ================= */

func sp(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	v := s
	return &v
}

func parseDOB(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	tm, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, err
	}
	return &tm, nil
}

// scrypt hashing: menghasilkan "key:salt" (base64)
func hashScrypt(plain string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	dk, err := scrypt.Key([]byte(plain), salt, 1<<15, 8, 1, 64)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(dk) + ":" + base64.StdEncoding.EncodeToString(salt), nil
}

/* =============== Hospital =============== */

func (s *Service) CreateHospital(ctx context.Context, in *request.CreateHospitalRequest) (*entity.Hospital, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	name := strings.TrimSpace(in.Name)
	address := strings.TrimSpace(in.Address)
	city := strings.TrimSpace(in.City)
	province := strings.TrimSpace(in.Province)
	phone := strings.TrimSpace(in.Phone)
	country := strings.TrimSpace(in.Country)
	if country == "" {
		country = "Indonesia"
	}

	// Wajib
	if code == "" || name == "" || address == "" || city == "" || province == "" || phone == "" {
		return nil, constant.ErrValidationFailed
	}
	// Range lat/lon jika diisi
	if in.Latitude != nil && (*in.Latitude < -90 || *in.Latitude > 90) {
		return nil, constant.ErrValidationFailed
	}
	if in.Longitude != nil && (*in.Longitude < -180 || *in.Longitude > 180) {
		return nil, constant.ErrValidationFailed
	}
	// Uniqueness
	if ok, err := s.hospitalRepo.IsCodeExists(ctx, code); err != nil {
		return nil, constant.ErrInternalServerError
	} else if ok {
		return nil, constant.ErrConflict
	}
	if ok, err := s.hospitalRepo.IsNameExists(ctx, name); err != nil {
		return nil, constant.ErrInternalServerError
	} else if ok {
		return nil, constant.ErrConflict
	}

	// Facilities -> JSONB
	var facJSON datatypes.JSON
	if in.Facilities != nil {
		b, err := json.Marshal(in.Facilities)
		if err != nil {
			return nil, constant.ErrValidationFailed
		}
		facJSON = datatypes.JSON(b)
	}

	now := time.Now().UTC()
	h := &entity.Hospital{
		// id: default gen_random_uuid() dari DB
		Code:        sp(code),
		Name:        name,
		Address:     sp(address),
		City:        sp(city),
		Province:    sp(province),
		Country:     sp(country),
		Latitude:    in.Latitude,
		Longitude:   in.Longitude,
		Phone:       sp(phone),
		Description: sp(strings.TrimSpace(in.Description)),
		Facilities:  facJSON,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.hospitalRepo.Create(ctx, h); err != nil {
		return nil, constant.ErrInternalServerError
	}
	return h, nil
}

/* =============== Users (Admin/Staff) =============== */

// CreateHospitalAdmin: hanya super_admin, user baru Wajib Unik (email & username)
func (s *Service) CreateHospitalAdmin(ctx context.Context, req request.CreateHospitalAdminRequest) (string, error) {
	hospitalID, err := s.hospitalRepo.ResolveHospitalID(ctx, req.HospitalID)
	if err != nil {
		return "", constant.ErrRecordNotFound
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Username) == "" {
		return "", constant.ErrValidationFailed
	}

	dob, err := parseDOB(req.DOB)
	if err != nil {
		return "", constant.ErrValidationFailed
	}

	uid, err := s.ensureUserActiveFull(ctx, urepo.InsertUserFull{
		Email:         strings.ToLower(strings.TrimSpace(req.Email)),
		Username:      sp(req.Username), // request bertipe string → konversi ke *string
		Phone:         req.Phone,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		DOB:           dob,
		Address:       req.Address,
		Gender:        req.Gender, // validator menjaga L|P
		NIK:           req.NIK,    // validator menjaga 16 digit
		PasswordPlain: req.Password,
	})
	if err != nil {
		return "", err
	}

	// membership & role admin (tenant)
	if err := s.hospitalRepo.EnsureMembership(ctx, uid, hospitalID, true); err != nil {
		return "", constant.ErrInternalServerError
	}
	roleID, err := s.roleRepo.GetRoleIDBySlug(ctx, constant.RoleAdmin)
	if err != nil {
		return "", constant.ErrInternalServerError
	}
	if err := s.hospitalRepo.AssignHospitalRole(ctx, uid, hospitalID, roleID); err != nil {
		return "", constant.ErrInternalServerError
	}
	return uid, nil
}

// CreateHospitalStaff: admin tenant, user baru Wajib Unik (email & username)
func (s *Service) CreateHospitalStaff(ctx context.Context, req request.CreateHospitalStaffRequest) (string, error) {
	// 1) Resolve hospital
	hospitalID, err := s.hospitalRepo.ResolveHospitalID(ctx, req.HospitalID)
	if err != nil || strings.TrimSpace(hospitalID) == "" {
		return "", constant.ErrHospitalNotFound // <=== pakai custom error khusus hospital
	}

	// 2) Validasi field wajib
	if strings.TrimSpace(req.Email) == "" ||
		strings.TrimSpace(req.Password) == "" ||
		strings.TrimSpace(req.Role) == "" ||
		strings.TrimSpace(req.Username) == "" {
		return "", constant.ErrValidationFailed
	}

	// 3) Validasi/parse DOB
	dob, err := parseDOB(req.DOB)
	if err != nil {
		return "", constant.ErrInvalidDateFormat
	}

	// 4) Buat/aktifkan user (idempotent sesuai implementasi ensureUserActiveFull kamu)
	uid, err := s.ensureUserActiveFull(ctx, urepo.InsertUserFull{
		Email:         strings.ToLower(strings.TrimSpace(req.Email)),
		Username:      sp(req.Username),
		Phone:         req.Phone,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		DOB:           dob,
		Address:       req.Address,
		Gender:        req.Gender,
		NIK:           req.NIK,
		PasswordPlain: req.Password,
	})
	if err != nil {
		// Mapping umum untuk duplikat & password
		// Sesuaikan bagian ini dengan error yang dikembalikan ensureUserActiveFull (mis. string contains/ sentinel)
		if errors.Is(err, gorm.ErrDuplicatedKey) ||
			strings.Contains(strings.ToLower(err.Error()), "duplicate key") ||
			strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
			return "", constant.ErrDuplicateUsernameOrEmail
		}
		if strings.Contains(strings.ToLower(err.Error()), "password") &&
			(strings.Contains(strings.ToLower(err.Error()), "invalid") ||
				strings.Contains(strings.ToLower(err.Error()), "require")) {
			return "", constant.ErrInvalidPassword
		}
		if strings.Contains(strings.ToLower(err.Error()), "similar to username") ||
			strings.Contains(strings.ToLower(err.Error()), "similar to email") {
			return "", constant.ErrPasswordSimilarToUserInfo
		}
		return "", constant.ErrInternalServerError
	}

	// 5) Pastikan membership user->hospital
	if err := s.hospitalRepo.EnsureMembership(ctx, uid, hospitalID, false); err != nil {
		return "", constant.ErrInternalServerError
	}

	// 6) Validasi role (gunakan constants & izinkan DOCTOR juga)
	roleSlug := strings.ToUpper(strings.TrimSpace(req.Role))
	switch roleSlug {
	case constant.RoleNurse, constant.RoleReceptionist, constant.RoleBOD, constant.RoleAdmin:
	// OK
	case constant.RoleDoctor:
		return "", constant.ErrRegistrationError
	default:
		return "", constant.ErrRoleNotFound
	}

	// 7) Ambil role id (repo sudah case-insensitive)
	roleID, err := s.roleRepo.GetRoleIDBySlug(ctx, roleSlug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", constant.ErrRoleNotFound
		}
		return "", constant.ErrInternalServerError
	}

	// 8) Assign role ke user pada hospital
	if err := s.hospitalRepo.AssignHospitalRole(ctx, uid, hospitalID, roleID); err != nil {
		// Tangkap conflict unique (sudah pernah di-assign)
		if errors.Is(err, gorm.ErrDuplicatedKey) ||
			strings.Contains(strings.ToLower(err.Error()), "duplicate key") ||
			strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
			return "", constant.ErrRoleAlreadyAssigned
		}
		return "", constant.ErrInternalServerError
	}

	return uid, nil
}

/* =============== ensureUserActiveFull =============== */

// Aturan baru: HARUS user baru (tidak “find-or-create”).
// Jika email/username sudah ada → ErrDuplicateUsernameOrEmail (409).
func (s *Service) ensureUserActiveFull(ctx context.Context, in urepo.InsertUserFull) (string, error) {
	// email unik
	if u, err := s.userRepo.FindByEmail(in.Email); err != nil {
		return "", constant.ErrInternalServerError
	} else if u != nil {
		return "", constant.ErrDuplicateUsernameOrEmail
	}

	// username unik (jika ada)
	if in.Username != nil && *in.Username != "" {
		exists, err := s.userRepo.ExistsUsername(*in.Username)
		if err != nil {
			return "", constant.ErrInternalServerError
		}
		if exists {
			return "", constant.ErrDuplicateUsernameOrEmail
		}
	}

	// NIK unik (jika ada)
	if in.NIK != nil && *in.NIK != "" {
		exists, err := s.userRepo.ExistsNIK(*in.NIK)
		if err != nil {
			return "", constant.ErrInternalServerError
		}
		if exists {
			return "", constant.ErrConflict
		}
	}

	// Hash scrypt
	h, err := hashScrypt(in.PasswordPlain)
	if err != nil {
		return "", constant.ErrInternalServerError
	}
	in.PasswordHash = h

	// defaults
	in.ID = uuid.NewString()
	in.Status = "active"
	now := time.Now().UTC()
	in.VerifiedAt = &now

	if err := s.userRepo.InsertActiveFull(ctx, in); err != nil {
		return "", constant.ErrInternalServerError
	}
	return in.ID, nil
}
