package review

import (
	"context"
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

// pqLit returns a safe inline UUID literal for raw SQL.
func pqLit(s string) string {
	uid, err := uuid.Parse(strings.TrimSpace(s))
	if err != nil {
		return "(SELECT gen_random_uuid() LIMIT 0)"
	}
	return "'" + uid.String() + "'::uuid"
}

// pqLitPtr returns a safe inline UUID literal for a *string; nil yields NULL.
func pqLitPtr(s *string) string {
	if s == nil {
		return "NULL"
	}
	return pqLit(*s)
}

func db(ctx context.Context) *gorm.DB {
	return infrastructure.GetDB().WithContext(ctx)
}

// ====== Review Round ======

func (r *Repository) CreateRound(ctx context.Context, round *entity.ReviewRound) error {
	return db(ctx).Create(round).Error
}

// GetRoundByID fetches a review round with all related entities hydrated via raw SQL.
func (r *Repository) GetRoundByID(ctx context.Context, id string) (*entity.ReviewRound, error) {
	return r.fetchRoundByQuery(ctx, "id = "+pqLit(id), "")
}

// GetLatestRoundByManuscript fetches the latest round (highest round_number) for a manuscript.
func (r *Repository) GetLatestRoundByManuscript(ctx context.Context, manuscriptID string) (*entity.ReviewRound, error) {
	round, err := r.fetchRoundByQuery(ctx, "manuscript_id = "+pqLit(manuscriptID), "ORDER BY round_number DESC LIMIT 1")
	if err != nil {
		return nil, err
	}
	return round, nil
}

// ListRoundsByManuscript lists all rounds for a manuscript, ordered by round_number ASC.
func (r *Repository) ListRoundsByManuscript(ctx context.Context, manuscriptID string) ([]entity.ReviewRound, error) {
	var rounds []entity.ReviewRound
	if err := db(ctx).Raw(`
		SELECT * FROM review_rounds
		WHERE manuscript_id = ` + pqLit(manuscriptID) + `
		ORDER BY round_number ASC
	`).Scan(&rounds).Error; err != nil {
		return nil, err
	}
	if len(rounds) == 0 {
		return rounds, nil
	}

	r.hydrateRounds(ctx, rounds)
	return rounds, nil
}

// fetchRoundByQuery fetches a single round with custom WHERE suffix (e.g. "ORDER BY ... LIMIT 1").
func (r *Repository) fetchRoundByQuery(ctx context.Context, where, orderSuffix string) (*entity.ReviewRound, error) {
	sql := `SELECT * FROM review_rounds WHERE ` + where
	if orderSuffix != "" {
		sql += ` ` + orderSuffix
	} else {
		sql += ` LIMIT 1`
	}

	var round entity.ReviewRound
	if err := db(ctx).Raw(sql).Scan(&round).Error; err != nil {
		return nil, err
	}
	if round.ID == "" {
		return nil, nil
	}

	// hydrateRounds works on the slice elements; pass pointer so modifications persist.
	rounds := []entity.ReviewRound{round}
	r.hydrateRounds(ctx, rounds)
	return &rounds[0], nil
}

// hydrateRounds loads all related entities for a list of rounds in place (assignments, files, users, etc.).
func (r *Repository) hydrateRounds(ctx context.Context, rounds []entity.ReviewRound) {
	if len(rounds) == 0 {
		return
	}

	// Build round IDs clause
	rids := make([]string, len(rounds))
	for i, rnd := range rounds {
		rids[i] = pqLit(rnd.ID)
	}
	ridsClause := strings.Join(rids, ",")

	// Batch-load assignments
	var allAssignments []entity.ReviewAssignment
	if err := db(ctx).Raw(`
		SELECT * FROM review_assignments
		WHERE review_round_id IN (` + ridsClause + `)
		ORDER BY created_at ASC
	`).Scan(&allAssignments).Error; err == nil && len(allAssignments) > 0 {
		asgnMap := make(map[string][]entity.ReviewAssignment)
		for _, a := range allAssignments {
			asgnMap[a.ReviewRoundID] = append(asgnMap[a.ReviewRoundID], a)
		}
		for i, rnd := range rounds {
			if a, ok := asgnMap[rnd.ID]; ok {
				rounds[i].Assignments = a
			}
		}

		// Hydrate per-round assignments (they're slices, not pointers, so do it per round)
		for i := range rounds {
			r.hydrateAssignments(ctx, rounds[i].Assignments)
		}
	}

	// Batch-load round files
	var allFiles []entity.ReviewFile
	if err := db(ctx).Raw(`
		SELECT * FROM review_files
		WHERE review_round_id IN (` + ridsClause + `)
		ORDER BY uploaded_at ASC
	`).Scan(&allFiles).Error; err == nil && len(allFiles) > 0 {
		fileMap := make(map[string][]entity.ReviewFile)
		for _, f := range allFiles {
			fileMap[f.ReviewRoundID] = append(fileMap[f.ReviewRoundID], f)
		}
		for i, rnd := range rounds {
			if f, ok := fileMap[rnd.ID]; ok {
				rounds[i].Files = f
			}
		}
	}

	// Collect user IDs referenced by CreatedBy
	creatorIDs := make([]string, 0)
	for _, rnd := range rounds {
		if rnd.CreatedBy != "" {
			creatorIDs = append(creatorIDs, rnd.CreatedBy)
		}
		// Also collect from Manuscript.CreatedBy via Journal
	}
	hydrateUsers := func(ids []string) map[string]*entity.User {
		if len(ids) == 0 {
			return nil
		}
		seen := make(map[string]struct{})
		unique := make([]string, 0)
		for _, id := range ids {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				unique = append(unique, pqLit(id))
			}
		}
		if len(unique) == 0 {
			return nil
		}
		var users []entity.User
		if err := db(ctx).Raw(`
			SELECT * FROM users
			WHERE id IN (` + strings.Join(unique, ",") + `)
			AND deleted_at IS NULL
		`).Scan(&users).Error; err != nil || len(users) == 0 {
			return nil
		}
		um := make(map[string]*entity.User, len(users))
		for i := range users {
			um[users[i].ID] = &users[i]
		}
		return um
	}

	um := hydrateUsers(creatorIDs)
	if um != nil {
		for i, rnd := range rounds {
			if u, ok := um[rnd.CreatedBy]; ok {
				rounds[i].Creator = u
			}
		}
	}

	// Hydrate manuscript for each round
	for i, rnd := range rounds {
		if rnd.ManuscriptID != "" {
			rounds[i] = r.hydrateRoundManuscript(ctx, rnd)
		}
	}
}

