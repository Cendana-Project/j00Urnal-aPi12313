package volume

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/infrastructure"
	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{}

func NewRepository(db *gorm.DB) *Repository { return &Repository{} }

func pqLit(s string) string {
	uid, err := uuid.Parse(strings.TrimSpace(s))
	if err != nil {
		return "(SELECT gen_random_uuid() LIMIT 0)"
	}
	return "'" + uid.String() + "'::uuid"
}

func db(ctx context.Context) *gorm.DB {
	return infrastructure.GetDB().WithContext(ctx)
}

func (r *Repository) Create(ctx context.Context, volume *entity.Volume) error {
	return db(ctx).Create(volume).Error
}

func (r *Repository) Update(ctx context.Context, volume *entity.Volume) error {
	return db(ctx).Save(volume).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*entity.Volume, error) {
	var volume entity.Volume
	err := db(ctx).Raw(`
		SELECT * FROM volumes
		WHERE id = ` + pqLit(id) + `
		  AND deleted_at IS NULL
		LIMIT 1
	`).Scan(&volume).Error
	if err != nil {
		return nil, err
	}
	if volume.ID == "" {
		return nil, nil
	}

	// Hydrate Journal
	if volume.JournalID != "" {
		var j entity.Journal
		if err := db(ctx).Raw(`
			SELECT * FROM journals
			WHERE id = ` + pqLit(volume.JournalID) + `
			  AND deleted_at IS NULL
			LIMIT 1
		`).Scan(&j).Error; err == nil && j.ID != "" {
			volume.Journal = &j
		}
	}

	// Hydrate Issues
	var issues []entity.Issue
	if err := db(ctx).Raw(`
		SELECT * FROM issues
		WHERE volume_id = ` + pqLit(volume.ID) + `
		  AND deleted_at IS NULL
		ORDER BY number ASC
	`).Scan(&issues).Error; err == nil && len(issues) > 0 {
		volume.Issues = issues
	}

	return &volume, nil
}

func (r *Repository) Exists(ctx context.Context, journalID string, year, number int) (bool, error) {
	var count int64
	err := db(ctx).Model(&entity.Volume{}).
		Where("journal_id = ? AND year = ? AND number = ?", journalID, year, number).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return db(ctx).Delete(&entity.Volume{}, "id = ?", id).Error
}

func (r *Repository) FindAll(ctx context.Context) ([]*entity.Volume, error) {
	var volumeValues []entity.Volume
	if err := db(ctx).Raw(`
		SELECT * FROM volumes
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`).Scan(&volumeValues).Error; err != nil {
		return nil, err
	}
	if len(volumeValues) == 0 {
		return nil, nil
	}

	// Convert to pointer slice
	volumes := make([]*entity.Volume, len(volumeValues))
	for i := range volumeValues {
		volumes[i] = &volumeValues[i]
	}

	// Batch-load journals
	journalIDs := make(map[string]struct{})
	for _, v := range volumes {
		if v != nil && v.JournalID != "" {
			journalIDs[v.JournalID] = struct{}{}
		}
	}
	if len(journalIDs) > 0 {
		jidList := make([]string, 0, len(journalIDs))
		for jid := range journalIDs {
			jidList = append(jidList, pqLit(jid))
		}
		var journals []entity.Journal
		if err := db(ctx).Raw(`
			SELECT * FROM journals
			WHERE id IN (` + strings.Join(jidList, ",") + `)
			  AND deleted_at IS NULL
		`).Scan(&journals).Error; err == nil {
			jMap := make(map[string]*entity.Journal)
			for i := range journals {
				jMap[journals[i].ID] = &journals[i]
			}
			for _, v := range volumes {
				if v != nil {
					if j, ok := jMap[v.JournalID]; ok {
						v.Journal = j
					}
				}
			}
		}
	}

	return volumes, nil
}
