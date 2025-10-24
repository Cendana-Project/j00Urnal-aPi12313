package seeder

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type hospitalSeed struct {
	Code        string
	Name        string
	Address     string
	City        string
	Province    string
	Country     string
	Latitude    *float64
	Longitude   *float64
	Phone       string
	Description string
}

func SeedHospitals(db *gorm.DB) error {
	now := time.Now()

	lat1, lon1 := -6.200000, 106.816666 // Jakarta
	lat2, lon2 := -6.914744, 107.609810 // Bandung

	items := []hospitalSeed{
		{
			Code:        "HSP-MO-001",
			Name:        "MedikaOne General Hospital",
			Address:     "Jl. Kesehatan No. 1",
			City:        "Jakarta",
			Province:    "DKI Jakarta",
			Country:     "Indonesia",
			Latitude:    &lat1,
			Longitude:   &lon1,
			Phone:       "+62211234567",
			Description: "Rumah sakit umum MedikaOne",
		},
		{
			Code:        "HSP-MO-002",
			Name:        "MedikaOne Clinic Bandung",
			Address:     "Jl. Sehat No. 2",
			City:        "Bandung",
			Province:    "Jawa Barat",
			Country:     "Indonesia",
			Latitude:    &lat2,
			Longitude:   &lon2,
			Phone:       "+622287654321",
			Description: "Klinik MedikaOne Bandung",
		},
	}

	for _, h := range items {
		// upsert by code
		if err := db.Exec(`
			INSERT INTO hospitals (id, code, name, address, city, province, country, latitude, longitude, phone, description, is_active, created_at, updated_at)
			VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, true, ?, ?)
			ON CONFLICT (code) DO UPDATE
			SET name = EXCLUDED.name,
				address = EXCLUDED.address,
				city = EXCLUDED.city,
				province = EXCLUDED.province,
				country = EXCLUDED.country,
				latitude = EXCLUDED.latitude,
				longitude = EXCLUDED.longitude,
				phone = EXCLUDED.phone,
				description = EXCLUDED.description,
				is_active = true,
				updated_at = EXCLUDED.updated_at
		`, h.Code, h.Name, h.Address, h.City, h.Province, h.Country, h.Latitude, h.Longitude, h.Phone, h.Description, now, now).Error; err != nil {
			return fmt.Errorf("upsert hospital %s: %w", h.Code, err)
		}
	}
	return nil
}
