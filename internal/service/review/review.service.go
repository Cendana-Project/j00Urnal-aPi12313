package review

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/email"
	"github.com/api-monolith-template/internal/infrastructure"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
	reviewRepo "github.com/api-monolith-template/internal/repository/review"
	roleRepo "github.com/api-monolith-template/internal/repository/role"
	userRepo "github.com/api-monolith-template/internal/repository/user"
	manuscriptSvc "github.com/api-monolith-template/internal/service/manuscript"
	storageSvc "github.com/api-monolith-template/internal/service/storage"
	"github.com/api-monolith-template/internal/util"
	"github.com/api-monolith-template/pkg/pagination"
)

const invitationExpiryDays = 7

type Service struct {
	reviewRepo  *reviewRepo.Repository
	manuscripts *manuscriptSvc.Service
	storage     *storageSvc.Service
	emailSender email.Sender
	userRepo    *userRepo.Repository
	roleRepo    *roleRepo.Repository
}

func NewService(
	rr *reviewRepo.Repository,
	ms *manuscriptSvc.Service,
	ss *storageSvc.Service,
	es email.Sender,
	ur *userRepo.Repository,
	rlr *roleRepo.Repository,
) *Service {
	return &Service{
		reviewRepo:  rr,
		manuscripts: ms,
		storage:     ss,
		emailSender: es,
		userRepo:    ur,
		roleRepo:    rlr,
	}
}

// ====================================================================
// Chief Editor Methods
// ====================================================================

// ListSubmissions lists manuscripts by statuses (for chief editor view).
func (s *Service) ListSubmissions(ctx context.Context, statuses []constant.ManuscriptStatus, pg *pagination.Pagination) ([]entity.Manuscript, int64, error) {
	return s.manuscripts.ListByStatuses(ctx, statuses, pg)
}

// ListEditorCandidates returns users with EDITOR role.
func (s *Service) ListEditorCandidates(ctx context.Context, search string, pg *pagination.Pagination) ([]reviewRepo.EditorCandidate, int64, error) {
	return s.reviewRepo.ListEditorCandidates(ctx, search, pg)
}

// AssignEditor assigns an editor to a manuscript (Chief Editor action).
func (s *Service) AssignEditor(ctx context.Context, manuscriptID, editorID, chiefEditorID string) error {
	// 1. Validate manuscript exists and is in correct status
	m, err := s.manuscripts.GetByID(ctx, manuscriptID)
	if err != nil {
		return err
	}
	if m == nil {
		return constant.ErrRecordNotFound
	}
	if m.Status != constant.ManuscriptStatusSubmitted {
		return constant.ErrInvalidManuscriptStatus
	}

	// 2. Validate editor has EDITOR role
	hasRole, err := s.roleRepo.UserHasRole(editorID, constant.RoleEditor)
	if err != nil {
		return err
	}
	if !hasRole {
		return constant.ErrRecordNotFound
	}

	// 3. Assign editor (updates manuscript status to ASSIGNED_TO_EDITOR)
	if err := s.manuscripts.AssignEditor(ctx, manuscriptID, editorID); err != nil {
		return err
	}

	// 4. Send notification email to editor (async, don't fail on email error)
	go func() {
		editor, err := s.userRepo.GetByID(editorID)
		if err != nil || editor == nil {
			return
		}
		chiefEditor, err := s.userRepo.GetByID(chiefEditorID)
		if err != nil || chiefEditor == nil {
			return
		}
		editorName := formatUserName(editor)
		chiefName := formatUserName(chiefEditor)
		body := email.RenderEditorAssigned(editorName, m.Title, chiefName)
		if s.emailSender != nil {
			_ = s.emailSender.Send(editor.Email, "Anda Ditugaskan sebagai Editor - "+m.Title, body)
		}
	}()

	return nil
}

// ====================================================================
// Editor Methods
// ====================================================================