// hydrateRoundManuscript loads the manuscript (with journal) for a single round.
func (r *Repository) hydrateRoundManuscript(ctx context.Context, round entity.ReviewRound) entity.ReviewRound {
	var ms entity.Manuscript
	if err := db(ctx).Raw(`
		SELECT * FROM manuscripts
		WHERE id = ` + pqLit(round.ManuscriptID) + `
		  AND deleted_at IS NULL
		LIMIT 1
	`).Scan(&ms).Error; err != nil || ms.ID == "" {
		return round
	}

	// Load journal
	if ms.JournalID != "" {
		var j entity.Journal
		if err := db(ctx).Raw(`
			SELECT * FROM journals
			WHERE id = ` + pqLit(ms.JournalID) + `
			  AND deleted_at IS NULL
			LIMIT 1
		`).Scan(&j).Error; err == nil && j.ID != "" {
			ms.Journal = &j
		}
	}

	round.Manuscript = &ms
	return round
}

// hydrateAssignments loads reports and files for a list of assignments in place.
func (r *Repository) hydrateAssignments(ctx context.Context, assignments []entity.ReviewAssignment) {
	if len(assignments) == 0 {
		return
	}

	// Build assignment IDs clause
	aids := make([]string, len(assignments))
	for i, a := range assignments {
		aids[i] = pqLit(a.ID)
	}
	aidsClause := strings.Join(aids, ",")

	// Batch-load assigner users
	hydrateUser := func(assignments []entity.ReviewAssignment) {
		userIDs := make(map[string]struct{})
		for _, a := range assignments {
			if a.AssignedBy != "" {
				userIDs[a.AssignedBy] = struct{}{}
			}
			if a.ReviewerID != nil && *a.ReviewerID != "" {
				userIDs[*a.ReviewerID] = struct{}{}
			}
		}
		if len(userIDs) == 0 {
			return
		}
		uids := make([]string, 0, len(userIDs))
		for uid := range userIDs {
			uids = append(uids, pqLit(uid))
		}
		var users []entity.User
		if err := db(ctx).Raw(`
			SELECT * FROM users
			WHERE id IN (` + strings.Join(uids, ",") + `)
			  AND deleted_at IS NULL
		`).Scan(&users).Error; err != nil || len(users) == 0 {
			return
		}
		um := make(map[string]*entity.User, len(users))
		for i := range users {
			um[users[i].ID] = &users[i]
		}
		for i := range assignments {
			if u, ok := um[assignments[i].AssignedBy]; ok {
				assignments[i].Assigner = u
			}
			if assignments[i].ReviewerID != nil {
				if u, ok := um[*assignments[i].ReviewerID]; ok {
					assignments[i].Reviewer = u
				}
			}
		}
	}
	hydrateUser(assignments)

	// Batch-load reports
	var reports []entity.ReviewAssignmentReport
	if err := db(ctx).Raw(`
		SELECT * FROM review_assignment_reports
		WHERE review_assignment_id IN (` + aidsClause + `)
	`).Scan(&reports).Error; err == nil && len(reports) > 0 {
		repMap := make(map[string]*entity.ReviewAssignmentReport)
		for i := range reports {
			repMap[reports[i].ReviewAssignmentID] = &reports[i]
		}
		for i := range assignments {
			if rpt, ok := repMap[assignments[i].ID]; ok {
				assignments[i].Report = rpt
			}
		}
	}

	// Batch-load assignment files
	var asgnFiles []entity.ReviewFile
	if err := db(ctx).Raw(`
		SELECT * FROM review_files
		WHERE review_assignment_id IN (` + aidsClause + `)
		ORDER BY uploaded_at ASC
	`).Scan(&asgnFiles).Error; err == nil && len(asgnFiles) > 0 {
		fileMap := make(map[string][]entity.ReviewFile)
		for _, f := range asgnFiles {
			if f.ReviewAssignmentID != nil {
				fileMap[*f.ReviewAssignmentID] = append(fileMap[*f.ReviewAssignmentID], f)
			}
		}

		// Batch-load uploader users
		uploaderIDs := make(map[string]struct{})
		for _, f := range asgnFiles {
			if f.UploadedBy != "" {
				uploaderIDs[f.UploadedBy] = struct{}{}
			}
		}
		uploaderMap := make(map[string]*entity.User)
		for uid := range uploaderIDs {
			var u entity.User
			if err := db(ctx).Raw(`
				SELECT * FROM users
				WHERE id = ` + pqLit(uid) + `
				  AND deleted_at IS NULL
				LIMIT 1
			`).Scan(&u).Error; err == nil && u.ID != "" {
				uploaderMap[uid] = &u
			}
		}

		// Assign files + uploaders
		for i := range asgnFiles {
			if u, ok := uploaderMap[asgnFiles[i].UploadedBy]; ok {
				asgnFiles[i].Uploader = u
			}
		}
		for i := range assignments {
			if f, ok := fileMap[assignments[i].ID]; ok {
				assignments[i].Files = f
			}
		}
	}
}

