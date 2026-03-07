package seeder

import (
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
)

// Run: seed roles → permissions (+mapping) → optional super_admin → sample users → hospitals → user_hospitals
func Run(db *gorm.DB) error {
	start := time.Now()

	if err := db.Transaction(func(tx *gorm.DB) error {
		// 1) roles
		if err := SeedRoles(tx); err != nil {
			return fmt.Errorf("seed roles: %w", err)
		}
		// 3) permissions + mapping
		if err := SeedPermissions(tx); err != nil {
			return fmt.Errorf("seed permissions: %w", err)
		}

		// 4) sample users (includes super admin)
		if err := SeedSampleUsers(tx); err != nil {
			return fmt.Errorf("seed sample users: %w", err)
		}

		// 5) Terms & Conditions
		if err := SeedTerms(tx); err != nil {
			return fmt.Errorf("seed terms: %w", err)
		}

		// 6) Sample Manuscripts
		if err := SeedSampleManuscripts(tx); err != nil {
			return fmt.Errorf("seed sample manuscripts: %w", err)
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
	// role slugs yg kita seed
	roleSlugs := []string{
		constant.RoleSuperAdmin,
		constant.RoleEditor,
		constant.RoleChiefEditor,
		constant.RoleAuthor,
		constant.RoleReviewer,
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

	// 5) flush manuscripts
	if err := FlushManuscripts(db); err != nil {
		return err
	}

	return nil
}

// uniquePermSlugsFromDefaults mengumpulkan & meng-unique-kan slug dari default
func uniquePermSlugsFromDefaults() []string {
	var all []string
	for _, slugs := range constant.DefaultRolePermissions {
		all = append(all, slugs...)
	}
	slices.Sort(all)
	all = slices.Compact(all)
	return all
}