// ListEditorSubmissions returns manuscripts assigned to a specific editor.
func (s *Service) ListEditorSubmissions(ctx context.Context, editorID string, tab string, pg *pagination.Pagination) ([]entity.Manuscript, int64, error) {
	var statuses []constant.ManuscriptStatus

	switch tab {
	case "my_queue":
		statuses = []constant.ManuscriptStatus{
			constant.ManuscriptStatusAssignedToEditor,
			constant.ManuscriptStatusUnderReview,
			constant.ManuscriptStatusRevisionRequired,
			constant.ManuscriptStatusRevised,
		}
	case "archives":
		statuses = []constant.ManuscriptStatus{
			constant.ManuscriptStatusAccepted,
			constant.ManuscriptStatusRejected,
			constant.ManuscriptStatusPublished,
		}
	default:
		// Default to queue if unknown
		statuses = []constant.ManuscriptStatus{
			constant.ManuscriptStatusAssignedToEditor,
			constant.ManuscriptStatusUnderReview,
			constant.ManuscriptStatusRevisionRequired,
			constant.ManuscriptStatusRevised,
		}
	}

	return s.manuscripts.ListByAssignedEditor(ctx, editorID, statuses, pg)
}

// SendToReview creates a review round and updates manuscript status to UNDER_REVIEW.
func (s *Service) SendToReview(ctx context.Context, manuscriptID, editorID string) (*entity.ReviewRound, error) {
	// 1. Validate manuscript and editor assignment
	m, err := s.manuscripts.GetByID(ctx, manuscriptID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, constant.ErrRecordNotFound
	}
	if err := s.validateEditorAssignment(m, editorID); err != nil {
		return nil, err
	}

	// Validate status
	validStatuses := map[constant.ManuscriptStatus]bool{
		constant.ManuscriptStatusAssignedToEditor: true,
		constant.ManuscriptStatusRevised:          true,
	}
	if !validStatuses[m.Status] {
		return nil, constant.ErrInvalidManuscriptStatus
	}

	// 2. Determine round number
	latest, err := s.reviewRepo.GetLatestRoundByManuscript(ctx, manuscriptID)
	if err != nil {
		return nil, err
	}
	roundNumber := 1
	if latest != nil {
		roundNumber = latest.RoundNumber + 1
	}

	// 3. Create round
	round := &entity.ReviewRound{
		ManuscriptID: manuscriptID,
		RoundNumber:  roundNumber,
		Status:       constant.ReviewRoundStatusPending,
		CreatedBy:    editorID,
		CreatedAt:    time.Now(),
	}
	if err := s.reviewRepo.CreateRound(ctx, round); err != nil {
		return nil, err
	}

	// 4. Update manuscript status
	if err := s.manuscripts.UpdateStatus(ctx, manuscriptID, constant.ManuscriptStatusUnderReview); err != nil {
		return nil, err
	}

	return s.reviewRepo.GetRoundByID(ctx, round.ID)
}

// AcceptManuscript accepts the manuscript (editor decision).
func (s *Service) AcceptManuscript(ctx context.Context, manuscriptID, editorID string) error {
	m, err := s.manuscripts.GetByID(ctx, manuscriptID)
	if err != nil {
		return err
	}
	if m == nil {
		return constant.ErrRecordNotFound
	}
	if err := s.validateEditorAssignment(m, editorID); err != nil {
		return err
	}

	if err := s.manuscripts.UpdateStatus(ctx, manuscriptID, constant.ManuscriptStatusAccepted); err != nil {
		return err
	}

	// Notify author
	go s.notifyAuthorDecision(m, "ACCEPTED", "")

	return nil
}

// DeclineManuscript rejects the manuscript with a reason.
func (s *Service) DeclineManuscript(ctx context.Context, manuscriptID, editorID, reason string) error {
	m, err := s.manuscripts.GetByID(ctx, manuscriptID)
	if err != nil {
		return err
	}
	if m == nil {
		return constant.ErrRecordNotFound
	}
	if err := s.validateEditorAssignment(m, editorID); err != nil {
		return err
	}

	if err := s.manuscripts.UpdateStatus(ctx, manuscriptID, constant.ManuscriptStatusRejected); err != nil {
		return err
	}

	// Notify author
	go s.notifyAuthorDecision(m, "REJECTED", reason)

	return nil
}

