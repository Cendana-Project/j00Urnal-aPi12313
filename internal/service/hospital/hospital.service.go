package hospital

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
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

// ====== Create Hospital (super_admin only) ======
type CreateHospitalReq struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (s *Service) CreateHospital(ctx context.Context, req CreateHospitalReq) (string, error) {
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return "", constant.ErrValidationFailed
	}
	h := &hrepo.Hospital{Name: req.Name}
	if req.Code != "" {
		h.Code = &req.Code
	}
	if err := s.hospitalRepo.Create(ctx, h); err != nil {
		return "", err
	}
	if req.Code != "" {
		found, err := s.hospitalRepo.FindByCode(ctx, req.Code)
		if err != nil {
			return "", err
		}
		return found.ID, nil
	}
	// NOTE: untuk skenario tanpa code, sebaiknya pakai RETURNING id di repo Create.
	return "", nil
}

// ====== Create Admin Hospital (by super_admin) ======
type CreateHospitalAdminReq struct {
	HospitalID string `json:"hospital_id"` // UUID atau code (akan di-resolve)
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Password   string `json:"password"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
}

func (s *Service) CreateHospitalAdmin(ctx context.Context, req CreateHospitalAdminReq) (string, error) {
	hospitalID, err := s.hospitalRepo.ResolveHospitalID(ctx, req.HospitalID)
	if err != nil {
		return "", constant.ErrRecordNotFound
	}
	if req.Email == "" || req.Password == "" || req.FirstName == "" {
		return "", constant.ErrValidationFailed
	}

	// upsert user (aktif tanpa verifikasi)
	var uid string
	u, err := s.userRepo.FindByEmail(req.Email) // <- tanpa context
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", constant.ErrInternalServerError
	}
	if u == nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		uid = uuid.NewString()
		if err := s.userRepo.InsertActive(ctx, urepo.InsertUser{
			ID:           uid,
			Email:        req.Email,
			Phone:        req.Phone,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			PasswordHash: string(hash),
			VerifiedAt:   time.Now(),
		}); err != nil {
			return "", err
		}
	} else {
		uid = u.ID
	}

	// membership
	if err := s.hospitalRepo.EnsureMembership(ctx, uid, hospitalID, true); err != nil {
		return "", err
	}

	// assign role admin (tenant scoped)
	roleID, err := s.roleRepo.GetRoleIDBySlug(ctx, constant.RoleAdmin)
	if err != nil {
		return "", err
	}
	if err := s.hospitalRepo.AssignHospitalRole(ctx, uid, hospitalID, roleID); err != nil {
		return "", err
	}
	return uid, nil
}

// ====== Create Staff (by tenant admin) ======
type CreateHospitalStaffReq struct {
	HospitalHint string `json:"-"` // diisi dari context / path
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Password     string `json:"password"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	RoleSlug     string `json:"role"` // doctor|nurse|receptionist|bod|admin
}

func (s *Service) CreateHospitalStaff(ctx context.Context, req CreateHospitalStaffReq) (string, error) {
	hospitalID, err := s.hospitalRepo.ResolveHospitalID(ctx, req.HospitalHint)
	if err != nil {
		return "", constant.ErrRecordNotFound
	}
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.RoleSlug == "" {
		return "", constant.ErrValidationFailed
	}

	// find or create user (aktif)
	var uid string
	u, err := s.userRepo.FindByEmail(req.Email) // <- tanpa context
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", constant.ErrInternalServerError
	}
	if u == nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		uid = uuid.NewString()
		if err := s.userRepo.InsertActive(ctx, urepo.InsertUser{
			ID:           uid,
			Email:        req.Email,
			Phone:        req.Phone,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			PasswordHash: string(hash),
			VerifiedAt:   time.Now(),
		}); err != nil {
			return "", err
		}
	} else {
		uid = u.ID
	}

	// membership
	if err := s.hospitalRepo.EnsureMembership(ctx, uid, hospitalID, false); err != nil {
		return "", err
	}

	// assign tenant role
	roleID, err := s.roleRepo.GetRoleIDBySlug(ctx, req.RoleSlug)
	if err != nil {
		return "", err
	}
	if err := s.hospitalRepo.AssignHospitalRole(ctx, uid, hospitalID, roleID); err != nil {
		return "", err
	}
	return uid, nil
}
