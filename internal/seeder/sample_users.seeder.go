package seeder

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
)

func SeedSampleUsers(db *gorm.DB) error {
	type sample struct {
		Email     string
		FirstName string
		LastName  string
		Password  string
		RoleSlug  string
		Phone     string
		ID        string
		Gender    string // L|P
		NIK       string // 16 digit
		DOB       string // YYYY-MM-DD
		Address   string
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
		Email: "chief@medikaone.id", FirstName: "Chief", LastName: "Editor", Password: "Password123!", RoleSlug: constant.RoleChiefEditor, Phone: "081200000011", ID: genID(11),
	})

	// 4. Reviewers (ID: ...0020 - 0022)
	for i := 0; i < 3; i++ {
		uid := 20 + i
		users = append(users, sample{
			Email: fmt.Sprintf("reviewer%d@medikaone.id", i+1), FirstName: "Reviewer", LastName: fmt.Sprintf("%d", i+1), Password: "Password123!", RoleSlug: constant.RoleReviewer, Phone: fmt.Sprintf("0812000000%d", uid), ID: genID(uid),
		})
	}

	// 5. Authors (ID: ...0030 - 0032)
	for i := 0; i < 3; i++ {
		uid := 30 + i
		users = append(users, sample{
			Email: fmt.Sprintf("author%d@medikaone.id", i+1), FirstName: "Author", LastName: fmt.Sprintf("%d", i+1), Password: "Password123!", RoleSlug: constant.RoleAuthor, Phone: fmt.Sprintf("0812000000%d", uid), ID: genID(uid),
		})
	}

	for i, u := range users {
		created, err := CreateUserActiveWithRole(db, u.ID, u.Email, u.FirstName, u.LastName, u.Password, u.RoleSlug)
		if err != nil {
			return err
		}
		// var dobPtr *time.Time
		// if u.DOB != "" {
		// 	if tm, err := time.Parse("2006-01-02", u.DOB); err == nil {
		// 		dobPtr = &tm
		// 	}
		// }
		// Update phone (since it exists in schema)
		if u.Phone != "" {
			if err := db.Model(&entity.User{}).Where("id = ?", created.ID).Update("phone", u.Phone).Error; err != nil {
				return fmt.Errorf("update user phone idx %d (%s): %w", i, u.Email, err)
			}
		}

		/*
			// Fields not present in Users table schema yet:
			updates := map[string]any{
				"address": u.Address,
			}
			if u.Gender == "L" || u.Gender == "P" {
				updates["gender"] = u.Gender
			}
			if len(u.NIK) == 16 {
				updates["nik"] = u.NIK
			}
			if dobPtr != nil {
				updates["dob"] = dobPtr
			}
			if len(updates) > 0 {
				if err := db.Model(&entity.User{}).Where("id = ?", created.ID).Updates(updates).Error; err != nil {
					return fmt.Errorf("update user idx %d (%s): %w", i, u.Email, err)
				}
			}
		*/
	}
	return nil
}
