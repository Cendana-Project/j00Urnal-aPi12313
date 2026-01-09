package volume

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, volume *entity.Volume) error {
	return r.db.WithContext(ctx).Create(volume).Error
}

func (r *Repository) Update(ctx context.Context, volume *entity.Volume) error {
	return r.db.WithContext(ctx).Save(volume).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*entity.Volume, error) {
	var volume entity.Volume
	err := r.db.WithContext(ctx).
		Preload("Journal").
		Preload("Issues").
		First(&volume, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &volume, err
}

func (r *Repository) Exists(ctx context.Context, journalID string, year, number int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Volume{}).
		Where("journal_id = ? AND year = ? AND number = ?", journalID, year, number).
		Count(&count).Error
	return count > 0, err
}