// RequestRevision requests revisions from the author.
func (s *Service) RequestRevision(ctx context.Context, manuscriptID, editorID, comments string) error {
	m, err := s.manuscripts.GetByID(ctx, manuscriptID)
	if err != nil {
		return err
	}
	if m == nil {
		return constant.ErrRecordNotFound
	}
	if err := s.validateEditorAssignment(m, editorID); err != nil {
		return err
	}

	if err := s.manuscripts.UpdateStatus(ctx, manuscriptID, constant.ManuscriptStatusRevisionRequired); err != nil {
		return err
	}

	// Notify author
	go s.notifyAuthorDecision(m, "REVISION_REQUIRED", comments)

	return nil
}

// GetReviewDetails returns all review rounds and assignments for a manuscript.
func (s *Service) GetReviewDetails(ctx context.Context, manuscriptID string) (*entity.Manuscript, []entity.ReviewRound, error) {
	m, err := s.manuscripts.GetByID(ctx, manuscriptID)
	if err != nil {
		return nil, nil, err
	}
	if m == nil {
		return nil, nil, constant.ErrRecordNotFound
	}

	rounds, err := s.reviewRepo.ListRoundsByManuscript(ctx, manuscriptID)
	if err != nil {
		return nil, nil, err
	}

	return m, rounds, nil
}

// ListReviewerCandidates returns users with REVIEWER role.
func (s *Service) ListReviewerCandidates(ctx context.Context, search string, pg *pagination.Pagination) ([]reviewRepo.ReviewerCandidate, int64, error) {
	return s.reviewRepo.ListReviewerCandidates(ctx, search, pg)
}

// ====================================================================
// Reviewer Invitation
// ====================================================================

// InviteReviewer invites a prospective reviewer by email for the manuscript's current review round
// (latest round with status PENDING or IN_REVIEW). Invitation and review due date both default to 7 days from now.
func (s *Service) InviteReviewer(ctx context.Context, manuscriptID, inviteEmail, editorID string) (*entity.ReviewAssignment, error) {
	m, err := s.manuscripts.GetByID(ctx, manuscriptID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, constant.ErrRecordNotFound
	}
	if err := s.validateEditorAssignment(m, editorID); err != nil {
		return nil, err
	}

	latest, err := s.reviewRepo.GetLatestRoundByManuscript(ctx, manuscriptID)
	if err != nil {
		return nil, err
	}
	if latest == nil {
		return nil, constant.ErrNoActiveReviewRound
	}
	if latest.Status != constant.ReviewRoundStatusPending && latest.Status != constant.ReviewRoundStatusInReview {
		return nil, constant.ErrInvalidManuscriptStatus
	}

	round, err := s.reviewRepo.GetRoundByID(ctx, latest.ID)
	if err != nil {
		return nil, err
	}
	if round == nil {
		return nil, constant.ErrRecordNotFound
	}

	emailNorm := strings.ToLower(strings.TrimSpace(inviteEmail))
	if emailNorm == "" {
		return nil, constant.ErrValidationFailed
	}

	dup, err := s.reviewRepo.CountPendingInvitationByRoundEmail(ctx, round.ID, emailNorm)
	if err != nil {
		return nil, err
	}
	if dup > 0 {
		return nil, constant.ErrReviewerInviteDuplicate
	}

	token := util.GenerateSecureToken(32)
	now := time.Now()
	dueDate := now.Add(invitationExpiryDays * 24 * time.Hour)

	assignment := &entity.ReviewAssignment{
		ReviewRoundID:       round.ID,
		ReviewerID:          nil,
		InvitedEmail:        emailNorm,
		AssignedBy:          editorID,
		Status:              constant.ReviewAssignmentStatusInvited,
		InvitationToken:     &token,
		InvitationExpiresAt: dueDate,
		DueDate:             dueDate,
		CreatedAt:           now,
	}

	if err := s.reviewRepo.CreateAssignment(ctx, assignment); err != nil {
		return nil, err
	}

	if round.Status == constant.ReviewRoundStatusPending {
		_ = s.reviewRepo.UpdateRound(ctx, round.ID, map[string]any{
			"status": constant.ReviewRoundStatusInReview,
		})
	}

	go s.sendReviewerInviteEmail(round, editorID, emailNorm, token, dueDate)

	return s.reviewRepo.GetAssignmentByID(ctx, assignment.ID)
}

