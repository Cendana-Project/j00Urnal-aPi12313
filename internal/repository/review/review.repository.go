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
	"github.com/api-monolith-template/internal/util"
	"github.com/api-monolith-template/pkg/pagination"
)

type Repository struct{}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{}
}

// ====== Review Round ======

func (r *Repository) CreateRound(ctx context.Context, round *entity.ReviewRound) error {
	return infrastructure.GetDB().WithContext(ctx).Create(round).Error
}

func (r *Repository) GetRoundByID(ctx context.Context, id string) (*entity.ReviewRound, error) {
	var round entity.ReviewRound
	err := infrastructure.GetDB().WithContext(ctx).
		Preload("Assignments.Reviewer").
		Preload("Assignments.Files").
		Preload("Manuscript.Journal").
		Preload("Creator").
		Preload("Files").
		First(&round, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &round, err
}

func (r *Repository) GetLatestRoundByManuscript(ctx context.Context, manuscriptID string) (*entity.ReviewRound, error) {
	var round entity.ReviewRound
	err := infrastructure.GetDB().WithContext(ctx).
		Where("manuscript_id = ?", manuscriptID).
		Order("round_number DESC").
		Preload("Assignments.Reviewer").
		Preload("Assignments.Files").
		Preload("Creator").
		First(&round).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &round, err
}

func (r *Repository) ListRoundsByManuscript(ctx context.Context, manuscriptID string) ([]entity.ReviewRound, error) {
	var rounds []entity.ReviewRound
	err := infrastructure.GetDB().WithContext(ctx).
		Where("manuscript_id = ?", manuscriptID).
		Order("round_number ASC").
		Preload("Assignments.Reviewer").
		Preload("Assignments.Files").
		Preload("Creator").
		Preload("Files").
		Find(&rounds).Error
	return rounds, err
}

func (r *Repository) UpdateRound(ctx context.Context, id string, updates map[string]any) error {
	updates["updated_at"] = time.Now()
	return infrastructure.GetDB().WithContext(ctx).Model(&entity.ReviewRound{}).
		Where("id = ?", id).Updates(updates).Error
}

// ====== Review Assignment ======

func (r *Repository) CreateAssignment(ctx context.Context, assignment *entity.ReviewAssignment) error {
	return infrastructure.GetDB().WithContext(ctx).Create(assignment).Error
}

func (r *Repository) GetAssignmentByID(ctx context.Context, id string) (*entity.ReviewAssignment, error) {
	uid, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, nil
	}
	idLit := uid.String()
	var assignment entity.ReviewAssignment
	// Raw without bind placeholders: PgBouncer transaction pooling drops unnamed prepared statements
	// used by GORM First/Preload ("pq: unnamed prepared statement does not exist").
	err = infrastructure.GetDB().WithContext(ctx).Raw(
		`SELECT * FROM review_assignments WHERE id = '` + idLit + `'::uuid LIMIT 1`,
	).Scan(&assignment).Error
	if err != nil {
		return nil, err
	}
	if assignment.ID == "" {
		return nil, nil
	}
	if assignment.ReviewerID != nil {
		rid := strings.TrimSpace(*assignment.ReviewerID)
		if rid == "" {
			return &assignment, nil
		}
		if ruid, perr := uuid.Parse(rid); perr == nil {
			var rev entity.User
			_ = infrastructure.GetDB().WithContext(ctx).Raw(
				`SELECT * FROM users WHERE id = '` + ruid.String() + `'::uuid AND deleted_at IS NULL LIMIT 1`,
			).Scan(&rev).Error
			if rev.ID != "" {
				assignment.Reviewer = &rev
			}
		}
	}
	return &assignment, nil
}

func (r *Repository) GetAssignmentByToken(ctx context.Context, token string) (*entity.ReviewAssignment, error) {
	candidates := util.InvitationTokenCandidates(token)
	if len(candidates) == 0 {
		return nil, nil
	}
	db := infrastructure.GetDB().WithContext(ctx)
	for _, tok := range candidates {
		esc := strings.ReplaceAll(tok, "'", "''")
		var assignment entity.ReviewAssignment
		err := db.Raw(
			`SELECT * FROM review_assignments WHERE invitation_token = '` + esc + `' LIMIT 1`,
		).Scan(&assignment).Error
		if err != nil {
			return nil, err
		}
		if assignment.ID == "" {
			continue
		}
		if err := r.hydrateAssignmentInvitationGraph(ctx, &assignment); err != nil {
			return nil, err
		}
		return &assignment, nil
	}
	return nil, nil
}

