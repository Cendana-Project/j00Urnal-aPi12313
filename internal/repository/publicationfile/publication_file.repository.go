package publicationfile

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

func (r *Repository) Create(ctx context.Context, file *entity.PublicationFile) error {
	return infrastructure.GetDB().WithContext(ctx).Create(file).Error
}

func (r *Repository) GetByEntity(ctx context.Context, entityType constant.EntityType, entityID string, fileType constant.FileType) (*entity.PublicationFile, error) {
	var file entity.PublicationFile
	err := infrastructure.GetDB().WithContext(ctx).
		Where("entity_type = ? AND entity_id = ? AND file_type = ?", entityType, entityID, fileType).
		Order("uploaded_at DESC").
		First(&file).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &file, err
}

func (r *Repository) ListByEntity(ctx context.Context, entityType constant.EntityType, entityID string) ([]entity.PublicationFile, error) {
	var files []entity.PublicationFile
	err := infrastructure.GetDB().WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return infrastructure.GetDB().WithContext(ctx).Delete(&entity.PublicationFile{}, "id = ?", id).Error
}
