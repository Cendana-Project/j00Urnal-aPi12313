package seeder

import (
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
)

// Run adalah entry utama seeding: roles, permissions, mapping, optional super_admin via ENV, dan sample users.
func Run(db *gorm.DB) error {
	start := time.Now()

	// 1) Roles (7 role termasuk super_admin)
	if err := SeedRoles(db); err != nil {
		return fmt.Errorf("seed roles: %w", err)
	}

	// 2) Permissions (semua slug)
	if err := SeedPermissions(db); err != nil {
		return fmt.Errorf("seed permissions: %w", err)
	}

	// 3) Mapping role ↔ permissions (termasuk super_admin → all)
	if err := SeedRolePermissions(db); err != nil {
		return fmt.Errorf("seed role_permissions: %w", err)
	}

	// 4) Optional super_admin via ENV (aktif)
	if email := os.Getenv("SUPERADMIN_EMAIL"); email != "" {
		if pass := os.Getenv("SUPERADMIN_PASSWORD"); pass != "" {
			first := os.Getenv("SUPERADMIN_FIRST_NAME")
			if first == "" {
				first = "Super"
			}
			last := os.Getenv("SUPERADMIN_LAST_NAME")
			if last == "" {
				last = "Admin"
			}
			if _, err := CreateUserActiveWithRole(db, email, first, last, pass, constant.RoleSuperAdmin); err != nil {
				return fmt.Errorf("seed super_admin: %w", err)
			}
		}
	}

	// 5) Sample users: 1 super_admin, 1 admin, 1 nurse, 1 receptionist, 1 bod, 3 patient, 3 doctor
	if err := SeedSampleUsers(db); err != nil {
		return fmt.Errorf("seed sample users: %w", err)
	}

	fmt.Printf("[seeder] done in %s\n", time.Since(start))
	return nil
}

// Flush menghapus data hasil seeding (idempotent, aman):
// - Hapus mapping role_permissions & user_roles untuk role seeded,
// - Hapus permissions hasil seed,
// - Hapus roles hasil seed,
// - Tidak menghapus user; hanya memutus mapping user_roles yang terkait role seeded.
func Flush(db *gorm.DB) error {
	// Role slugs yang kita seed
	roleSlugs := []string{
		constant.RoleSuperAdmin,
		constant.RoleAdmin,
		constant.RolePatient,
		constant.RoleDoctor,
		constant.RoleNurse,
		constant.RoleReceptionist,
		constant.RoleBOD,
	}
	// Permission slugs hasil seed
	permSlugs := constant.AllPermissions()

	// Ambil role IDs
	var roles []entity.Role
	if err := db.Where("slug IN ?", roleSlugs).Find(&roles).Error; err != nil {
		return err
	}
	roleIDs := make([]any, 0, len(roles))
	for _, r := range roles {
		roleIDs = append(roleIDs, r.ID)
	}

	// Ambil permission IDs
	var perms []entity.Permission
	if err := db.Where("slug IN ?", permSlugs).Find(&perms).Error; err != nil {
		return err
	}
	permIDs := make([]any, 0, len(perms))
	for _, p := range perms {
		permIDs = append(permIDs, p.ID)
	}

	// 1) Hapus mapping role_permissions
	if len(roleIDs) > 0 || len(permIDs) > 0 {
		if err := db.Table("role_permissions").
			Where("(role_id IN ?)", roleIDs).
			Or("(permission_id IN ?)", permIDs).
			Delete(nil).Error; err != nil {
			return err
		}
	}

	// 2) Hapus mapping user_roles untuk role seeded (jangan hapus user)
	if len(roleIDs) > 0 {
		if err := db.Table("user_roles").
			Where("role_id IN ?", roleIDs).
			Delete(nil).Error; err != nil {
			return err
		}
	}

	// 3) Hapus permissions hasil seed
	if len(permIDs) > 0 {
		if err := db.Where("id IN ?", permIDs).Delete(&entity.Permission{}).Error; err != nil {
			return err
		}
	}

	// 4) Hapus roles hasil seed
	if len(roleIDs) > 0 {
		if err := db.Where("id IN ?", roleIDs).Delete(&entity.Role{}).Error; err != nil {
			return err
		}
	}

	return nil
}
