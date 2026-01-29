package seeder

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/model/entity"
)

type TermItem struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func SeedTerms(db *gorm.DB) error {
	// Check if any active term exists
	var count int64
	if err := db.Model(&entity.PublicationTerm{}).Where("is_active = ?", true).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // Already seeded or exists
	}

	// Default T&C Content as JSON List
	items := []TermItem{
		{
			ID:   "copyright",
			Text: "Yes, I agree to abide by the terms of the copyright statement.",
		},
		{
			ID:   "privacy",
			Text: "Yes, I agree to have my data collected and stored according to the privacy statement.",
		},
	}

	contentBytes, err := json.Marshal(items)
	if err != nil {
		return err
	}

	term := entity.PublicationTerm{
		Content:   string(contentBytes),
		Version:   1,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if err := db.Create(&term).Error; err != nil {
		return fmt.Errorf("failed to seed terms: %w", err)
	}

	fmt.Println("✅ Seeded default Terms & Conditions (v1)")
	return nil
}
