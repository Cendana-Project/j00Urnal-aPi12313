package role

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant" // <=== added
	"github.com/api-monolith-template/internal/infrastructure"
	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{}

func NewRepository(db *gorm.DB) *Repository { return &Repository{} }

// =====================
// Global (non-tenant)
// =====================

func (r *Repository) FindBySlug(slug string) (*entity.Role, error) {
	var out entity.Role
	if err := infrastructure.GetDB().Where("slug = ?", slug).First(&out).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *Repository) Assign(userID, roleID string) error {
	return infrastructure.GetDB().Exec(`
		INSERT INTO user_roles (user_id, role_id, assigned_at)
		VALUES (?, ?, NOW())
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userID, roleID).Error
}

func (r *Repository) UserHasRole(userID, roleSlug string) (bool, error) {
	var cnt int64
	err := infrastructure.GetDB().Table("user_roles ur").
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
	if err := infrastructure.GetDB().WithContext(ctx).Raw(q, userID).Scan(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// Ambil role.id dari role.slug
func (r *Repository) GetRoleIDBySlug(ctx context.Context, slug string) (string, error) {
	var id string
	const q = `SELECT id FROM roles WHERE UPPER(slug) = UPPER(?) LIMIT 1` // <=== changed
	if err := infrastructure.GetDB().WithContext(ctx).Raw(q, slug).Scan(&id).Error; err != nil {
		return "", err
	}
	if id == "" {
		return "", gorm.ErrRecordNotFound
	}
	return id, nil
}

// Cek apakah user punya role global 'SUPER_ADMIN'
func (r *Repository) IsUserSuperAdmin(ctx context.Context, userID string) (bool, error) {
	type row struct{ C int64 }
	var out row
	q := `
SELECT COUNT(1) AS c
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = ? AND r.slug = ?` // <=== changed (pakai placeholder)
	if err := infrastructure.GetDB().WithContext(ctx).Raw(q, userID, constant.RoleSuperAdmin).Scan(&out).Error; err != nil { // <=== changed
		return false, err
	}
	return out.C > 0, nil
}

// ListActiveEditorialLeads returns active Chief Editor and Super Admin users, used to broadcast
// notifications (e.g. new manuscript submitted, needs editor assignment) to whoever can act on them.
func (r *Repository) ListActiveEditorialLeads(ctx context.Context) ([]entity.User, error) {
	var users []entity.User
	const q = `
SELECT DISTINCT u.* FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles rl ON rl.id = ur.role_id
WHERE rl.slug IN (?, ?) AND u.deleted_at IS NULL AND u.status = 'active'`
	if err := infrastructure.GetDB().WithContext(ctx).Raw(q, constant.RoleChiefEditor, constant.RoleSuperAdmin).Scan(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// ListRolesByUser: daftar role global (aktif) milik user
func (r *Repository) ListRolesByUser(ctx context.Context, userID string) ([]entity.Role, error) {
	var roles []entity.Role
	const q = `
SELECT r.id, r.name, r.slug, r.description, r.active, r.created_at, r.updated_at, r.deleted_at
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = ? AND r.active = TRUE
ORDER BY r.name`
	if err := infrastructure.GetDB().WithContext(ctx).Raw(q, userID).Scan(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}
