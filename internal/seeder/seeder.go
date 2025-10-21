package seeder

import (
	"fmt"
	"os"
	"slices"
	"time"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
)

// Run: seed roles → permissions (+mapping) → optional super_admin → sample users
func Run(db *gorm.DB) error {
	start := time.Now()

	if err := db.Transaction(func(tx *gorm.DB) error {
		// 1) roles (7 role termasuk super_admin)
		if err := SeedRoles(tx); err != nil {
			return fmt.Errorf("seed roles: %w", err)
		}
		// 2) permissions + mapping (mengacu constant.DefaultRolePermissions)
		if err := SeedPermissions(tx); err != nil {
			return fmt.Errorf("seed permissions: %w", err)
		}

		// 3) optional super_admin dari ENV (aktif)
		if email := os.Getenv("SUPERADMIN_EMAIL"); email != "" {
			pass := os.Getenv("SUPERADMIN_PASSWORD")
			if pass == "" {
				return fmt.Errorf("env SUPERADMIN_PASSWORD required when SUPERADMIN_EMAIL set")
			}
			first := os.Getenv("SUPERADMIN_FIRST_NAME")
			if first == "" {
				first = "Super"
			}
			last := os.Getenv("SUPERADMIN_LAST_NAME")
			if last == "" {
				last = "Admin"
			}
			if _, err := CreateUserActiveWithRole(tx, email, first, last, pass, constant.RoleSuperAdmin); err != nil {
				return fmt.Errorf("seed super_admin: %w", err)
			}
		}

		// 4) sample users (optional – aman idempotent)
		if err := SeedSampleUsers(tx); err != nil {
			return fmt.Errorf("seed sample users: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	fmt.Printf("[seeder] done in %s\n", time.Since(start))
	return nil
}

// Flush: bersihkan data hasil seeding (idempotent, tidak hapus user).
func Flush(db *gorm.DB) error {
	// role slugs yg kita seed
	roleSlugs := []string{
		constant.RoleSuperAdmin,
		constant.RoleAdmin,
		constant.RolePatient,
		constant.RoleDoctor,
		constant.RoleNurse,
		constant.RoleReceptionist,
		constant.RoleBOD,
	}

	// kumpulkan semua permission slugs dari peta default
	permSlugs := uniquePermSlugsFromDefaults()

	// ambil role ids
	type row struct{ ID string }
	var roles []row
	if err := db.Raw(`SELECT id FROM roles WHERE slug IN ?`, roleSlugs).Scan(&roles).Error; err != nil {
		return err
	}
	roleIDs := make([]any, 0, len(roles))
	for _, r := range roles {
		roleIDs = append(roleIDs, r.ID)
	}

	// ambil permission ids
	var perms []row
	if err := db.Raw(`SELECT id FROM permissions WHERE slug IN ?`, permSlugs).Scan(&perms).Error; err != nil {
		return err
	}
	permIDs := make([]any, 0, len(perms))
	for _, p := range perms {
		permIDs = append(permIDs, p.ID)
	}

	// 1) hapus mapping role_permissions
	if len(roleIDs) > 0 || len(permIDs) > 0 {
		if err := db.Exec(`DELETE FROM role_permissions
                           WHERE (role_id IN (?)) OR (permission_id IN (?))`, roleIDs, permIDs).Error; err != nil {
			return err
		}
	}

	// 2) putus mapping user_roles untuk role seeded (jangan hapus user)
	if len(roleIDs) > 0 {
		if err := db.Exec(`DELETE FROM user_roles WHERE role_id IN (?)`, roleIDs).Error; err != nil {
			return err
		}
	}

	// 3) hapus permissions hasil seed
	if len(permIDs) > 0 {
		if err := db.Exec(`DELETE FROM permissions WHERE id IN (?)`, permIDs).Error; err != nil {
			return err
		}
	}

	// 4) hapus roles hasil seed
	if len(roleIDs) > 0 {
		if err := db.Exec(`DELETE FROM roles WHERE id IN (?)`, roleIDs).Error; err != nil {
			return err
		}
	}

	return nil
}

// uniquePermSlugsFromDefaults mengumpulkan & meng-unique-kan slug dari constant.DefaultRolePermissions
func uniquePermSlugsFromDefaults() []string {
	var all []string
	for _, slugs := range constant.DefaultRolePermissions {
		all = append(all, slugs...)
	}
	slices.Sort(all)
	all = slices.Compact(all)
	return all
}
