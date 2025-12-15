package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) FindByEmail(email string) (*entity.User, error) {
	var u entity.User
	err := r.db.Where("LOWER(email)=LOWER(?)", strings.ToLower(email)).
		First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *Repository) FindByUsername(uname string) (*entity.User, error) {
	var u entity.User
	err := r.db.Where("LOWER(username)=LOWER(?)", strings.ToLower(uname)).
		First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *Repository) ExistsUsername(uname string) (bool, error) {
	var cnt int64
	if err := r.db.Model(&entity.User{}).
		Where("LOWER(username)=LOWER(?)", strings.ToLower(uname)).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (r *Repository) Create(u *entity.User) error { return r.db.Create(u).Error }

func (r *Repository) MarkVerified(id string) error {
	now := time.Now()
	return r.db.Model(&entity.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":      "active",
			"verified_at": now,
			"updated_at":  now,
		}).Error
}

func (r *Repository) UpdateByID(id string, fields map[string]any) error {
	fields["updated_at"] = time.Now()
	return r.db.Model(&entity.User{}).Where("id = ?", id).Updates(fields).Error
}

func (r *Repository) GetByID(id string) (*entity.User, error) {
	var u entity.User
	if err := r.db.First(&u, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetByEmail(email string) (*entity.User, error) {
	var u entity.User
	err := r.db.First(&u, "email = ? AND deleted_at IS NULL", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *Repository) Update(u *entity.User) error { return r.db.Save(u).Error }

type InsertUser struct {
	ID           string
	Email        string
	Username     string
	Phone        string
	FirstName    *string
	LastName     *string
	PasswordHash string
	VerifiedAt   time.Time
}

func (r *Repository) InsertActive(ctx context.Context, in InsertUser) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO users (id, email, username, phone, first_name, last_name, password_hash, status, verified_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'active', ?, NOW())
	`, in.ID, in.Email, in.Username, in.Phone, in.FirstName, in.LastName, in.PasswordHash, in.VerifiedAt).Error
}
func (r *Repository) ExistsNIK(nik string) (bool, error) {
	if strings.TrimSpace(nik) == "" {
		return false, nil
	}
	var cnt int64
	if err := r.db.Model(&entity.User{}).
		Where("nik = ? AND deleted_at IS NULL", nik).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

type InsertUserFull struct {
	ID            string
	Email         string
	Username      string // Fixed: now string
	Phone         *string
	FirstName     *string
	LastName      *string
	DOB           *time.Time
	Address       *string
	Gender        *string // "L" | "P"
	NIK           *string
	PasswordPlain string // input (akan di-hash di service)
	PasswordHash  string // hasil scrypt "hash:salt"
	Status        string // "active"
	VerifiedAt    *time.Time
}

// InsertActiveFull: membuat user aktif dengan field lengkap (tanpa verifikasi)
func (r *Repository) InsertActiveFull(ctx context.Context, in InsertUserFull) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO users (
			id, email, username, first_name, last_name, phone, dob, address, gender, nik,
			password_hash, status, verified_at, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active', ?, NOW())
	`, in.ID, in.Email, in.Username, in.FirstName, in.LastName, in.Phone, in.DOB, in.Address, in.Gender, in.NIK,
		in.PasswordHash, in.VerifiedAt).Error
}

// ====== Untuk /v1/me (global) ======

func (r *Repository) GetUserRoleSlug(userID string) (string, error) { // role global (opsional)
	var slug string
	err := r.db.Raw(`
		SELECT r.slug
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = ?
		LIMIT 1
	`, userID).Scan(&slug).Error
	if err != nil {
		return "", err
	}
	return strings.ToUpper(slug), nil // <=== changed: pastikan UPPERCASE
}

type HospitalBrief struct {
	ID   string
	Code string
	Name string
}

func (r *Repository) ListHospitalsByUserID(userID string) ([]HospitalBrief, error) {
	var rows []HospitalBrief
	err := r.db.Raw(`
		SELECT h.id, h.code, h.name
		FROM user_hospitals uh
		JOIN hospitals h ON h.id = uh.hospital_id
		WHERE uh.user_id = ?
	`, userID).Scan(&rows).Error
	return rows, err
}

// ====== Untuk /v1/tenant/me (scoped) ======

// ResolveHospitalHint: hint bisa berupa UUID (id) atau code (string).
func (r *Repository) ResolveHospitalHint(hint string) (*HospitalBrief, error) {
	type row struct{ ID, Code, Name string }
	var out row
	// coba cocokkan ID (UUID)
	err := r.db.Raw(`
		SELECT id, code, name FROM hospitals
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`, hint).Scan(&out).Error
	if err == nil && out.ID != "" {
		return &HospitalBrief{ID: out.ID, Code: out.Code, Name: out.Name}, nil
	}
	// fallback ke CODE
	out = row{}
	err = r.db.Raw(`
		SELECT id, code, name FROM hospitals
		WHERE code = ? AND deleted_at IS NULL
		LIMIT 1
	`, hint).Scan(&out).Error
	if err != nil {
		return nil, err
	}
	if out.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &HospitalBrief{ID: out.ID, Code: out.Code, Name: out.Name}, nil
}

func (r *Repository) IsMemberOfHospital(userID, hospitalID string) (bool, error) {
	var exists bool
	err := r.db.Raw(`
		SELECT EXISTS(
			SELECT 1 FROM user_hospitals
			WHERE user_id = ? AND hospital_id = ?
		)
	`, userID, hospitalID).Scan(&exists).Error
	return exists, err
}

// Role user di hospital tertentu (dari hospital_user_roles).
func (r *Repository) GetHospitalRoleSlug(userID, hospitalID string) (string, error) { // <=== changed: normalisasi
	var slug string
	err := r.db.Raw(`
		SELECT r.slug
		FROM hospital_user_roles hur
		JOIN roles r ON r.id = hur.role_id
		WHERE hur.user_id = ? AND hur.hospital_id = ?
		LIMIT 1
	`, userID, hospitalID).Scan(&slug).Error
	if err != nil {
		return "", err
	}
	return strings.ToUpper(slug), nil // <=== changed: pastikan UPPERCASE
}
