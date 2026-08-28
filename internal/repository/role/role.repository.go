package role

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant" // <=== added
	"github.com/api-monolith-template/internal/infrastructure"
	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{}

func NewRepository(db *gorm.DB) *Repository { return &Repository{} }

// pqLit returns a safe inline UUID literal, or a deliberately-empty-result fallback if s isn't a
// valid UUID. All queries below use inline literals instead of "?" placeholders: placeholder-bound
// Raw()/Where() calls have caused "unnamed prepared statement does not exist" errors under
// PgBouncer transaction pooling elsewhere in this codebase (see review/manuscript/issue/volume
// repositories), and the same failure mode was hit here too.
func pqLit(s string) string {
	uid, err := uuid.Parse(strings.TrimSpace(s))
	if err != nil {
		return "(SELECT gen_random_uuid() LIMIT 0)"
	}
	return "'" + uid.String() + "'::uuid"
}

// pqStrLit returns a single-quote-escaped inline string literal for raw SQL.
func pqStrLit(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// =====================
// Global (non-tenant)
// =====================

func (r *Repository) FindBySlug(slug string) (*entity.Role, error) {
	var out entity.Role
	q := `SELECT * FROM roles WHERE slug = ` + pqStrLit(slug) + ` LIMIT 1`
	if err := infrastructure.GetDB().Raw(q).Scan(&out).Error; err != nil {
		return nil, err
	}
	if out.ID == "" {
		return nil, nil
	}
	return &out, nil
}

func (r *Repository) Assign(userID, roleID string) error {
	q := `
		INSERT INTO user_roles (user_id, role_id, assigned_at)
		VALUES (` + pqLit(userID) + `, ` + pqLit(roleID) + `, NOW())
		ON CONFLICT (user_id, role_id) DO NOTHING
	`
	return infrastructure.GetDB().Exec(q).Error
}

func (r *Repository) UserHasRole(userID, roleSlug string) (bool, error) {
	var cnt int64
	q := `
SELECT COUNT(1) FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = ` + pqLit(userID) + ` AND r.slug = ` + pqStrLit(roleSlug)
	err := infrastructure.GetDB().Raw(q).Scan(&cnt).Error
	return cnt > 0, err
}

func (r *Repository) ListPermissionsByUser(ctx context.Context, userID string) ([]entity.Permission, error) {
	var perms []entity.Permission
	q := `
SELECT DISTINCT p.id, p.name, p.slug, p.description, p.is_active, p.created_at, p.updated_at, p.deleted_at
FROM user_roles ur
JOIN role_permissions rp ON rp.role_id = ur.role_id
JOIN permissions p ON p.id = rp.permission_id
WHERE ur.user_id = ` + pqLit(userID) + ` AND p.is_active = TRUE
`
	if err := infrastructure.GetDB().WithContext(ctx).Raw(q).Scan(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// Ambil role.id dari role.slug
func (r *Repository) GetRoleIDBySlug(ctx context.Context, slug string) (string, error) {
	var id string
	q := `SELECT id FROM roles WHERE UPPER(slug) = UPPER(` + pqStrLit(slug) + `) LIMIT 1`
	if err := infrastructure.GetDB().WithContext(ctx).Raw(q).Scan(&id).Error; err != nil {
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
WHERE ur.user_id = ` + pqLit(userID) + ` AND r.slug = ` + pqStrLit(constant.RoleSuperAdmin)
	if err := infrastructure.GetDB().WithContext(ctx).Raw(q).Scan(&out).Error; err != nil {
		return false, err
	}
	return out.C > 0, nil
}

// ListActiveEditorialLeads returns active Chief Editor and Super Admin users, used to broadcast
// notifications (e.g. new manuscript submitted, needs editor assignment) to whoever can act on them.
// Role slugs are fixed Go constants (not user input), so it's safe to inline them directly rather
// than use "?" placeholders — placeholder-bound Raw() queries have caused "unnamed prepared
// statement does not exist" errors under PgBouncer transaction pooling elsewhere in this codebase.
func (r *Repository) ListActiveEditorialLeads(ctx context.Context) ([]entity.User, error) {
	var users []entity.User
	q := `
SELECT DISTINCT u.* FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles rl ON rl.id = ur.role_id
WHERE rl.slug IN ('` + constant.RoleChiefEditor + `', '` + constant.RoleSuperAdmin + `') AND u.deleted_at IS NULL AND u.status = 'active'`
	if err := infrastructure.GetDB().WithContext(ctx).Raw(q).Scan(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// ListRolesByUser: daftar role global (aktif) milik user
func (r *Repository) ListRolesByUser(ctx context.Context, userID string) ([]entity.Role, error) {
	var roles []entity.Role
	q := `
SELECT r.id, r.name, r.slug, r.description, r.active, r.created_at, r.updated_at, r.deleted_at
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = ` + pqLit(userID) + ` AND r.active = TRUE
ORDER BY r.name`
	if err := infrastructure.GetDB().WithContext(ctx).Raw(q).Scan(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}
