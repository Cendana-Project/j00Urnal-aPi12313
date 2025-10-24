package role

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// =====================
// Global (non-tenant)
// =====================

func (r *Repository) FindBySlug(slug string) (*entity.Role, error) {
	var out entity.Role
	if err := r.db.Where("slug = ?", slug).First(&out).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *Repository) Assign(userID, roleID string) error {
	return r.db.Exec(`
		INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES (?, ?, NOW())
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userID, roleID).Error
}

func (r *Repository) UserHasRole(userID, roleSlug string) (bool, error) {
	var cnt int64
	err := r.db.Table("user_roles ur").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.slug = ?", userID, roleSlug).
		Count(&cnt).Error
	return cnt > 0, err
}

func (r *Repository) ListPermissionsByUser(ctx context.Context, userID string) ([]entity.Permission, error) {
	var perms []entity.Permission
	q := `
SELECT DISTINCT p.id, p.name, p.slug, p.description, p.is_active, p.created_at, p.updated_at, p.deleted_at
FROM user_roles ur
JOIN role_permissions rp ON rp.role_id = ur.role_id
JOIN permissions p ON p.id = rp.permission_id
WHERE ur.user_id = ? AND p.is_active = TRUE
`
	if err := r.db.WithContext(ctx).Raw(q, userID).Scan(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// Ambil role.id dari role.slug
func (r *Repository) GetRoleIDBySlug(ctx context.Context, slug string) (string, error) {
	var id string
	if err := r.db.WithContext(ctx).
		Raw(`SELECT id FROM roles WHERE slug = ? LIMIT 1`, slug).
		Scan(&id).Error; err != nil {
		return "", err
	}
	if id == "" {
		return "", gorm.ErrRecordNotFound
	}
	return id, nil
}

// Cek apakah user punya role global 'super_admin'
func (r *Repository) IsUserSuperAdmin(ctx context.Context, userID string) (bool, error) {
	type row struct{ C int64 }
	var out row
	q := `
SELECT COUNT(1) AS c
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = ? AND r.slug = 'super_admin'
`
	if err := r.db.WithContext(ctx).Raw(q, userID).Scan(&out).Error; err != nil {
		return false, err
	}
	return out.C > 0, nil
}

// =====================
// Tenant-scoped (hospital)
// =====================

// Assign role pada user di scope hospital (idempotent)
func (r *Repository) AssignHospitalRole(ctx context.Context, hospitalID, userID, roleID string) error {
	row := map[string]any{
		"hospital_id": hospitalID,
		"user_id":     userID,
		"role_id":     roleID,
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "hospital_id"}, {Name: "user_id"}, {Name: "role_id"}},
			DoNothing: true,
		}).
		Table("hospital_user_roles").
		Create(row).Error
}

// Daftar permissions user pada hospital tertentu
func (r *Repository) ListHospitalPermissionsByUser(ctx context.Context, hospitalID, userID string) ([]entity.Permission, error) {
	var perms []entity.Permission
	q := `
SELECT DISTINCT p.id, p.name, p.slug, p.description, p.is_active, p.created_at, p.updated_at, p.deleted_at
FROM hospital_user_roles hur
JOIN role_permissions rp ON rp.role_id = hur.role_id
JOIN permissions p ON p.id = rp.permission_id
WHERE hur.hospital_id = ? AND hur.user_id = ? AND p.is_active = TRUE
`
	if err := r.db.WithContext(ctx).Raw(q, hospitalID, userID).Scan(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}
