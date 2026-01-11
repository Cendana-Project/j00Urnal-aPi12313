package manuscript

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

func (r *Repository) Create(ctx context.Context, manuscript *entity.Manuscript) error {
	return r.db.WithContext(ctx).Create(manuscript).Error
}

func (r *Repository) Update(ctx context.Context, manuscript *entity.Manuscript) error {
	return r.db.WithContext(ctx).Save(manuscript).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*entity.Manuscript, error) {
	var manuscript entity.Manuscript
	err := r.db.WithContext(ctx).
		Preload("Authors").
		Preload("Files").
		Preload("MainAuthor").
		Preload("Issue.Volume.Journal").
		First(&manuscript, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &manuscript, err
}

func (r *Repository) ListByIssue(ctx context.Context, issueID string) ([]entity.Manuscript, error) {
	var manuscripts []entity.Manuscript
	err := r.db.WithContext(ctx).
		Where("volume_number_id = ?", issueID).
		Preload("Authors").
		Preload("MainAuthor").
		Preload("Files").
		Order("created_at DESC").
		Find(&manuscripts).Error
	return manuscripts, err
}

func (r *Repository) ListByMainAuthor(ctx context.Context, authorID string) ([]entity.Manuscript, error) {
	var manuscripts []entity.Manuscript
	err := r.db.WithContext(ctx).
		Where("main_author_id = ?", authorID).
		Preload("Files").
		Find(&manuscripts).Error
	return manuscripts, err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Manuscript{}, "id = ?", id).Error
}

// Author Methods
func (r *Repository) AddAuthor(ctx context.Context, author *entity.ManuscriptAuthor) error {
	return r.db.WithContext(ctx).Create(author).Error
}

func (r *Repository) UpdateAuthors(ctx context.Context, manuscriptID string, authors []entity.ManuscriptAuthor) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing authors
		if err := tx.Where("manuscript_id = ?", manuscriptID).Delete(&entity.ManuscriptAuthor{}).Error; err != nil {
			return err
		}
		// Insert new authors
		for i := range authors {
			authors[i].ManuscriptID = manuscriptID
		}
		if len(authors) > 0 {
			return tx.Create(&authors).Error
		}
		return nil
	})
}

// File Methods
func (r *Repository) AddFile(ctx context.Context, file *entity.ManuscriptFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *Repository) GetLatestMainFileVersion(ctx context.Context, manuscriptID string) (int, error) {
	var version int
	err := r.db.WithContext(ctx).
		Model(&entity.ManuscriptFile{}).
		Where("manuscript_id = ? AND file_type = 'MAIN'", manuscriptID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&version).Error
	return version, err
}
