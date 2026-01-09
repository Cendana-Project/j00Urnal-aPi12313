package publicationfile

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, file *entity.PublicationFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *Repository) GetByEntity(ctx context.Context, entityType constant.EntityType, entityID string, fileType constant.FileType) (*entity.PublicationFile, error) {
	var file entity.PublicationFile
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ? AND file_type = ?", entityType, entityID, fileType).
		Order("uploaded_at DESC").
		First(&file).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &file, err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.PublicationFile{}, "id = ?", id).Error
}
