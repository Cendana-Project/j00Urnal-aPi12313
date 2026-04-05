package review

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/infrastructure"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/pkg/pagination"
)

// ReviewerAssignmentSummaryRow is a flat row for reviewer assignment lists.
type ReviewerAssignmentSummaryRow struct {
	AssignmentID     string
	ReviewRoundID    string
	ManuscriptID     string
	ManuscriptTitle  string
	ManuscriptSection string
	ReferenceNumber  *int64
	RoundNumber      int
	RoundStatus      string
	AssignmentStatus string
	DueDate          time.Time
	CreatedAt        time.Time
}

// ListReviewerAssignments lists assignments for a reviewer with manuscript summary (PgBouncer-safe literals).
func (r *Repository) ListReviewerAssignments(ctx context.Context, reviewerID string, assignmentStatus constant.ReviewAssignmentStatus, pg *pagination.Pagination) ([]ReviewerAssignmentSummaryRow, int64, error) {
	rid, err := uuid.Parse(strings.TrimSpace(reviewerID))
	if err != nil {
		return nil, 0, nil
	}
	ridLit := rid.String()
	st := string(assignmentStatus)

	limit := pagination.DefaultPageSize
	page := 1
	if pg != nil {
		if pg.PageSize > 0 {
			limit = pg.PageSize
			if limit > pagination.MaxPageSize {
				limit = pagination.MaxPageSize
			}
		}
		if pg.Page > 0 {
			page = pg.Page
		}
	}
	offset := (page - 1) * limit

	baseFrom := `
FROM review_assignments ra
JOIN review_rounds rr ON rr.id = ra.review_round_id
JOIN manuscripts m ON m.id = rr.manuscript_id
WHERE ra.reviewer_id = '` + ridLit + `'::uuid
  AND ra.status = '` + strings.ReplaceAll(st, "'", "''") + `'
  AND m.deleted_at IS NULL
`
	countSQL := `SELECT COUNT(1) ` + baseFrom
	var total int64
	if err := infrastructure.GetDB().WithContext(ctx).Raw(countSQL).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	selectSQL := `
SELECT ra.id AS assignment_id, ra.review_round_id AS review_round_id, m.id AS manuscript_id,
       m.title AS manuscript_title, COALESCE(NULLIF(TRIM(m.section), ''), '') AS manuscript_section,
       m.reference_number AS reference_number, rr.round_number AS round_number,
       rr.status AS round_status, ra.status AS assignment_status, ra.due_date AS due_date, ra.created_at AS created_at
` + baseFrom + `
ORDER BY ra.due_date ASC, ra.created_at ASC
LIMIT ? OFFSET ?
`
	var rows []ReviewerAssignmentSummaryRow
	if err := infrastructure.GetDB().WithContext(ctx).Raw(selectSQL, limit, offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// GetReportByAssignmentID loads the single report row for an assignment.
func (r *Repository) GetReportByAssignmentID(ctx context.Context, assignmentID string) (*entity.ReviewAssignmentReport, error) {
	return r.GetReportByAssignmentDB(ctx, infrastructure.GetDB().WithContext(ctx), assignmentID)
}

// UpsertReviewReport creates or updates the one report row per assignment.
func (r *Repository) UpsertReviewReport(ctx context.Context, rep *entity.ReviewAssignmentReport) error {
	return r.UpsertReviewReportDB(ctx, infrastructure.GetDB().WithContext(ctx), rep)
}

// UpsertReviewReportDB is like UpsertReviewReport but uses the provided DB handle (e.g. transaction).
func (r *Repository) UpsertReviewReportDB(ctx context.Context, db *gorm.DB, rep *entity.ReviewAssignmentReport) error {
	if rep == nil {
		return errors.New("nil report")
	}
	existing, err := r.GetReportByAssignmentDB(ctx, db, rep.ReviewAssignmentID)
	if err != nil {
		return err
	}
	now := time.Now()
	rep.UpdatedAt = &now
	if existing == nil {
		if rep.ID == "" {
			rep.ID = uuid.New().String()
		}
		if rep.CreatedAt.IsZero() {
			rep.CreatedAt = now
		}
		return db.WithContext(ctx).Create(rep).Error
	}
	return db.WithContext(ctx).Model(&entity.ReviewAssignmentReport{}).
		Where("review_assignment_id = ?", rep.ReviewAssignmentID).
		Updates(map[string]any{
			"status":         rep.Status,
			"schema_version": rep.SchemaVersion,
			"payload":        rep.Payload,
			"updated_at":     rep.UpdatedAt,
		}).Error
}

// GetReportByAssignmentDB loads the report row using db (e.g. inside a transaction).
func (r *Repository) GetReportByAssignmentDB(ctx context.Context, db *gorm.DB, assignmentID string) (*entity.ReviewAssignmentReport, error) {
	aid, err := uuid.Parse(strings.TrimSpace(assignmentID))
	if err != nil {
		return nil, nil
	}
	lit := aid.String()
	var rep entity.ReviewAssignmentReport
	err = db.WithContext(ctx).Raw(
		`SELECT * FROM review_assignment_reports WHERE review_assignment_id = '`+lit+`'::uuid LIMIT 1`,
	).Scan(&rep).Error
	if err != nil {
		return nil, err
	}
	if rep.ID == "" {
		return nil, nil
	}
	return &rep, nil
}

// LoadManuscriptAuthors lists authors for a manuscript (PgBouncer-safe UUID literal).
func (r *Repository) LoadManuscriptAuthors(ctx context.Context, manuscriptID string) ([]entity.ManuscriptAuthor, error) {
	mid, err := uuid.Parse(strings.TrimSpace(manuscriptID))
	if err != nil {
		return nil, nil
	}
	lit := mid.String()
	var authors []entity.ManuscriptAuthor
	err = infrastructure.GetDB().WithContext(ctx).Raw(
		`SELECT * FROM manuscript_authors WHERE manuscript_id = '`+lit+`'::uuid ORDER BY order_position ASC`,
	).Scan(&authors).Error
	return authors, err
}

// CreateExtensionRequest inserts an extension request row.
func (r *Repository) CreateExtensionRequest(ctx context.Context, req *entity.ReviewExtensionRequest) error {
	return infrastructure.GetDB().WithContext(ctx).Create(req).Error
}

// CountPendingExtensionRequestsByAssignment returns pending extension rows for an assignment.
func (r *Repository) CountPendingExtensionRequestsByAssignment(ctx context.Context, assignmentID string) (int64, error) {
	aid, err := uuid.Parse(strings.TrimSpace(assignmentID))
	if err != nil {
		return 0, nil
	}
	lit := aid.String()
	var n int64
	err = infrastructure.GetDB().WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM review_extension_requests
		WHERE review_assignment_id = '`+lit+`'::uuid AND status = 'PENDING'`,
	).Scan(&n).Error
	return n, err
}