func (s *Service) sendReviewerInviteEmail(round *entity.ReviewRound, editorID, inviteEmail, token string, dueDate time.Time) {
	ctx := context.Background()
	editor, err := s.userRepo.GetByID(editorID)
	if err != nil {
		util.Errorf(ctx, "reviewer invite email: load editor %s: %v", editorID, err)
		return
	}
	if editor == nil {
		util.Errorf(ctx, "reviewer invite email: editor %s not found", editorID)
		return
	}
	editorName := formatUserName(editor)
	dueDateStr := dueDate.Format("2 January 2006")

	manuscriptTitle := ""
	if round.Manuscript != nil {
		manuscriptTitle = round.Manuscript.Title
	} else {
		m, _ := s.manuscripts.GetByID(context.Background(), round.ManuscriptID)
		if m != nil {
			manuscriptTitle = m.Title
		}
	}

	reviewerName := reviewerDisplayNameFromEmail(inviteEmail)

	baseURL := config.Env.Server.FrontendURL
	if baseURL == "" {
		baseURL = config.Env.Server.BaseURL
	}
	acceptURL := fmt.Sprintf("%s/reviewer/invitation/%s/accept", baseURL, token)
	declineURL := fmt.Sprintf("%s/reviewer/invitation/%s/decline", baseURL, token)

	body := email.RenderReviewerInvitation(reviewerName, manuscriptTitle, editorName, dueDateStr, acceptURL, declineURL)
	plain := email.RenderReviewerInvitationPlain(reviewerName, manuscriptTitle, editorName, dueDateStr, acceptURL, declineURL)
	subject := email.ReviewerInviteSubject(manuscriptTitle)
	if s.emailSender == nil {
		util.Errorf(ctx, "reviewer invite email: sender disabled (EMAIL_ENABLED=false or not configured)")
		return
	}
	if err := s.emailSender.SendWithContextAndText(ctx, inviteEmail, subject, body, plain); err != nil {
		util.Errorf(ctx, "reviewer invite email: send to %s: %v", maskInvitationEmail(inviteEmail), err)
	}
}

func reviewerDisplayNameFromEmail(addr string) string {
	parts := strings.Split(addr, "@")
	if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
		local := strings.TrimSpace(parts[0])
		runes := []rune(local)
		if len(runes) > 0 && runes[0] >= 'a' && runes[0] <= 'z' {
			runes[0] = runes[0] - 'a' + 'A'
		}
		return string(runes)
	}
	return "Reviewer"
}

func maskInvitationEmail(addr string) string {
	addr = strings.TrimSpace(addr)
	at := strings.LastIndex(addr, "@")
	if at <= 1 || at >= len(addr)-1 {
		return "***"
	}
	local, domain := addr[:at], addr[at+1:]
	if len(local) < 2 {
		return "***@" + domain
	}
	return local[:2] + "***@" + domain
}

func (s *Service) assertInvitationOpen(a *entity.ReviewAssignment) error {
	if time.Now().After(a.InvitationExpiresAt) {
		return constant.ErrInvitationExpired
	}
	switch a.Status {
	case constant.ReviewAssignmentStatusInvited:
		return nil
	case constant.ReviewAssignmentStatusAccepted:
		return constant.ErrInvitationAlreadyAccepted
	default:
		return constant.ErrInvitationNotFound
	}
}

