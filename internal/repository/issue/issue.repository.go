package issue

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, issue *entity.Issue) error {
	return r.db.WithContext(ctx).Create(issue).Error
}

func (r *Repository) Update(ctx context.Context, issue *entity.Issue) error {
	return r.db.WithContext(ctx).Save(issue).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*entity.Issue, error) {
	var issue entity.Issue
	err := r.db.WithContext(ctx).
		Preload("Volume.Journal").
		First(&issue, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &issue, err
}
