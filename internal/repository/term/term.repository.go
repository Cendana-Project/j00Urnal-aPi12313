package term

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetActive returns the currently active T&C (highest version active)
func (r *Repository) GetActive(ctx context.Context) (*entity.PublicationTerm, error) {
	var term entity.PublicationTerm
	// Logic: Find Active=true, order by Version DESC, Limit 1
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("version DESC").
		First(&term).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &term, err
}

func (r *Repository) Create(ctx context.Context, content string) (*entity.PublicationTerm, error) {
	return r.CreateWithVersion(ctx, content)
}

// CreateWithVersion creates a new term, auto-incrementing version
func (r *Repository) CreateWithVersion(ctx context.Context, content string) (*entity.PublicationTerm, error) {
	var latest entity.PublicationTerm
	var newVersion int = 1

	// Get latest version
	err := r.db.WithContext(ctx).Order("version DESC").First(&latest).Error
	if err == nil {
		newVersion = latest.Version + 1
	}

	term := &entity.PublicationTerm{
		Content:  content,
		Version:  newVersion,
		IsActive: true,
	}

	if err := r.db.WithContext(ctx).Create(term).Error; err != nil {
		return nil, err
	}
	return term, nil
}
