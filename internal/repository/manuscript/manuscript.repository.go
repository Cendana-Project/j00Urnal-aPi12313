package manuscript

import (
	"context"
	"errors"
	"time"

	"github.com/api-monolith-template/internal/infrastructure"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/pkg/pagination"
)

type Repository struct{}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{}
}

func (r *Repository) Create(ctx context.Context, manuscript *entity.Manuscript) error {
	return infrastructure.GetDB().WithContext(ctx).Create(manuscript).Error
}

func (r *Repository) Update(ctx context.Context, manuscript *entity.Manuscript) error {
	return infrastructure.GetDB().WithContext(ctx).Save(manuscript).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*entity.Manuscript, error) {
	var manuscript entity.Manuscript
	err := infrastructure.GetDB().WithContext(ctx).
		Preload("Authors").
		Preload("Files").
		Preload("MainAuthor").
		Preload("AssignedEditor").
		Preload("Issue.Volume.Journal").
		First(&manuscript, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &manuscript, err
}

func (r *Repository) ListByIssue(ctx context.Context, issueID string) ([]entity.Manuscript, error) {
	var manuscripts []entity.Manuscript
	err := infrastructure.GetDB().WithContext(ctx).
		Where("volume_number_id = ?", issueID).
		Preload("Authors").
		Preload("MainAuthor").
		Preload("Files").
		Order("created_at DESC").
		Find(&manuscripts).Error
	return manuscripts, err
}

func (r *Repository) ListByMainAuthor(
	ctx context.Context,
	authorID string,
	req request.AuthorManuscriptFilterRequest,
	pg *pagination.Pagination,
) ([]entity.Manuscript, int64, error) {
	var manuscripts []entity.Manuscript
	var total int64

	base := infrastructure.GetDB().WithContext(ctx).
		Model(&entity.Manuscript{}).
		Where("main_author_id = ?", authorID).
		Where("deleted_at IS NULL")

	if len(req.Statuses) > 0 {
		base = base.Where("status IN ?", req.Statuses)
	}
	if req.StartDate != nil && *req.StartDate != "" {
		base = base.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil && *req.EndDate != "" {
		base = base.Where("created_at <= ?", *req.EndDate)
	}

	// Search logic: either title matches or one of authors names match
	if req.SearchTitle != "" || req.SearchAuthor != "" {
		base = base.Where(
			"manuscripts.title ILIKE ? OR EXISTS (SELECT 1 FROM manuscript_authors ma WHERE ma.manuscript_id = manuscripts.id AND ma.author_name ILIKE ?)",
			"%"+req.SearchTitle+"%", "%"+req.SearchAuthor+"%",
		)
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := base.
		Scopes(pg.Paginate()).
		Preload("Files").
		Preload("Authors").
		Order("created_at DESC").
		Find(&manuscripts).Error

	return manuscripts, total, err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return infrastructure.GetDB().WithContext(ctx).Delete(&entity.Manuscript{}, "id = ?", id).Error
}

// Author Methods
func (r *Repository) AddAuthor(ctx context.Context, author *entity.ManuscriptAuthor) error {
	return infrastructure.GetDB().WithContext(ctx).Create(author).Error
}

func (r *Repository) UpdateAuthors(ctx context.Context, manuscriptID string, authors []entity.ManuscriptAuthor) error {
	return infrastructure.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
	return infrastructure.GetDB().WithContext(ctx).Create(file).Error
}

func (r *Repository) GetLatestMainFileVersion(ctx context.Context, manuscriptID string) (int, error) {
	var version int
	err := infrastructure.GetDB().WithContext(ctx).
		Model(&entity.ManuscriptFile{}).
		Where("manuscript_id = ? AND file_type = 'MAIN'", manuscriptID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&version).Error
	return version, err
}

// ====== Status & Editor Assignment (DRY: single place for status updates) ======

func (r *Repository) UpdateStatus(ctx context.Context, id string, status constant.ManuscriptStatus) error {
	now := time.Now()
	return infrastructure.GetDB().WithContext(ctx).Model(&entity.Manuscript{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": status, "updated_at": now}).Error
}

func (r *Repository) AssignEditor(ctx context.Context, manuscriptID, editorID string) error {
	now := time.Now()
	return infrastructure.GetDB().WithContext(ctx).Model(&entity.Manuscript{}).
		Where("id = ?", manuscriptID).
		Updates(map[string]any{
			"assigned_editor_id": editorID,
			"status":             constant.ManuscriptStatusAssignedToEditor,
			"updated_at":         now,
		}).Error
}

// ====== List queries with pagination ======

func (r *Repository) ListByStatuses(ctx context.Context, statuses []constant.ManuscriptStatus, pg *pagination.Pagination) ([]entity.Manuscript, int64, error) {
	var manuscripts []entity.Manuscript
	var total int64

	base := infrastructure.GetDB().WithContext(ctx).Model(&entity.Manuscript{}).
		Where("status IN ?", statuses).
		Where("deleted_at IS NULL")

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := base.
		Scopes(pg.Paginate()).
		Preload("MainAuthor").
		Preload("AssignedEditor").
		Preload("Journal").
		Order("created_at DESC").
		Find(&manuscripts).Error

	return manuscripts, total, err
}

func (r *Repository) ListByAssignedEditor(ctx context.Context, editorID string, statuses []constant.ManuscriptStatus, pg *pagination.Pagination) ([]entity.Manuscript, int64, error) {
	var manuscripts []entity.Manuscript
	var total int64

	base := infrastructure.GetDB().WithContext(ctx).Model(&entity.Manuscript{}).
		Where("assigned_editor_id = ?", editorID).
		Where("deleted_at IS NULL")

	if len(statuses) > 0 {
		base = base.Where("status IN ?", statuses)
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := base.
		Scopes(pg.Paginate()).
		Preload("MainAuthor").
		Preload("AssignedEditor").
		Preload("Authors").
		Preload("Journal").
		Order("created_at DESC").
		Find(&manuscripts).Error

	return manuscripts, total, err
}
