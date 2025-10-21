package role

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

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
	return r.db.Exec(`INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES (?, ?, NOW())
		ON CONFLICT (user_id, role_id) DO NOTHING`, userID, roleID).Error
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
	// Ambil kolom yg ada di entity.Permission (id, name, slug, description, is_active, created_at, updated_at, deleted_at)
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
