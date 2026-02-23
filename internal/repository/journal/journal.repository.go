package journal

import (
"github.com/api-monolith-template/internal/infrastructure"
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{  }

func NewRepository(db *gorm.DB) *Repository { return &Repository{} }

func (r *Repository) Create(ctx context.Context, journal *entity.Journal) error {
	return infrastructure.GetDB().WithContext(ctx).Create(journal).Error
}

func (r *Repository) Update(ctx context.Context, journal *entity.Journal) error {
	return infrastructure.GetDB().WithContext(ctx).Save(journal).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*entity.Journal, error) {
	var journal entity.Journal
	err := infrastructure.GetDB().WithContext(ctx).
		Preload("Volumes").
		Preload("Volumes.Issues").
		First(&journal, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &journal, err
}

func (r *Repository) GetAll(ctx context.Context, status *constant.PublicationStatus, offset, limit int) ([]entity.Journal, int64, error) {
	var journals []entity.Journal
	var total int64

	query := infrastructure.GetDB().WithContext(ctx).Model(&entity.Journal{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&journals).Error
	return journals, total, err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return infrastructure.GetDB().WithContext(ctx).Delete(&entity.Journal{}, "id = ?", id).Error
}
