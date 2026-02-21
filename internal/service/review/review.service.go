package review

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/email"
	"github.com/api-monolith-template/internal/model/entity"
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
	if m.Status != constant.ManuscriptStatusSubmitted && m.Status != constant.ManuscriptStatusUnderChiefReview {
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
		statuses = []constant.ManuscriptStatus{constant.ManuscriptStatusAssignedToEditor}
	case "in_review":
		statuses = []constant.ManuscriptStatus{constant.ManuscriptStatusUnderReview}
	case "revision":
		statuses = []constant.ManuscriptStatus{constant.ManuscriptStatusRevisionRequired, constant.ManuscriptStatusRevised}
	case "archives":
		statuses = []constant.ManuscriptStatus{constant.ManuscriptStatusAccepted, constant.ManuscriptStatusRejected}
	default:
		// all active
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

// InviteReviewer invites a reviewer to a review round.
func (s *Service) InviteReviewer(ctx context.Context, roundID, reviewerID, editorID string, dueDate time.Time) (*entity.ReviewAssignment, error) {
	// 1. Validate round
	round, err := s.reviewRepo.GetRoundByID(ctx, roundID)
	if err != nil {
		return nil, err
	}
	if round == nil {
		return nil, constant.ErrRecordNotFound
	}

	// 2. Validate reviewer has REVIEWER role
	hasRole, err := s.roleRepo.UserHasRole(reviewerID, constant.RoleReviewer)
	if err != nil {
		return nil, err
	}
	if !hasRole {
		return nil, constant.ErrRecordNotFound
	}

	// 3. Generate invitation token
	token := util.GenerateSecureToken(32)

	// 4. Create assignment
	now := time.Now()
	expiresAt := now.Add(invitationExpiryDays * 24 * time.Hour)

	assignment := &entity.ReviewAssignment{
		ReviewRoundID:       roundID,
		ReviewerID:          reviewerID,
		AssignedBy:          editorID,
		Status:              constant.ReviewAssignmentStatusInvited,
		InvitationToken:     &token,
		InvitationExpiresAt: expiresAt,
		DueDate:             dueDate,
		CreatedAt:           now,
	}

	if err := s.reviewRepo.CreateAssignment(ctx, assignment); err != nil {
		return nil, err
	}

	// 5. Update round status to IN_REVIEW if still PENDING
	if round.Status == constant.ReviewRoundStatusPending {
		_ = s.reviewRepo.UpdateRound(ctx, roundID, map[string]any{
			"status": constant.ReviewRoundStatusInReview,
		})
	}

	// 6. Send invitation email
	go func() {
		reviewer, err := s.userRepo.GetByID(reviewerID)
		if err != nil || reviewer == nil {
			return
		}
		editor, err := s.userRepo.GetByID(editorID)
		if err != nil || editor == nil {
			return
		}

		reviewerName := formatUserName(reviewer)
		editorName := formatUserName(editor)
		dueDateStr := dueDate.Format("2 January 2006")

		// Get manuscript title from round
		manuscriptTitle := ""
		if round.Manuscript != nil {
			manuscriptTitle = round.Manuscript.Title
		} else {
			m, _ := s.manuscripts.GetByID(context.Background(), round.ManuscriptID)
			if m != nil {
				manuscriptTitle = m.Title
			}
		}

		baseURL := config.Env.Server.FrontendURL
		if baseURL == "" {
			baseURL = config.Env.Server.BaseURL
		}
		acceptURL := fmt.Sprintf("%s/reviewer/invitation/%s/accept", baseURL, token)
		declineURL := fmt.Sprintf("%s/reviewer/invitation/%s/decline", baseURL, token)

		body := email.RenderReviewerInvitation(reviewerName, manuscriptTitle, editorName, dueDateStr, acceptURL, declineURL)
		if s.emailSender != nil {
			_ = s.emailSender.Send(reviewer.Email, "Undangan Review Manuskrip - "+manuscriptTitle, body)
		}
	}()

	return s.reviewRepo.GetAssignmentByID(ctx, assignment.ID)
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