// hydrateAssignmentInvitationGraph loads round, manuscript, journal, and round creator using literal UUID
// queries only (PgBouncer-safe). Matches prior Preload scope for invitation preview / complete / decline.
func (r *Repository) hydrateAssignmentInvitationGraph(ctx context.Context, a *entity.ReviewAssignment) error {
	if a == nil || strings.TrimSpace(a.ReviewRoundID) == "" {
		return nil
	}
	rid, err := uuid.Parse(strings.TrimSpace(a.ReviewRoundID))
	if err != nil {
		return nil
	}
	db := infrastructure.GetDB().WithContext(ctx)
	var round entity.ReviewRound
	if err := db.Raw(
		`SELECT * FROM review_rounds WHERE id = '` + rid.String() + `'::uuid LIMIT 1`,
	).Scan(&round).Error; err != nil {
		return err
	}
	if round.ID == "" {
		return nil
	}
	a.ReviewRound = &round

	if mid, err := uuid.Parse(strings.TrimSpace(round.ManuscriptID)); err == nil {
		var ms entity.Manuscript
		if err := db.Raw(
			`SELECT * FROM manuscripts WHERE id = '` + mid.String() + `'::uuid AND deleted_at IS NULL LIMIT 1`,
		).Scan(&ms).Error; err != nil {
			return err
		}
		if ms.ID != "" {
			round.Manuscript = &ms
			if jid, err := uuid.Parse(strings.TrimSpace(ms.JournalID)); err == nil {
				var j entity.Journal
				if err := db.Raw(
					`SELECT * FROM journals WHERE id = '` + jid.String() + `'::uuid AND deleted_at IS NULL LIMIT 1`,
				).Scan(&j).Error; err != nil {
					return err
				}
				if j.ID != "" {
					ms.Journal = &j
				}
			}
		}
	}
	if cid, err := uuid.Parse(strings.TrimSpace(round.CreatedBy)); err == nil {
		var creator entity.User
		if err := db.Raw(
			`SELECT * FROM users WHERE id = '` + cid.String() + `'::uuid AND deleted_at IS NULL LIMIT 1`,
		).Scan(&creator).Error; err != nil {
			return err
		}
		if creator.ID != "" {
			round.Creator = &creator
		}
	}
	return nil
}

func (r *Repository) ListAssignmentsByRound(ctx context.Context, roundID string) ([]entity.ReviewAssignment, error) {
	var assignments []entity.ReviewAssignment
	err := infrastructure.GetDB().WithContext(ctx).
		Where("review_round_id = ?", roundID).
		Preload("Reviewer").
		Preload("Assigner").
		Preload("Files").
		Order("created_at ASC").
		Find(&assignments).Error
	return assignments, err
}

func (r *Repository) UpdateAssignment(ctx context.Context, id string, updates map[string]any) error {
	updates["updated_at"] = time.Now()
	return infrastructure.GetDB().WithContext(ctx).Model(&entity.ReviewAssignment{}).
		Where("id = ?", id).Updates(updates).Error
}

// ====== Review File ======

func (r *Repository) CreateReviewFile(ctx context.Context, file *entity.ReviewFile) error {
	return infrastructure.GetDB().WithContext(ctx).Create(file).Error
}

func (r *Repository) ListFilesByRound(ctx context.Context, roundID string) ([]entity.ReviewFile, error) {
	var files []entity.ReviewFile
	err := infrastructure.GetDB().WithContext(ctx).
		Where("review_round_id = ?", roundID).
		Preload("Uploader").
		Order("uploaded_at ASC").
		Find(&files).Error
	return files, err
}

func (r *Repository) ListFilesByAssignment(ctx context.Context, assignmentID string) ([]entity.ReviewFile, error) {
	var files []entity.ReviewFile
	err := infrastructure.GetDB().WithContext(ctx).
		Where("review_assignment_id = ?", assignmentID).
		Preload("Uploader").
		Order("uploaded_at ASC").
		Find(&files).Error
	return files, err
}

// ====== Candidate Queries ======

type ReviewerCandidate struct {
	ID          string
	FirstName   *string
	LastName    *string
	Email       string
	ActiveCount int
	DoneCount   int
}

