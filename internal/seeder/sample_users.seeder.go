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

	// super_admin (1)
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
		created, err := CreateUserActiveWithRole(db, u.Email, u.FirstName, u.LastName, u.Password, u.RoleSlug)
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