// GetInvitationPreview returns safe metadata for the invitation landing page.
func (s *Service) GetInvitationPreview(ctx context.Context, token string) (*response.ReviewerInvitationPreviewResponse, error) {
	a, err := s.reviewRepo.GetAssignmentByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, constant.ErrInvitationNotFound
	}
	if err := s.assertInvitationOpen(a); err != nil {
		return nil, err
	}

	out := &response.ReviewerInvitationPreviewResponse{
		Valid:               true,
		InvitationExpiresAt: a.InvitationExpiresAt,
		DueDate:             a.DueDate,
		InvitedEmailMasked:  maskInvitationEmail(a.InvitedEmail),
		RequiresPassword:    true,
		Section:             "Article Text",
	}
	if a.ReviewRound != nil && a.ReviewRound.Manuscript != nil {
		out.ManuscriptTitle = a.ReviewRound.Manuscript.Title
		if a.ReviewRound.Manuscript.Journal != nil {
			out.JournalName = a.ReviewRound.Manuscript.Journal.Name
		}
	}
	if a.ReviewRound != nil && a.ReviewRound.Creator != nil {
		out.EditorName = formatUserName(a.ReviewRound.Creator)
	}
	return out, nil
}

// CompleteReviewerInvitation validates the token, sets password, assigns REVIEWER role, and marks the assignment accepted.
func (s *Service) CompleteReviewerInvitation(ctx context.Context, token, password, firstName, lastName string) error {
	if !util.IsValidPassword(password) {
		return constant.ErrInvalidPassword
	}

	a, err := s.reviewRepo.GetAssignmentByToken(ctx, token)
	if err != nil {
		return err
	}
	if a == nil {
		return constant.ErrInvitationNotFound
	}
	if err := s.assertInvitationOpen(a); err != nil {
		return err
	}

	invitedEmail := strings.ToLower(strings.TrimSpace(a.InvitedEmail))
	if invitedEmail == "" {
		return constant.ErrInvitationNotFound
	}

	if util.IsPasswordSimilarToUserInfo(password, invitedEmail, invitedEmail) {
		return constant.ErrPasswordSimilarToUserInfo
	}

	existing, err := s.userRepo.FindByEmail(invitedEmail)
	if err != nil {
		return constant.ErrInternalServerError
	}
	if existing != nil {
		isSA, err := s.roleRepo.IsUserSuperAdmin(ctx, existing.ID)
		if err != nil {
			return constant.ErrInternalServerError
		}
		if isSA {
			return constant.ErrForbidden
		}
	}

	username, err := s.pickUsernameForInvitation(ctx, invitedEmail, existing)
	if err != nil {
		return err
	}
	if util.IsPasswordSimilarToUserInfo(password, username, invitedEmail) {
		return constant.ErrPasswordSimilarToUserInfo
	}

	hash, err := util.HashPasswordScrypt(password)
	if err != nil {
		return constant.ErrInternalServerError
	}

	roleID, err := s.roleRepo.GetRoleIDBySlug(ctx, constant.RoleReviewer)
	if err != nil {
		return constant.ErrInternalServerError
	}

	now := time.Now()
	return infrastructure.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userID string

		if existing == nil {
			fn := strings.TrimSpace(firstName)
			ln := strings.TrimSpace(lastName)
			var fnPtr, lnPtr *string
			if fn != "" {
				fnPtr = &fn
			}
			if ln != "" {
				lnPtr = &ln
			}
			u := entity.User{
				ID:           uuid.New().String(),
				Email:        invitedEmail,
				Username:     username,
				PasswordHash: hash,
				FirstName:    fnPtr,
				LastName:     lnPtr,
				Status:       "active",
				VerifiedAt:   &now,
			}
			if err := tx.Create(&u).Error; err != nil {
				return err
			}
			userID = u.ID
		} else {
			userID = existing.ID
			updates := map[string]any{
				"password_hash": hash,
				"updated_at":    now,
			}
			if strings.TrimSpace(firstName) != "" {
				updates["first_name"] = strings.TrimSpace(firstName)
			}
			if strings.TrimSpace(lastName) != "" {
				updates["last_name"] = strings.TrimSpace(lastName)
			}
			if strings.ToLower(existing.Status) != "active" {
				updates["status"] = "active"
				updates["verified_at"] = now
			}
			if err := tx.Model(&entity.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
				return err
			}
		}

		if err := tx.Exec(`
			INSERT INTO user_roles (user_id, role_id, assigned_at)
			VALUES (?, ?, NOW())
			ON CONFLICT (user_id, role_id) DO NOTHING
		`, userID, roleID).Error; err != nil {
			return err
		}

		accepted := now
		up := map[string]any{
			"reviewer_id":            userID,
			"status":                 constant.ReviewAssignmentStatusAccepted,
			"invitation_accepted_at": accepted,
			"invitation_token":       gorm.Expr("NULL"),
			"updated_at":             now,
		}
		if err := tx.Model(&entity.ReviewAssignment{}).Where("id = ?", a.ID).Updates(up).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *Service) pickUsernameForInvitation(ctx context.Context, invitedEmail string, existing *entity.User) (string, error) {
	if existing != nil {
		return existing.Username, nil
	}
	local := strings.ToLower(strings.TrimSpace(strings.Split(invitedEmail, "@")[0]))
	var b strings.Builder
	for _, r := range local {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.', r == '_', r == '-':
			b.WriteRune('_')
		default:
			b.WriteRune('_')
		}
	}
	base := strings.Trim(b.String(), "_")
	if len(base) < 3 {
		base = "reviewer"
	}
	if len(base) > 40 {
		base = base[:40]
	}
	for i := 0; i < 1000; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s_%d", base, i)
		}
		ok, err := s.userRepo.ExistsUsername(candidate)
		if err != nil {
			return "", err
		}
		if !ok {
			return candidate, nil
		}
	}
	return "", constant.ErrInternalServerError
}

