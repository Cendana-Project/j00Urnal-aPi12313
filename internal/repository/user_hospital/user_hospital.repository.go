package user_hospital

import (
	"errors"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) UpsertMembership(userID, hospitalID string) error {
	var uh entity.UserHospital
	err := r.db.First(&uh, "user_id = ? AND hospital_id = ? AND deleted_at IS NULL", userID, hospitalID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		uh = entity.UserHospital{UserID: userID, HospitalID: hospitalID, IsActive: true}
		return r.db.Create(&uh).Error
	}
	if err != nil {
		return err
	}
	if !uh.IsActive {
		uh.IsActive = true
		return r.db.Save(&uh).Error
	}
	return nil
}

func (r *Repository) IsMember(userID, hospitalID string) (bool, error) {
	var c int64
	if err := r.db.Table("user_hospitals").
		Where("user_id = ? AND hospital_id = ? AND deleted_at IS NULL AND is_active = TRUE", userID, hospitalID).
		Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}
