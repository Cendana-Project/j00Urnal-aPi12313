package manuscript

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/infrastructure"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/pkg/pagination"
)

type Repository struct{}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{}
}

// =============================================================================
// Write operations — these use simple GORM calls (no Preload chain, no IN ?)
// and are safe from the prepared-statement corruption issue.
// =============================================================================

func (r *Repository) Create(ctx context.Context, manuscript *entity.Manuscript) error {
	return infrastructure.GetDB().WithContext(ctx).Create(manuscript).Error
}

func (r *Repository) Update(ctx context.Context, manuscript *entity.Manuscript) error {
	return infrastructure.GetDB().WithContext(ctx).Save(manuscript).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return infrastructure.GetDB().WithContext(ctx).Delete(&entity.Manuscript{}, "id = ?", id).Error
}

func (r *Repository) AddAuthor(ctx context.Context, author *entity.ManuscriptAuthor) error {
	return infrastructure.GetDB().WithContext(ctx).Create(author).Error
}

func (r *Repository) UpdateAuthors(ctx context.Context, manuscriptID string, authors []entity.ManuscriptAuthor) error {
	return infrastructure.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return r.replaceAuthorsTx(tx, manuscriptID, authors)
	})
}

func (r *Repository) replaceAuthorsTx(tx *gorm.DB, manuscriptID string, authors []entity.ManuscriptAuthor) error {
	if err := tx.Where("manuscript_id = ?", manuscriptID).Delete(&entity.ManuscriptAuthor{}).Error; err != nil {
		return err
	}
	for i := range authors {
		authors[i].ManuscriptID = manuscriptID
	}
	if len(authors) > 0 {
		return tx.Create(&authors).Error
	}
	return nil
}

func (r *Repository) CreateWithAuthors(ctx context.Context, manuscript *entity.Manuscript, coAuthors []entity.ManuscriptAuthor) error {
	return infrastructure.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(manuscript).Error; err != nil {
			return err
		}
		return r.replaceAuthorsTx(tx, manuscript.ID, coAuthors)
	})
}

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

// =============================================================================
// Read operations — rewritten with raw SQL + inline UUID literals
// to avoid the GORM Preload/Find prepared-statement corruption under concurrent
// load with lib/pq and PgBouncer transaction pooling.
// Pattern matches review/repository.workspace.go and user/user.repository.go.
// =============================================================================

// pqLit returns a safe inline UUID literal for use in raw SQL.
// If s is not a valid UUID it returns "(SELECT gen_random_uuid() LIMIT 0)"
// so the query returns zero rows rather than crashing.
func pqLit(s string) string {
	uid, err := uuid.Parse(strings.TrimSpace(s))
	if err != nil {
		return "(SELECT gen_random_uuid() LIMIT 0)"
	}
	return "'" + uid.String() + "'::uuid"
}

// pqLitPtr returns a safe inline UUID literal for a *string; empty/nil yields NULL.
func pqLitPtr(s *string) string {
	if s == nil {
		return "NULL"
	}
	return pqLit(*s)
}