// DeclineReviewerInvitation marks the assignment declined (token must still be valid).
func (s *Service) DeclineReviewerInvitation(ctx context.Context, token, reason string) error {
	a, err := s.reviewRepo.GetAssignmentByToken(ctx, token)
	if err != nil {
		return err
	}
	if a == nil {
		return constant.ErrInvitationNotFound
	}
	if err := s.assertInvitationOpen(a); err != nil {
		return err
	}
	now := time.Now()
	updates := map[string]any{
		"status":           constant.ReviewAssignmentStatusDeclined,
		"comments":         strings.TrimSpace(reason),
		"invitation_token": gorm.Expr("NULL"),
		"updated_at":       now,
	}
	return s.reviewRepo.UpdateAssignment(ctx, a.ID, updates)
}

// ListReviewerHistory returns completed assignments for the reviewer portal.
func (s *Service) ListReviewerHistory(ctx context.Context, reviewerID, search, recommendationEq, editorDecisionEq string, pg *pagination.Pagination) ([]reviewRepo.ReviewerHistoryRow, int64, error) {
	return s.reviewRepo.ListReviewerHistory(ctx, reviewerID, search, recommendationEq, editorDecisionEq, pg)
}

// SubmitReviewerReview records the reviewer's recommendation and marks the assignment completed (shows in History).
func (s *Service) SubmitReviewerReview(ctx context.Context, reviewerID, assignmentID, recommendation, comments string) error {
	a, err := s.reviewRepo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}
	if a == nil {
		return constant.ErrRecordNotFound
	}
	if a.ReviewerID == nil || *a.ReviewerID != reviewerID {
		return constant.ErrForbidden
	}
	if a.Status == constant.ReviewAssignmentStatusCompleted {
		return constant.ErrReviewAlreadyCompleted
	}
	if a.Status != constant.ReviewAssignmentStatusAccepted {
		return constant.ErrReviewerAssignmentNotAllowed
	}

	rec := strings.TrimSpace(recommendation)
	switch rec {
	case "ACCEPT", "REJECT", "MAJOR_REVISION", "MINOR_REVISION":
	default:
		return constant.ErrValidationFailed
	}

	now := time.Now()
	var cmt *string
	if t := strings.TrimSpace(comments); t != "" {
		cmt = &t
	}
	return s.reviewRepo.UpdateAssignment(ctx, a.ID, map[string]any{
		"recommendation": rec,
		"comments":       cmt,
		"status":         constant.ReviewAssignmentStatusCompleted,
		"completed_at":   now,
	})
}

