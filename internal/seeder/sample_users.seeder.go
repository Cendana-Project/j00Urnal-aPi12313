package seeder

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
)

func SeedSampleUsers(db *gorm.DB) error {
	type sample struct {
		ID          string
		Email       string
		FirstName   string
		LastName    string
		Password    string
		RoleSlug    string
		Phone       string
		Affiliation string
	}

	// Helper to fmt ID
	genID := func(i int) string {
		return fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)
	}

	var users []sample

	// 1. Super Admin (ID: ...0010)
	users = append(users, sample{
		ID:          genID(10),
		Email:       "superadmin@journalapi.id",
		FirstName:   "Super",
		LastName:    "Admin",
		Password:    "Password123!",
		RoleSlug:    constant.RoleSuperAdmin,
		Phone:       "081270000001",
		Affiliation: "Journal API Team",
	})

	// 2. Editor (ID: ...0020)
	users = append(users, sample{
		ID:          genID(20),
		Email:       "editor001@journalapi.id",
		FirstName:   "Editor",
		LastName:    "One",
		Password:    "Password123!",
		RoleSlug:    constant.RoleEditor,
		Phone:       "081280000001",
		Affiliation: "Journal of Science",
	})

	// 3. Chief Editor (ID: ...0030)
	users = append(users, sample{
		ID:          genID(30),
		Email:       "chiefeditor@journalapi.id",
		FirstName:   "Chief",
		LastName:    "Editor",
		Password:    "Password123!",
		RoleSlug:    constant.RoleChiefEditor,
		Phone:       "081290000001",
		Affiliation: "University of Technology",
	})

	// 4. Author (ID: ...0040)
	users = append(users, sample{
		ID:          genID(40),
		Email:       "author001@journalapi.id",
		FirstName:   "Author",
		LastName:    "One",
		Password:    "Password123!",
		RoleSlug:    constant.RoleAuthor,
		Phone:       "081230000001",
		Affiliation: "Research Institute",
	})

	// 5. Reviewer (ID: ...0050)
	users = append(users, sample{
		ID:          genID(50),
		Email:       "reviewer001@journalapi.id",
		FirstName:   "Reviewer",
		LastName:    "One",
		Password:    "Password123!",
		RoleSlug:    constant.RoleReviewer,
		Phone:       "081240000001",
		Affiliation: "Peer Review Board",
	})

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