// pqStrLit returns a single-quote-escaped inline string literal for raw SQL.
func pqStrLit(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// db returns a fresh GORM session with the request context.
func db(ctx context.Context) *gorm.DB {
	return infrastructure.GetDB().WithContext(ctx)
}

// GetByID fetches a single manuscript with all related entities hydrated.
func (r *Repository) GetByID(ctx context.Context, id string) (*entity.Manuscript, error) {
	idLit := pqLit(id)

	// 1. Fetch manuscript row
	var m entity.Manuscript
	if err := db(ctx).Raw(`
		SELECT * FROM manuscripts 
		WHERE id = ` + idLit + ` 
		  AND deleted_at IS NULL 
		LIMIT 1
	`).Scan(&m).Error; err != nil {
		return nil, err
	}
	if m.ID == "" {
		return nil, nil
	}

	// 2. Hydrate relationships
	r.hydrateManuscript(ctx, &m)
	return &m, nil
}

// hydrateManuscript loads all related entities for a single manuscript in place.
func (r *Repository) hydrateManuscript(ctx context.Context, m *entity.Manuscript) {
	idLit := pqLit(m.ID)

	// Authors
	var authors []entity.ManuscriptAuthor
	if err := db(ctx).Raw(`
		SELECT * FROM manuscript_authors 
		WHERE manuscript_id = ` + idLit + ` 
		ORDER BY order_position ASC
	`).Scan(&authors).Error; err == nil && len(authors) > 0 {
		m.Authors = authors
	}

	// Files (order by uploaded_at DESC — matches original GORM Preload)
	var files []entity.ManuscriptFile
	if err := db(ctx).Raw(`
		SELECT * FROM manuscript_files 
		WHERE manuscript_id = ` + idLit + ` 
		ORDER BY uploaded_at DESC
	`).Scan(&files).Error; err == nil && len(files) > 0 {
		m.Files = files
	}

	// User references (MainAuthor, SubmittedBy, AssignedEditor)
	r.hydrateUserRefs(ctx, m, m.MainAuthorID, m.SubmittedByUserID, m.AssignedEditorID)

	// Issue -> Volume -> Journal chain
	if m.IssueID != nil && *m.IssueID != "" {
		r.hydrateIssueChain(ctx, m)
	}

	// Journal (direct FK, not via Issue)
	if m.JournalID != "" {
		r.hydrateJournal(ctx, m)
	}
}

// hydrateUserRefs loads up to three user references for a manuscript.
func (r *Repository) hydrateUserRefs(ctx context.Context, m *entity.Manuscript, userIDs ...*string) {
	// Collect non-nil user IDs
	ids := make([]string, 0, len(userIDs))
	for _, uid := range userIDs {
		if uid != nil && *uid != "" {
			ids = append(ids, *uid)
		}
	}
	if len(ids) == 0 {
		return
	}

	// Deduplicate
	seen := make(map[string]struct{}, len(ids))
	unique := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			uid, err := uuid.Parse(strings.TrimSpace(id))
			if err == nil {
				unique = append(unique, "'"+uid.String()+"'::uuid")
			}
		}
	}
	if len(unique) == 0 {
		return
	}

	var users []entity.User
	if err := db(ctx).Raw(`
		SELECT * FROM users 
		WHERE id IN (` + strings.Join(unique, ",") + `) 
		  AND deleted_at IS NULL
	`).Scan(&users).Error; err != nil || len(users) == 0 {
		return
	}

	userMap := make(map[string]*entity.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	if m.MainAuthorID != nil {
		if u, ok := userMap[*m.MainAuthorID]; ok {
			m.MainAuthor = u
		}
	}
	if m.SubmittedByUserID != nil {
		if u, ok := userMap[*m.SubmittedByUserID]; ok {
			m.SubmittedBy = u
		}
	}
	if m.AssignedEditorID != nil {
		if u, ok := userMap[*m.AssignedEditorID]; ok {
			m.AssignedEditor = u
		}
	}
}

// hydrateIssueChain loads Issue -> Volume -> Journal for a manuscript that has IssueID set.
func (r *Repository) hydrateIssueChain(ctx context.Context, m *entity.Manuscript) {
	if m.IssueID == nil || *m.IssueID == "" {
		return
	}
	issueID := *m.IssueID

	// Fetch Issue
	var iss entity.Issue
	if err := db(ctx).Raw(`
		SELECT * FROM issues 
		WHERE id = ` + pqLit(issueID) + ` 
		  AND deleted_at IS NULL 
		LIMIT 1
	`).Scan(&iss).Error; err != nil || iss.ID == "" {
		return
	}
	m.Issue = &iss

	// Fetch Volume
	var vol entity.Volume
	if err := db(ctx).Raw(`
		SELECT * FROM volumes 
		WHERE id = ` + pqLit(iss.VolumeID) + ` 
		  AND deleted_at IS NULL 
		LIMIT 1
	`).Scan(&vol).Error; err != nil || vol.ID == "" {
		return
	}
	iss.Volume = &vol

	// Fetch Journal from Volume
	var j entity.Journal
	if err := db(ctx).Raw(`
		SELECT * FROM journals 
		WHERE id = ` + pqLit(vol.JournalID) + ` 
		  AND deleted_at IS NULL 
		LIMIT 1
	`).Scan(&j).Error; err != nil || j.ID == "" {
		return
	}
	vol.Journal = &j
}

// hydrateJournal loads the journal directly via JournalID FK.
func (r *Repository) hydrateJournal(ctx context.Context, m *entity.Manuscript) {
	if m.JournalID == "" {
		return
	}
	var j entity.Journal
	if err := db(ctx).Raw(`
		SELECT * FROM journals 
		WHERE id = ` + pqLit(m.JournalID) + ` 
		  AND deleted_at IS NULL 
		LIMIT 1
	`).Scan(&j).Error; err != nil || j.ID == "" {
		return
	}
	m.Journal = &j
}