// ====================================================================
// Round Decision
// ====================================================================

// MakeRoundDecision makes a decision for a review round.
func (s *Service) MakeRoundDecision(ctx context.Context, roundID, editorID, decision, comments string) error {
	round, err := s.reviewRepo.GetRoundByID(ctx, roundID)
	if err != nil {
		return err
	}
	if round == nil {
		return constant.ErrRecordNotFound
	}

	// Validate editor assignment on manuscript
	m, err := s.manuscripts.GetByID(ctx, round.ManuscriptID)
	if err != nil {
		return err
	}
	if m == nil {
		return constant.ErrRecordNotFound
	}
	if err := s.validateEditorAssignment(m, editorID); err != nil {
		return err
	}

	// Update round
	now := time.Now()
	if err := s.reviewRepo.UpdateRound(ctx, roundID, map[string]any{
		"status":            constant.ReviewRoundStatusCompleted,
		"editor_decision":   decision,
		"decision_comments": comments,
		"decision_at":       now,
	}); err != nil {
		return err
	}

	// Update manuscript status based on decision
	var newStatus constant.ManuscriptStatus
	switch decision {
	case "ACCEPT":
		newStatus = constant.ManuscriptStatusAccepted
	case "REJECT":
		newStatus = constant.ManuscriptStatusRejected
	case "REVISION_REQUIRED":
		newStatus = constant.ManuscriptStatusRevisionRequired
	default:
		return constant.ErrValidationError
	}

	if err := s.manuscripts.UpdateStatus(ctx, round.ManuscriptID, newStatus); err != nil {
		return err
	}

	// Notify author
	go s.notifyAuthorDecision(m, decision, comments)

	return nil
}

// ====================================================================
// Helpers (private)
// ====================================================================

func (s *Service) validateEditorAssignment(m *entity.Manuscript, editorID string) error {
	if m.AssignedEditorID == nil || *m.AssignedEditorID != editorID {
		return constant.ErrEditorNotAssigned
	}
	return nil
}

func (s *Service) notifyAuthorDecision(m *entity.Manuscript, decision, comments string) {
	if m.MainAuthor == nil {
		return
	}
	authorName := formatUserName(m.MainAuthor)
	body := email.RenderSubmissionDecision(authorName, m.Title, decision, comments)
	if s.emailSender != nil {
		_ = s.emailSender.Send(m.MainAuthor.Email, "Keputusan Editorial - "+m.Title, body)
	}
}

// UploadReviewFileToStorage uploads file bytes to Supabase storage (DRY: reuses storage.Service).
func (s *Service) UploadReviewFileToStorage(ctx context.Context, fileBytes []byte, path, contentType string) (string, error) {
	return s.storage.Upload(ctx, fileBytes, path, contentType)
}

func formatUserName(u *entity.User) string {
	firstName := ""
	if u.FirstName != nil {
		firstName = *u.FirstName
	}
	lastName := ""
	if u.LastName != nil {
		lastName = *u.LastName
	}
	name := strings.TrimSpace(fmt.Sprintf("%s %s", firstName, lastName))
	if name == "" {
		return u.Email
	}
	return name
}
