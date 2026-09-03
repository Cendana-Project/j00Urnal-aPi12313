package user

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/infrastructure"
	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{}

func NewRepository(db *gorm.DB) *Repository { return &Repository{} }

func (r *Repository) FindByEmail(email string) (*entity.User, error) {
	var u entity.User
	// Raw + Scan: avoids GORM First + soft-delete + PgBouncer "bind supplies N parameters" errors.
	err := infrastructure.GetDB().Raw(
		`SELECT * FROM users WHERE LOWER(email) = LOWER(?) AND deleted_at IS NULL LIMIT 1`,
		strings.TrimSpace(email),
	).Scan(&u).Error
	if err != nil {
		return nil, err
	}
	if u.ID == "" {
		return nil, nil
	}
	return &u, nil
}

func (r *Repository) FindByUsername(uname string) (*entity.User, error) {
	var u entity.User
	err := infrastructure.GetDB().Raw(
		`SELECT * FROM users WHERE LOWER(username) = LOWER(?) AND deleted_at IS NULL LIMIT 1`,
		strings.TrimSpace(uname),
	).Scan(&u).Error
	if err != nil {
		return nil, err
	}
	if u.ID == "" {
		return nil, nil
	}
	return &u, nil
}

func (r *Repository) ExistsUsername(uname string) (bool, error) {
	var cnt int64
	if err := infrastructure.GetDB().Model(&entity.User{}).
		Where("LOWER(username)=LOWER(?)", strings.ToLower(uname)).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (r *Repository) Create(u *entity.User) error { return infrastructure.GetDB().Create(u).Error }

func (r *Repository) MarkVerified(id string) error {
	now := time.Now()
	return infrastructure.GetDB().Model(&entity.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":      "active",
			"verified_at": now,
			"updated_at":  now,
		}).Error
}

func (r *Repository) UpdateByID(id string, fields map[string]any) error {
	fields["updated_at"] = time.Now()
	return infrastructure.GetDB().Model(&entity.User{}).Where("id = ?", id).Updates(fields).Error
}

func (r *Repository) GetByID(id string) (*entity.User, error) {
	uid, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, nil
	}
	// No bind parameters: avoids PgBouncer / unnamed-prepared-statement mix-ups
	// ("bind message supplies 1 parameters, but prepared statement requires 2").
	idLit := uid.String()
	var u entity.User
	err = infrastructure.GetDB().Raw(
		`SELECT * FROM users WHERE id = '` + idLit + `'::uuid AND deleted_at IS NULL LIMIT 1`,
	).Scan(&u).Error
	if err != nil {
		return nil, err
	}
	if u.ID == "" {
		return nil, nil
	}
	return &u, nil
}

func (r *Repository) GetByEmail(email string) (*entity.User, error) {
	var u entity.User
	err := infrastructure.GetDB().Raw(
		`SELECT * FROM users WHERE email = ? AND deleted_at IS NULL LIMIT 1`,
		strings.TrimSpace(email),
	).Scan(&u).Error
	if err != nil {
		return nil, err
	}
	if u.ID == "" {
		return nil, nil
	}
	return &u, nil
}

func (r *Repository) Update(u *entity.User) error { return infrastructure.GetDB().Save(u).Error }

func (r *Repository) Delete(id string) error {
	return infrastructure.GetDB().Delete(&entity.User{}, "id = ?", id).Error
}

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
	return infrastructure.GetDB().WithContext(ctx).Exec(`
		INSERT INTO users (id, email, username, phone, first_name, last_name, password_hash, status, verified_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'active', ?, NOW())
	`, in.ID, in.Email, in.Username, in.Phone, in.FirstName, in.LastName, in.PasswordHash, in.VerifiedAt).Error
}
func (r *Repository) ExistsNIK(nik string) (bool, error) {
	if strings.TrimSpace(nik) == "" {
		return false, nil
	}
	var cnt int64
	if err := infrastructure.GetDB().Model(&entity.User{}).
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
	return infrastructure.GetDB().WithContext(ctx).Exec(`
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
	uid, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return "", nil
	}

	// A user may legitimately have more than one role (for example AUTHOR + REVIEWER).
	// Fetch every active role without bound parameters so the result does not depend on
	// PostgreSQL's unspecified LIMIT order or PgBouncer prepared-statement state.
	var rows []struct {
		Slug string
	}
	err = infrastructure.GetDB().Raw(`
		SELECT r.slug
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = '` + uid.String() + `'::uuid
		  AND r.active = TRUE
		  AND r.deleted_at IS NULL
	`).Scan(&rows).Error
	if err != nil {
		return "", err
	}

	slugs := make([]string, 0, len(rows))
	for _, row := range rows {
		slugs = append(slugs, row.Slug)
	}
	return PrimaryGlobalRoleSlug(slugs), nil
}

// PrimaryGlobalRoleSlug defines the product's deterministic role used by clients that still
// consume the legacy single `role` field. Authorization continues to check the complete role
// set; this priority only controls navigation/presentation in /v1/me and auth responses.
func PrimaryGlobalRoleSlug(slugs []string) string {
	priority := map[string]int{
		constant.RoleSuperAdmin:  0,
		constant.RoleChiefEditor: 1,
		constant.RoleEditor:      2,
		constant.RoleReviewer:    3,
		constant.RoleAuthor:      4,
	}

	bestSlug := ""
	bestPriority := int(^uint(0) >> 1)
	for _, slug := range slugs {
		normalized := strings.ToUpper(strings.TrimSpace(slug))
		if normalized == "" {
			continue
		}
		p, known := priority[normalized]
		if !known {
			p = len(priority)
		}
		if p < bestPriority || (p == bestPriority && (bestSlug == "" || normalized < bestSlug)) {
			bestSlug = normalized
			bestPriority = p
		}
	}
	return bestSlug
}

type HospitalBrief struct {
	ID   string
	Code string
	Name string
}

func (r *Repository) ListHospitalsByUserID(userID string) ([]HospitalBrief, error) {
	var rows []HospitalBrief
	err := infrastructure.GetDB().Raw(`
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
	err := infrastructure.GetDB().Raw(`
		SELECT id, code, name FROM hospitals
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`, hint).Scan(&out).Error
	if err == nil && out.ID != "" {
		return &HospitalBrief{ID: out.ID, Code: out.Code, Name: out.Name}, nil
	}
	// fallback ke CODE
	out = row{}
	err = infrastructure.GetDB().Raw(`
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
	err := infrastructure.GetDB().Raw(`
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
	err := infrastructure.GetDB().Raw(`
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
