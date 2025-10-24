package seeder

import (
	"fmt"
	"time"

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
		Gender    string // L|P
		NIK       string // 16 digit
		DOB       string // YYYY-MM-DD
		Address   string
	}

	genNIK := func(prefix string, i int) string {
		base := fmt.Sprintf("%s%012d", prefix, i)
		if len(base) > 16 {
			return base[:16]
		}
		return base
	}

	var users []sample

	// super_admin (1)
	users = append(users, sample{
		Email:     "superadmin@medikaone.id",
		FirstName: "Super",
		LastName:  "Admin",
		Password:  "Password123",
		RoleSlug:  constant.RoleSuperAdmin,
		Phone:     "081270000001",
		Gender:    "L",
		NIK:       genNIK("1001", 1),
		DOB:       "1970-01-01",
		Address:   "Jl. Pusat No. 1, Jakarta",
	})

	// patient (3)
	for i := 1; i <= 3; i++ {
		users = append(users, sample{
			Email:     fmt.Sprintf("patient%03d@medikaone.id", i),
			FirstName: "Patient",
			LastName:  fmt.Sprintf("%03d", i),
			Password:  "Password123",
			RoleSlug:  constant.RolePatient,
			Phone:     fmt.Sprintf("081200000%03d", i),
			Gender:    []string{"L", "P"}[i%2],
			NIK:       genNIK("1101", i),
			DOB:       "1990-01-01",
			Address:   "Jl. Contoh No. 123, Jakarta",
		})
	}

	// doctor (3)
	for i := 1; i <= 3; i++ {
		users = append(users, sample{
			Email:     fmt.Sprintf("doctor%03d@medikaone.id", i),
			FirstName: "Doctor",
			LastName:  fmt.Sprintf("%03d", i),
			Password:  "Password123",
			RoleSlug:  constant.RoleDoctor,
			Phone:     fmt.Sprintf("081210000%03d", i),
			Gender:    []string{"L", "P"}[i%2],
			NIK:       genNIK("1201", i),
			DOB:       "1985-02-02",
			Address:   "Jl. Sehat No. 45, Jakarta",
		})
	}

	// staff (admin, nurse, receptionist, bod) masing-masing 1
	users = append(users,
		sample{Email: "admin001@medikaone.id", FirstName: "Admin", LastName: "001", Password: "Password123", RoleSlug: constant.RoleAdmin, Phone: "081230000001", Gender: "L", NIK: genNIK("1301", 1), DOB: "1980-03-03", Address: "Jl. Klinik No. 1, Jakarta"},
		sample{Email: "nurse001@medikaone.id", FirstName: "Nurse", LastName: "001", Password: "Password123", RoleSlug: constant.RoleNurse, Phone: "081240000001", Gender: "P", NIK: genNIK("1401", 1), DOB: "1992-04-04", Address: "Jl. Perawat No. 7, Jakarta"},
		sample{Email: "receptionist001@medikaone.id", FirstName: "Receptionist", LastName: "001", Password: "Password123", RoleSlug: constant.RoleReceptionist, Phone: "081250000001", Gender: "P", NIK: genNIK("1501", 1), DOB: "1993-05-05", Address: "Jl. Lobi No. 2, Jakarta"},
		sample{Email: "bod001@medikaone.id", FirstName: "BOD", LastName: "001", Password: "Password123", RoleSlug: constant.RoleBOD, Phone: "081260000001", Gender: "L", NIK: genNIK("1601", 1), DOB: "1975-06-06", Address: "Jl. Direktur No. 9, Jakarta"},
	)

	for i, u := range users {
		created, err := CreateUserActiveWithRole(db, u.Email, u.FirstName, u.LastName, u.Password, u.RoleSlug)
		if err != nil {
			return err
		}
		var dobPtr *time.Time
		if u.DOB != "" {
			if tm, err := time.Parse("2006-01-02", u.DOB); err == nil {
				dobPtr = &tm
			}
		}
		updates := map[string]any{
			"phone":   u.Phone,
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

		if err := db.Model(&entity.User{}).Where("id = ?", created.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("update user idx %d (%s): %w", i, u.Email, err)
		}
	}
	return nil
}