func (r *Repository) UpdateRound(ctx context.Context, id string, updates map[string]any) error {
	updates["updated_at"] = time.Now()
	return db(ctx).Model(&entity.ReviewRound{}).
		Where("id = ?", id).Updates(updates).Error
}

// ====== Review Assignment ======

func (r *Repository) CreateAssignment(ctx context.Context, assignment *entity.ReviewAssignment) error {
	return db(ctx).Create(assignment).Error
}

func (r *Repository) GetAssignmentByID(ctx context.Context, id string) (*entity.ReviewAssignment, error) {
	uid, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, nil
	}
	idLit := uid.String()
	var assignment entity.ReviewAssignment
	err = db(ctx).Raw(
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
			_ = db(ctx).Raw(
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
	for _, tok := range candidates {
		esc := strings.ReplaceAll(tok, "'", "''")
		var assignment entity.ReviewAssignment
		err := db(ctx).Raw(
			`SELECT * FROM review_assignments WHERE invitation_token = '` + esc + `' LIMIT 1`,
		).Scan(&assignment).Error
		if err != nil {
			return nil, err
		}
		if assignment.ID == "" {
			continue
		}
		if err := r.HydrateAssignmentInvitationGraph(ctx, &assignment); err != nil {
			return nil, err
		}
		return &assignment, nil
	}
	return nil, nil
}

// HydrateAssignmentInvitationGraph loads round, manuscript, journal, and round creator using literal UUID
// queries only (PgBouncer-safe).
func (r *Repository) HydrateAssignmentInvitationGraph(ctx context.Context, a *entity.ReviewAssignment) error {
	if a == nil || strings.TrimSpace(a.ReviewRoundID) == "" {
		return nil
	}
	rid, err := uuid.Parse(strings.TrimSpace(a.ReviewRoundID))
	if err != nil {
		return nil
	}
	var round entity.ReviewRound
	if err := db(ctx).Raw(
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
		if err := db(ctx).Raw(
			`SELECT * FROM manuscripts WHERE id = '` + mid.String() + `'::uuid AND deleted_at IS NULL LIMIT 1`,
		).Scan(&ms).Error; err != nil {
			return err
		}
		if ms.ID != "" {
			var mFiles []entity.ManuscriptFile
			if err := db(ctx).Raw(
				`SELECT * FROM manuscript_files WHERE manuscript_id = '` + mid.String() + `'::uuid ORDER BY version ASC, uploaded_at ASC`,
			).Scan(&mFiles).Error; err != nil {
				return err
			}
			ms.Files = mFiles

			round.Manuscript = &ms
			if jid, err := uuid.Parse(strings.TrimSpace(ms.JournalID)); err == nil {
				var j entity.Journal
				if err := db(ctx).Raw(
					`SELECT * FROM journals WHERE id = '` + jid.String() + `'::uuid AND deleted_at IS NULL LIMIT 1`,
				).Scan(&j).Error; err != nil {
					return err
				}
				if j.ID != "" {
					ms.Journal = &j
				}
			}
			if ms.AssignedEditorID != nil && strings.TrimSpace(*ms.AssignedEditorID) != "" {
				if eid, err := uuid.Parse(strings.TrimSpace(*ms.AssignedEditorID)); err == nil {
					var ed entity.User
					if err := db(ctx).Raw(
						`SELECT * FROM users WHERE id = '` + eid.String() + `'::uuid AND deleted_at IS NULL LIMIT 1`,
					).Scan(&ed).Error; err != nil {
						return err
					}
					if ed.ID != "" {
						ms.AssignedEditor = &ed
					}
				}
			}
		}
	}
	if cid, err := uuid.Parse(strings.TrimSpace(round.CreatedBy)); err == nil {
		var creator entity.User
		if err := db(ctx).Raw(
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
	if err := db(ctx).Raw(`
		SELECT * FROM review_assignments
		WHERE review_round_id = ` + pqLit(roundID) + `
		ORDER BY created_at ASC
	`).Scan(&assignments).Error; err != nil {
		return nil, err
	}
	if len(assignments) > 0 {
		r.hydrateAssignments(ctx, assignments)
	}
	return assignments, nil
}

func (r *Repository) UpdateAssignment(ctx context.Context, id string, updates map[string]any) error {
	updates["updated_at"] = time.Now()
	return db(ctx).Model(&entity.ReviewAssignment{}).
		Where("id = ?", id).Updates(updates).Error
}

// ====== Review File ======

func (r *Repository) CreateReviewFile(ctx context.Context, file *entity.ReviewFile) error {
	return db(ctx).Create(file).Error
}

func (r *Repository) ListFilesByAssignment(ctx context.Context, assignmentID string) ([]entity.ReviewFile, error) {
	aid, err := uuid.Parse(strings.TrimSpace(assignmentID))
	if err != nil {
		return nil, nil
	}
	lit := aid.String()
	var files []entity.ReviewFile
	if err := db(ctx).Raw(
		`SELECT * FROM review_files WHERE review_assignment_id = '` + lit + `'::uuid ORDER BY uploaded_at ASC`,
	).Scan(&files).Error; err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return files, nil
	}
	seen := make(map[string]struct{})
	for _, f := range files {
		if f.UploadedBy != "" {
			seen[f.UploadedBy] = struct{}{}
		}
	}
	usersByID := make(map[string]entity.User)
	for uid := range seen {
		if uu, perr := uuid.Parse(uid); perr == nil {
			var u entity.User
			if err := db(ctx).Raw(
				`SELECT * FROM users WHERE id = '` + uu.String() + `'::uuid AND deleted_at IS NULL LIMIT 1`,
			).Scan(&u).Error; err != nil {
				return nil, err
			}
			if u.ID != "" {
				usersByID[uid] = u
			}
		}
	}
	for i := range files {
		if u, ok := usersByID[files[i].UploadedBy]; ok {
			uu := u
			files[i].Uploader = &uu
		}
	}
	return files, nil
}

func (r *Repository) ListFilesByRound(ctx context.Context, roundID string) ([]entity.ReviewFile, error) {
	var files []entity.ReviewFile
	if err := db(ctx).Raw(`
		SELECT * FROM review_files
		WHERE review_round_id = ` + pqLit(roundID) + `
		ORDER BY uploaded_at ASC
	`).Scan(&files).Error; err != nil {
		return nil, err
	}

	// Hydrate uploaders
	if len(files) > 0 {
		seen := make(map[string]struct{})
		for _, f := range files {
			if f.UploadedBy != "" {
				seen[f.UploadedBy] = struct{}{}
			}
		}
		usersByID := make(map[string]entity.User)
		for uid := range seen {
			if uu, perr := uuid.Parse(uid); perr == nil {
				var u entity.User
				if err := db(ctx).Raw(
					`SELECT * FROM users WHERE id = '` + uu.String() + `'::uuid AND deleted_at IS NULL LIMIT 1`,
				).Scan(&u).Error; err != nil {
					return nil, err
				}
				if u.ID != "" {
					usersByID[uid] = u
				}
			}
		}
		for i := range files {
			if u, ok := usersByID[files[i].UploadedBy]; ok {
				uu := u
				files[i].Uploader = &uu
			}
		}
	}

	return files, nil
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
	if err := db(ctx).Raw("SELECT COUNT(DISTINCT u.id) "+baseQuery, countArgs...).Scan(&total).Error; err != nil {
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

	if err := db(ctx).Raw(selectQuery, args...).Scan(&candidates).Error; err != nil {
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
	if err := db(ctx).Raw("SELECT COUNT(DISTINCT u.id) "+baseQuery, countArgs...).Scan(&total).Error; err != nil {
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

	if err := db(ctx).Raw(selectQuery, args...).Scan(&candidates).Error; err != nil {
		return nil, 0, err
	}

	return candidates, total, nil
}

// CountPendingInvitationByRoundEmail counts INVITED rows for the same round and email.
func (r *Repository) CountPendingInvitationByRoundEmail(ctx context.Context, roundID, emailNorm string) (int64, error) {
	var n int64
	err := db(ctx).Raw(`
		SELECT COUNT(*) FROM review_assignments
		WHERE review_round_id = ?
		  AND status = ?
		  AND LOWER(TRIM(COALESCE(invited_email, ''))) = ?`,
		roundID, constant.ReviewAssignmentStatusInvited, emailNorm,
	).Scan(&n).Error
	return n, err
}

// CountActiveAssignmentByRoundReviewer counts INVITED or ACCEPTED rows for the same round and reviewer user.
func (r *Repository) CountActiveAssignmentByRoundReviewer(ctx context.Context, roundID, reviewerUserID string) (int64, error) {
	var n int64
	err := db(ctx).Raw(`
		SELECT COUNT(*) FROM review_assignments
		WHERE review_round_id = ?
		  AND reviewer_id = ?
		  AND status IN (?, ?)`,
		roundID,
		reviewerUserID,
		constant.ReviewAssignmentStatusInvited,
		constant.ReviewAssignmentStatusAccepted,
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
	if err := db(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
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
	if err := db(ctx).Raw(selectSQL, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
