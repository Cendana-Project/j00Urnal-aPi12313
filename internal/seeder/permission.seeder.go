package seeder

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
)

// SeedPermissions:
// 1) memastikan semua permission ada (idempotent)
// 2) assign permission ke setiap role sesuai DefaultRolePermissions
func SeedPermissions(db *gorm.DB) error {
	// --- 1) DAFTAR PERMISSION (rapi per domain) ---
	toCreate := []entity.Permission{
		// user & role & permission
		{Name: "User View", Slug: constant.PermissionUserView, IsActive: true},
		{Name: "User Create", Slug: constant.PermissionUserCreate, IsActive: true},
		{Name: "User Update", Slug: constant.PermissionUserUpdate, IsActive: true},
		{Name: "User Delete", Slug: constant.PermissionUserDelete, IsActive: true},

		{Name: "Role View", Slug: constant.PermissionRoleView, IsActive: true},
		{Name: "Role Assign", Slug: constant.PermissionRoleAssign, IsActive: true},

		{Name: "Permission View", Slug: constant.PermissionPermissionView, IsActive: true},

		// manuscript & journal
		{Name: "Manuscript Manage", Slug: constant.PermissionManuscriptManage, IsActive: true},
		{Name: "Journal Manage", Slug: constant.PermissionJournalManage, IsActive: true},
	}

	// upsert per slug (idempotent)
	for _, p := range toCreate {
		if err := db.Exec(`
			INSERT INTO permissions (name, slug, is_active, created_at)
			VALUES (?, ?, TRUE, NOW())
			ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name, is_active = TRUE;
		`, p.Name, p.Slug).Error; err != nil {
			return fmt.Errorf("insert permission %s: %w", p.Slug, err)
		}
	}

	// helper ambil id
	getRoleID := func(slug string) (string, error) {
		var id string
		if err := db.Raw(`SELECT id FROM roles WHERE slug = ? LIMIT 1`, slug).Scan(&id).Error; err != nil {
			return "", err
		}
		if id == "" {
			return "", fmt.Errorf("role not found: %s", slug)
		}
		return id, nil
	}
	getPermID := func(slug string) (string, error) {
		var id string
		if err := db.Raw(`SELECT id FROM permissions WHERE slug = ? LIMIT 1`, slug).Scan(&id).Error; err != nil {
			return "", err
		}
		if id == "" {
			return "", fmt.Errorf("permission not found: %s", slug)
		}
		return id, nil
	}

	// --- 2) ASSIGN PERMISSIONS KE ROLE ---
	for roleSlug, permSlugs := range constant.DefaultRolePermissions {
		rid, err := getRoleID(roleSlug)
		if err != nil {
			return err
		}
		for _, ps := range permSlugs {
			pid, err := getPermID(ps)
			if err != nil {
				return err
			}
			if err := db.Exec(`
				INSERT INTO role_permissions (role_id, permission_id, created_at)
				VALUES (?, ?, NOW())
				ON CONFLICT (role_id, permission_id) DO NOTHING;
			`, rid, pid).Error; err != nil {
				return fmt.Errorf("assign %s -> %s: %w", roleSlug, ps, err)
			}
		}
	}

	return nil
}