// ListByIssue returns manuscripts for a given issue.
func (r *Repository) ListByIssue(ctx context.Context, issueID string) ([]entity.Manuscript, error) {
	var manuscripts []entity.Manuscript
	err := db(ctx).Raw(`
		SELECT * FROM manuscripts 
		WHERE volume_number_id = ` + pqLit(issueID) + ` 
		  AND deleted_at IS NULL 
		ORDER BY created_at DESC
	`).Scan(&manuscripts).Error
	if err != nil {
		return nil, err
	}

	r.hydrateManuscriptList(ctx, manuscripts)
	return manuscripts, nil
}

// ListByMainAuthor fetches manuscripts by author with filters, search, and pagination.
func (r *Repository) ListByMainAuthor(
	ctx context.Context,
	authorID string,
	req request.AuthorManuscriptFilterRequest,
	pg *pagination.Pagination,
) ([]entity.Manuscript, int64, error) {
	authorLit := pqLit(authorID)

	// Build WHERE clause
	var clauses []string
	var args []any

	clauses = append(clauses, fmt.Sprintf(
		"(m.main_author_id = %s OR m.submitted_by_user_id = %s)",
		authorLit, authorLit,
	))
	clauses = append(clauses, "m.deleted_at IS NULL")

	if len(req.Statuses) > 0 {
		inClause := make([]string, len(req.Statuses))
		for i, st := range req.Statuses {
			inClause[i] = pqStrLit(string(st))
		}
		clauses = append(clauses, "m.status IN ("+strings.Join(inClause, ",")+")")
	}
	if req.StartDate != nil && *req.StartDate != "" {
		clauses = append(clauses, "m.created_at >= $1")
		args = append(args, *req.StartDate)
	}
	if req.EndDate != nil && *req.EndDate != "" {
		idx := len(args) + 1
		clauses = append(clauses, fmt.Sprintf("m.created_at <= $%d", idx))
		args = append(args, *req.EndDate)
	}
	if req.SearchTitle != "" || req.SearchAuthor != "" {
		// Use bind params for user-supplied search text (safe: simple query after Raw)
		titleIdx := len(args) + 1
		authorIdx := titleIdx + 1
		clauses = append(clauses, fmt.Sprintf(
			`(m.title ILIKE $%d OR EXISTS (SELECT 1 FROM manuscript_authors ma WHERE ma.manuscript_id = m.id AND ma.author_name ILIKE $%d))`,
			titleIdx, authorIdx,
		))
		args = append(args, "%"+req.SearchTitle+"%", "%"+req.SearchAuthor+"%")
	}

	where := strings.Join(clauses, " AND ")

	// Normalize pagination
	page, pageSize := pagination.NormalizeQuery(pg.Page, pg.PageSize)
	offset := (page - 1) * pageSize

	// Count
	var total int64
	countSQL := `SELECT COUNT(1) FROM manuscripts m WHERE ` + where
	if err := db(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch manuscripts
	var manuscripts []entity.Manuscript
	selectSQL := `SELECT m.* FROM manuscripts m WHERE ` + where +
		` ORDER BY m.created_at DESC LIMIT ` + fmt.Sprintf("%d", pageSize) + ` OFFSET ` + fmt.Sprintf("%d", offset)

	if err := db(ctx).Raw(selectSQL, args...).Scan(&manuscripts).Error; err != nil {
		return nil, 0, err
	}

	if len(manuscripts) > 0 {
		r.hydrateManuscriptList(ctx, manuscripts)
	}

	return manuscripts, total, nil
}

// ListByStatuses returns paginated manuscripts filtered by statuses (chief editor view).
func (r *Repository) ListByStatuses(ctx context.Context, statuses []constant.ManuscriptStatus, pg *pagination.Pagination) ([]entity.Manuscript, int64, error) {
	// Build status IN clause with inline literals (safe: enum constants)
	inClause := make([]string, len(statuses))
	for i, st := range statuses {
		inClause[i] = pqStrLit(string(st))
	}
	statusWhere := "m.status IN (" + strings.Join(inClause, ",") + ")"

	return r.listWithWhere(ctx, statusWhere, pg)
}

// ListByAssignedEditor returns paginated manuscripts assigned to a specific editor.
func (r *Repository) ListByAssignedEditor(ctx context.Context, editorID string, statuses []constant.ManuscriptStatus, pg *pagination.Pagination) ([]entity.Manuscript, int64, error) {
	clauses := []string{
		fmt.Sprintf("m.assigned_editor_id = %s", pqLit(editorID)),
		"m.deleted_at IS NULL",
	}
	if len(statuses) > 0 {
		inClause := make([]string, len(statuses))
		for i, st := range statuses {
			inClause[i] = pqStrLit(string(st))
		}
		clauses = append(clauses, "m.status IN ("+strings.Join(inClause, ",")+")")
	}

	return r.listWithWhere(ctx, strings.Join(clauses, " AND "), pg)
}

// listWithWhere is a shared helper for ListByStatuses and ListByAssignedEditor.
func (r *Repository) listWithWhere(ctx context.Context, where string, pg *pagination.Pagination) ([]entity.Manuscript, int64, error) {
	page, pageSize := pagination.NormalizeQuery(pg.Page, pg.PageSize)
	offset := (page - 1) * pageSize

	// Count
	var total int64
	if err := db(ctx).Raw(
		`SELECT COUNT(1) FROM manuscripts m WHERE ` + where,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch manuscripts (inline limit/offset as integers — no bind params)
	var manuscripts []entity.Manuscript
	selectSQL := `SELECT m.* FROM manuscripts m WHERE ` + where +
		` ORDER BY m.created_at DESC LIMIT ` + fmt.Sprintf("%d", pageSize) + ` OFFSET ` + fmt.Sprintf("%d", offset)

	if err := db(ctx).Raw(selectSQL).Scan(&manuscripts).Error; err != nil {
		return nil, 0, err
	}

	if len(manuscripts) > 0 {
		r.hydrateManuscriptList(ctx, manuscripts)
	}

	return manuscripts, total, nil
}

// hydrateManuscriptList batch-loads related entities for a list of manuscripts.
func (r *Repository) hydrateManuscriptList(ctx context.Context, manuscripts []entity.Manuscript) {
	if len(manuscripts) == 0 {
		return
	}

	// Build UUID clause for manuscript IDs
	ids := make([]string, len(manuscripts))
	for i, m := range manuscripts {
		ids[i] = pqLit(m.ID)
	}
	idsClause := strings.Join(ids, ",")

	// Batch-load authors
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

	// Batch-load files
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

	// Collect unique user IDs
	userIDs := make(map[string]struct{})
	for _, m := range manuscripts {
		if m.MainAuthorID != nil {
			userIDs[*m.MainAuthorID] = struct{}{}
		}
		if m.SubmittedByUserID != nil {
			userIDs[*m.SubmittedByUserID] = struct{}{}
		}
		if m.AssignedEditorID != nil {
			userIDs[*m.AssignedEditorID] = struct{}{}
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
					if manuscripts[i].SubmittedByUserID != nil {
						if u, ok := userMap[*manuscripts[i].SubmittedByUserID]; ok {
							manuscripts[i].SubmittedBy = u
						}
					}
					if manuscripts[i].AssignedEditorID != nil {
						if u, ok := userMap[*manuscripts[i].AssignedEditorID]; ok {
							manuscripts[i].AssignedEditor = u
						}
					}
				}
			}
		}
	}

	// Batch-load journals
	journalIDs := make(map[string]struct{})
	for _, m := range manuscripts {
		if m.JournalID != "" {
			journalIDs[m.JournalID] = struct{}{}
		}
	}
	if len(journalIDs) > 0 {
		jidList := make([]string, 0, len(journalIDs))
		for jid := range journalIDs {
			if pjid, err := uuid.Parse(strings.TrimSpace(jid)); err == nil {
				jidList = append(jidList, "'"+pjid.String()+"'::uuid")
			}
		}
		if len(jidList) > 0 {
			var journals []entity.Journal
			if err := db(ctx).Raw(`
				SELECT * FROM journals 
				WHERE id IN (` + strings.Join(jidList, ",") + `) 
				  AND deleted_at IS NULL
			`).Scan(&journals).Error; err == nil {
				jMap := make(map[string]*entity.Journal, len(journals))
				for i := range journals {
					jMap[journals[i].ID] = &journals[i]
				}
				for i := range manuscripts {
					if j, ok := jMap[manuscripts[i].JournalID]; ok {
						manuscripts[i].Journal = j
					}
				}
			}
		}
	}
}
