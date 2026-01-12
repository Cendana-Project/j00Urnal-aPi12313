package seeder

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
)

func SeedSampleUsers(db *gorm.DB) error {
	type sample struct {
		Email       string
		FirstName   string
		LastName    string
		Password    string
		RoleSlug    string
		Phone       string
		Affiliation string
	}

	var users []sample

	// 1. Super Admin (ID: ...0001) - handled in seeder.go if ENV set, but let's add default sample if not?
	// The seeder.go handles SuperAdmin via ENV. Here we seed others.
	// Actually current logic in seeder.go calls CreateUser for SuperAdmin.
	// We should avoid conflict. seeder.go uses ...0001. So start here from ...0002.

	// Helper to fmt ID
	genID := func(i int) string {
		return fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)
	}

	// 1. Super Admin (ID: ...0001)
	users = append(users, sample{
		Email: "superadmin@medikaone.id", FirstName: "Super", LastName: "Admin", Password: "Password123!", RoleSlug: constant.RoleSuperAdmin, Phone: "081270000001", ID: genID(1),
	})

	// 2. Editor (ID: ...0010)
	users = append(users, sample{
		Email: "editor@medikaone.id", FirstName: "Editor", LastName: "One", Password: "Password123!", RoleSlug: constant.RoleEditor, Phone: "081200000010", ID: genID(10),
	})

	// 3. Chief Editor (ID: ...0011)
	users = append(users, sample{
		Email:       "superadmin@journalapi.id",
		FirstName:   "Super",
		LastName:    "Admin",
		Password:    "Password123",
		RoleSlug:    constant.RoleSuperAdmin,
		Phone:       "081270000001",
		Affiliation: "Journal API Team",
	})

	// Add Editor and Chief Editor
	users = append(users,
		sample{Email: "editor001@journalapi.id", FirstName: "Editor", LastName: "One", Password: "Password123", RoleSlug: constant.RoleEditor, Phone: "081280000001", Affiliation: "Journal of Science"},
		sample{Email: "chiefeditor@journalapi.id", FirstName: "Chief", LastName: "Editor", Password: "Password123", RoleSlug: constant.RoleChiefEditor, Phone: "081290000001", Affiliation: "University of Technology"},
	)

	// regular users with different roles
	users = append(users,
		sample{Email: "admin001@journalapi.id", FirstName: "Admin", LastName: "001", Password: "Password123", RoleSlug: constant.RoleAdmin, Phone: "081230000001", Affiliation: "Faculty of Engineering"},
		sample{Email: "patient001@journalapi.id", FirstName: "User", LastName: "Patient", Password: "Password123", RoleSlug: constant.RolePatient, Phone: "081200000001", Affiliation: "Public User"},
		sample{Email: "doctor001@journalapi.id", FirstName: "User", LastName: "Doctor", Password: "Password123", RoleSlug: constant.RoleDoctor, Phone: "081210000001", Affiliation: "General Clinic"},
	)

	for i, u := range users {
		created, err := CreateUserActiveWithRole(db, u.ID, u.Email, u.FirstName, u.LastName, u.Password, u.RoleSlug)
		if err != nil {
			return err
		}

		updates := map[string]any{
			"phone":       u.Phone,
			"affiliation": u.Affiliation,
		}

		if err := db.Model(created).Updates(updates).Error; err != nil {
			return fmt.Errorf("update user idx %d (%s): %w", i, u.Email, err)
		}
	}
	return nil
}