func (r *Repository) ListReviewerCandidates(ctx context.Context, search string, pg *pagination.Pagination) ([]ReviewerCandidate, int64, error) {
	var candidates []ReviewerCandidate
	var total int64

	baseQuery := `
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles rl ON rl.id = ur.role_id
		WHERE rl.slug = ? AND u.deleted_at IS NULL AND u.status = 'active'
	`
	args := []any{constant.RoleReviewer}

	if search != "" {
		baseQuery += ` AND (LOWER(u.first_name) LIKE LOWER(?) OR LOWER(u.last_name) LIKE LOWER(?) OR LOWER(u.email) LIKE LOWER(?))`
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}

	// Count
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := infrastructure.GetDB().WithContext(ctx).Raw("SELECT COUNT(DISTINCT u.id) "+baseQuery, countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch
	selectQuery := `
		SELECT DISTINCT u.id, u.first_name, u.last_name, u.email,
		COALESCE((SELECT COUNT(*) FROM review_assignments ra WHERE ra.reviewer_id = u.id AND ra.status IN ('INVITED','ACCEPTED')), 0) AS active_count,
		COALESCE((SELECT COUNT(*) FROM review_assignments ra WHERE ra.reviewer_id = u.id AND ra.status = 'COMPLETED'), 0) AS done_count
	` + baseQuery + ` ORDER BY u.first_name ASC LIMIT ? OFFSET ?`

	offset := 0
	limit := 10
	if pg != nil {
		if pg.PageSize > 0 {
			limit = pg.PageSize
		}
		if pg.Page > 0 {
			offset = (pg.Page - 1) * limit
		}
	}
	args = append(args, limit, offset)

	if err := infrastructure.GetDB().WithContext(ctx).Raw(selectQuery, args...).Scan(&candidates).Error; err != nil {
		return nil, 0, err
	}

	return candidates, total, nil
}

type EditorCandidate struct {
	ID        string
	FirstName *string
	LastName  *string
	Email     string
}

func (r *Repository) ListEditorCandidates(ctx context.Context, search string, pg *pagination.Pagination) ([]EditorCandidate, int64, error) {
	var candidates []EditorCandidate
	var total int64

	baseQuery := `
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles rl ON rl.id = ur.role_id
		WHERE rl.slug = ? AND u.deleted_at IS NULL AND u.status = 'active'
	`
	args := []any{constant.RoleEditor}

	if search != "" {
		baseQuery += ` AND (LOWER(u.first_name) LIKE LOWER(?) OR LOWER(u.last_name) LIKE LOWER(?) OR LOWER(u.email) LIKE LOWER(?))`
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}

	// Count
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := infrastructure.GetDB().WithContext(ctx).Raw("SELECT COUNT(DISTINCT u.id) "+baseQuery, countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch
	selectQuery := `SELECT DISTINCT u.id, u.first_name, u.last_name, u.email ` + baseQuery + ` ORDER BY u.first_name ASC LIMIT ? OFFSET ?`

	offset := 0
	limit := 10
	if pg != nil {
		if pg.PageSize > 0 {
			limit = pg.PageSize
		}
		if pg.Page > 0 {
			offset = (pg.Page - 1) * limit
		}
	}
	args = append(args, limit, offset)

	if err := infrastructure.GetDB().WithContext(ctx).Raw(selectQuery, args...).Scan(&candidates).Error; err != nil {
		return nil, 0, err
	}

	return candidates, total, nil
}

// CountPendingInvitationByRoundEmail counts INVITED rows for the same round and email.
// Callers should pass email already normalized (e.g. lower + trim). Uses Raw SQL because GORM's
// Model().Where() rejects LOWER(TRIM(...)) on invited_email with "invalid field".
func (r *Repository) CountPendingInvitationByRoundEmail(ctx context.Context, roundID, emailNorm string) (int64, error) {
	var n int64
	err := infrastructure.GetDB().WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM review_assignments
		WHERE review_round_id = ?
		  AND status = ?
		  AND LOWER(TRIM(COALESCE(invited_email, ''))) = ?`,
		roundID, constant.ReviewAssignmentStatusInvited, emailNorm,
	).Scan(&n).Error
	return n, err
}

// ReviewerHistoryRow is a flat row for the reviewer History tab.
type ReviewerHistoryRow struct {
	AssignmentID   string
	ManuscriptID   string
	AssignedAt     time.Time
	Section        string
	Title          string
	Recommendation *string
	EditorDecision *string
	CompletedAt    *time.Time
}

// ListReviewerHistory lists completed review assignments for a reviewer with manuscript and round decision.
func (r *Repository) ListReviewerHistory(ctx context.Context, reviewerID, search, recommendationEq, editorDecisionEq string, pg *pagination.Pagination) ([]ReviewerHistoryRow, int64, error) {
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
WHERE ra.reviewer_id = ? AND ra.status = ? AND m.deleted_at IS NULL
`
	args := []any{reviewerID, constant.ReviewAssignmentStatusCompleted}
	if search != "" {
		baseFrom += ` AND m.title ILIKE ?`
		args = append(args, "%"+search+"%")
	}
	if recommendationEq != "" {
		baseFrom += ` AND ra.recommendation = ?`
		args = append(args, recommendationEq)
	}
	if editorDecisionEq != "" {
		baseFrom += ` AND rr.editor_decision = ?`
		args = append(args, editorDecisionEq)
	}

	var total int64
	countSQL := `SELECT COUNT(1) ` + baseFrom
	if err := infrastructure.GetDB().WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	selectSQL := `
SELECT ra.id AS assignment_id, m.id AS manuscript_id, ra.created_at AS assigned_at,
       m.section AS section, m.title,
       ra.recommendation, rr.editor_decision, ra.completed_at
` + baseFrom + `
ORDER BY ra.completed_at DESC NULLS LAST, ra.created_at DESC
LIMIT ? OFFSET ?
`
	args = append(args, limit, offset)

	var rows []ReviewerHistoryRow
	if err := infrastructure.GetDB().WithContext(ctx).Raw(selectSQL, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
