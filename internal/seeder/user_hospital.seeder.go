package seeder

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// helper
func getHospitalIDByCode(db *gorm.DB, code string) (string, error) {
	var id string
	if err := db.Raw(`SELECT id FROM hospitals WHERE code = ? LIMIT 1`, code).Scan(&id).Error; err != nil {
		return "", err
	}
	if id == "" {
		return "", fmt.Errorf("hospital not found: %s", code)
	}
	return id, nil
}

func getUserIDByEmail(db *gorm.DB, email string) (string, error) {
	var id string
	if err := db.Raw(`SELECT id FROM users WHERE email = ? LIMIT 1`, email).Scan(&id).Error; err != nil {
		return "", err
	}
	if id == "" {
		return "", fmt.Errorf("user not found: %s", email)
	}
	return id, nil
}

func SeedUserHospitals(db *gorm.DB) error {
	hid, err := getHospitalIDByCode(db, "HSP-MO-001")
	if err != nil {
		return err
	}

	emails := []string{
		"admin001@medikaone.id",
		"nurse001@medikaone.id",
		"receptionist001@medikaone.id",
		"bod001@medikaone.id",
		"doctor001@medikaone.id",
		"doctor002@medikaone.id",
		"doctor003@medikaone.id",
	}

	now := time.Now()
	for _, em := range emails {
		uid, err := getUserIDByEmail(db, em)
		if err != nil {
			return err
		}
		// asumsi schema: user_hospitals(user_id, hospital_id, created_at)
		if err := db.Exec(`
			INSERT INTO user_hospitals (user_id, hospital_id, created_at)
			VALUES (?, ?, ?)
			ON CONFLICT (user_id, hospital_id) DO NOTHING
		`, uid, hid, now).Error; err != nil {
			return fmt.Errorf("link user %s -> hospital %s: %w", em, "HSP-MO-001", err)
		}
	}

	return nil
}
