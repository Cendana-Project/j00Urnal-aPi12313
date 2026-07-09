package issue

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/google/uuid"

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

func (r *Repository) Create(ctx context.Context, issue *entity.Issue) error {
	return db(ctx).Create(issue).Error
}

func (r *Repository) Update(ctx context.Context, issue *entity.Issue) error {
	return db(ctx).Save(issue).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*entity.Issue, error) {
	var issue entity.Issue
	err := db(ctx).Raw(`
		SELECT * FROM issues
		WHERE id = ` + pqLit(id) + `
		  AND deleted_at IS NULL
		LIMIT 1
	`).Scan(&issue).Error
	if err != nil {
		return nil, err
	}
	if issue.ID == "" {
		return nil, nil
	}

	// Hydrate Volume -> Journal
	if issue.VolumeID != "" {
		var vol entity.Volume
		_ = db(ctx).Raw(`
			SELECT * FROM volumes
			WHERE id = ` + pqLit(issue.VolumeID) + `
			  AND deleted_at IS NULL
			LIMIT 1
		`).Scan(&vol)
		if vol.ID != "" {
			issue.Volume = &vol

			if vol.JournalID != "" {
				var j entity.Journal
				_ = db(ctx).Raw(`
					SELECT * FROM journals
					WHERE id = ` + pqLit(vol.JournalID) + `
					  AND deleted_at IS NULL
					LIMIT 1
				`).Scan(&j)
				if j.ID != "" {
					issue.Volume.Journal = &j
				}
			}
		}
	}

	// Hydrate manuscripts
	var manuscripts []entity.Manuscript
	_ = db(ctx).Raw(`
		SELECT * FROM manuscripts
		WHERE volume_number_id = ` + pqLit(issue.ID) + `
		  AND deleted_at IS NULL
		ORDER BY created_at DESC
	`).Scan(&manuscripts)
	if len(manuscripts) > 0 {
		r.hydrateManuscripts(ctx, manuscripts)
		issue.Manuscripts = manuscripts
	}

	return &issue, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return db(ctx).Delete(&entity.Issue{}, "id = ?", id).Error
}

func (r *Repository) FindAll(ctx context.Context) ([]*entity.Issue, error) {
	var issueValues []entity.Issue
	if err := db(ctx).Raw(`
		SELECT * FROM issues
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`).Scan(&issueValues).Error; err != nil {
		return nil, err
	}
	if len(issueValues) == 0 {
		return nil, nil
	}

	// Convert to pointer slice
	issues := make([]*entity.Issue, len(issueValues))
	for i := range issueValues {
		issues[i] = &issueValues[i]
	}

	// Batch-load volumes and journals
	volIDs := make([]string, len(issues))
	for i, iss := range issues {
		if iss != nil {
			volIDs[i] = pqLit(iss.VolumeID)
		}
	}
	volClause := strings.Join(volIDs, ",")

	var volumes []entity.Volume
	if err := db(ctx).Raw(`
		SELECT * FROM volumes
		WHERE id IN (` + volClause + `)
		  AND deleted_at IS NULL
	`).Scan(&volumes).Error; err == nil {
		volMap := make(map[string]*entity.Volume)
		for i := range volumes {
			volMap[volumes[i].ID] = &volumes[i]
		}
		for _, iss := range issues {
			if iss != nil {
				if v, ok := volMap[iss.VolumeID]; ok {
					iss.Volume = v
				}
			}
		}

		// Batch-load journals for volumes
		jIDs := make(map[string]struct{})
		for i := range volumes {
			if volumes[i].JournalID != "" {
				jIDs[volumes[i].JournalID] = struct{}{}
			}
		}
		if len(jIDs) > 0 {
			jidList := make([]string, 0, len(jIDs))
			for jid := range jIDs {
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
				for i := range volumes {
					if j, ok := jMap[volumes[i].JournalID]; ok {
						volumes[i].Journal = j
					}
				}
			}
		}
	}

	// Batch-load manuscripts per issue
	for _, iss := range issues {
		if iss == nil || iss.ID == "" {
			continue
		}
		var manuscripts []entity.Manuscript
		if err := db(ctx).Raw(`
			SELECT * FROM manuscripts
			WHERE volume_number_id = ` + pqLit(iss.ID) + `
			  AND deleted_at IS NULL
			ORDER BY created_at DESC
		`).Scan(&manuscripts).Error; err == nil && len(manuscripts) > 0 {
			r.hydrateManuscripts(ctx, manuscripts)
			iss.Manuscripts = manuscripts
		}
	}

	return issues, nil
}

// hydrateManuscripts loads authors + main author + files for a list of manuscripts in place.
func (r *Repository) hydrateManuscripts(ctx context.Context, manuscripts []entity.Manuscript) {
	if len(manuscripts) == 0 {
		return
	}
	ids := make([]string, len(manuscripts))
	for i, m := range manuscripts {
		ids[i] = pqLit(m.ID)
	}
	idsClause := strings.Join(ids, ",")

	// Authors
	var allAuthors []entity.ManuscriptAuthor
	if err := db(ctx).Raw(`
		SELECT * FROM manuscript_authors
		WHERE manuscript_id IN (` + idsClause + `)
		ORDER BY order_position ASC
	`).Scan(&allAuthors).Error; err == nil && len(allAuthors) > 0 {
		authorMap := make(map[string][]entity.ManuscriptAuthor)
		for _, a := range allAuthors {
			authorMap[a.ManuscriptID] = append(authorMap[a.ManuscriptID], a)
		}
		for i := range manuscripts {
			if a, ok := authorMap[manuscripts[i].ID]; ok {
				manuscripts[i].Authors = a
			}
		}
	}

	// Files
	var allFiles []entity.ManuscriptFile
	if err := db(ctx).Raw(`
		SELECT * FROM manuscript_files
		WHERE manuscript_id IN (` + idsClause + `)
		ORDER BY uploaded_at DESC
	`).Scan(&allFiles).Error; err == nil && len(allFiles) > 0 {
		fileMap := make(map[string][]entity.ManuscriptFile)
		for _, f := range allFiles {
			fileMap[f.ManuscriptID] = append(fileMap[f.ManuscriptID], f)
		}
		for i := range manuscripts {
			if f, ok := fileMap[manuscripts[i].ID]; ok {
				manuscripts[i].Files = f
			}
		}
	}

	// MainAuthor user
	userIDs := make(map[string]struct{})
	for _, m := range manuscripts {
		if m.MainAuthorID != nil {
			userIDs[*m.MainAuthorID] = struct{}{}
		}
	}
	if len(userIDs) > 0 {
		uidList := make([]string, 0, len(userIDs))
		for uid := range userIDs {
			if puid, err := uuid.Parse(strings.TrimSpace(uid)); err == nil {
				uidList = append(uidList, "'"+puid.String()+"'::uuid")
			}
		}
		if len(uidList) > 0 {
			var users []entity.User
			if err := db(ctx).Raw(`
				SELECT * FROM users
				WHERE id IN (` + strings.Join(uidList, ",") + `)
				  AND deleted_at IS NULL
			`).Scan(&users).Error; err == nil {
				userMap := make(map[string]*entity.User, len(users))
				for i := range users {
					userMap[users[i].ID] = &users[i]
				}
				for i := range manuscripts {
					if manuscripts[i].MainAuthorID != nil {
						if u, ok := userMap[*manuscripts[i].MainAuthorID]; ok {
							manuscripts[i].MainAuthor = u
						}
					}
				}
			}
		}
	}
}
