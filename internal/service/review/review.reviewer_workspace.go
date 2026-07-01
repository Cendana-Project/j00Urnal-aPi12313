package review

import (
	"context"
	"io"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/infrastructure"
	"github.com/api-monolith-template/internal/mapper"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/util"
	"github.com/api-monolith-template/pkg/pagination"
)

// buildReviewerWorkflow maps manuscript + round + this assignment to the five-step reviewer portal stepper.
func buildReviewerWorkflow(m *entity.Manuscript, round *entity.ReviewRound, a *entity.ReviewAssignment) ([]response.ReviewerWorkflowStep, int) {
	if m == nil || round == nil || a == nil {
		return nil, 0
	}
	rs := round.Status
	as := a.Status
	ms := m.Status

	initialDone := ms != constant.ManuscriptStatusSubmitted
	editorialDone := m.AssignedEditorID != nil && strings.TrimSpace(*m.AssignedEditorID) != ""
	indepDone := as == constant.ReviewAssignmentStatusCompleted
	finalValDone := rs == constant.ReviewRoundStatusRevisionRequested || rs == constant.ReviewRoundStatusCompleted
	finalDecDone := rs == constant.ReviewRoundStatusCompleted

	steps := []response.ReviewerWorkflowStep{
		{ID: "initial_validation", Title: "Initial validation", Done: initialDone},
		{ID: "editorial_assignment", Title: "Editorial assignment", Done: editorialDone},
		{ID: "independent_review", Title: "Independent review", Done: indepDone},
		{ID: "final_validation", Title: "Final validation", Done: finalValDone},
		{ID: "final_decision", Title: "Final decision", Done: finalDecDone},
	}

	currentIdx := 0
	switch {
	case as == constant.ReviewAssignmentStatusAccepted:
		currentIdx = 2
	case indepDone && !finalDecDone:
		currentIdx = 3
	case finalDecDone:
		currentIdx = 4
	default:
		for i := range steps {
			if !steps[i].Done {
				currentIdx = i
				break
			}
		}
	}
	for i := range steps {
		steps[i].Current = (i == currentIdx)
	}
	return steps, currentIdx
}

// ListReviewerAssignments returns the reviewer's assignments filtered by status (e.g. ACCEPTED or COMPLETED).
func (s *Service) ListReviewerAssignments(ctx context.Context, reviewerID string, status constant.ReviewAssignmentStatus, pg *pagination.Pagination) ([]response.ReviewerAssignmentListItemResponse, int64, error) {
	rows, total, err := s.reviewRepo.ListReviewerAssignments(ctx, reviewerID, status, pg)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToReviewerAssignmentListResponse(rows), total, nil
}

// GetReviewerAssignmentWorkspace returns one assignment with round + manuscript context, report, files, and workflow stepper.
// Form schema JSON: GET /v1/reviewer/review-form-schema (not duplicated here).
func (s *Service) GetReviewerAssignmentWorkspace(ctx context.Context, reviewerID, assignmentID string) (*response.ReviewerWorkspaceResponse, error) {
	a, err := s.reviewRepo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, constant.ErrRecordNotFound
	}
	if err := assertReviewerOwnsAssignment(a, reviewerID); err != nil {
		return nil, err
	}
	switch a.Status {
	case constant.ReviewAssignmentStatusAccepted, constant.ReviewAssignmentStatusCompleted, constant.ReviewAssignmentStatusWithdrawn:
	default:
		return nil, constant.ErrReviewerAssignmentNotAllowed
	}
	if err := s.reviewRepo.HydrateAssignmentInvitationGraph(ctx, a); err != nil {
		return nil, err
	}
	var m *entity.Manuscript
	if a.ReviewRound != nil {
		m = a.ReviewRound.Manuscript
	}
	var authors []entity.ManuscriptAuthor
	if m != nil {
		authors, err = s.reviewRepo.LoadManuscriptAuthors(ctx, m.ID)
		if err != nil {
			return nil, err
		}
	}
	files, err := s.reviewRepo.ListFilesByAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	rep, err := s.reviewRepo.GetReportByAssignmentID(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	isDelayed := a.Status == constant.ReviewAssignmentStatusAccepted && now.After(a.DueDate.UTC())
	workflow, cur := buildReviewerWorkflow(m, a.ReviewRound, a)
	out := mapper.ToReviewerWorkspaceResponse(a, m, authors, files, rep, isDelayed, workflow, cur)
	return &out, nil
}

// PatchReviewerReportDraft merges answers into the draft row and validates partial answers against the form schema.
func (s *Service) PatchReviewerReportDraft(ctx context.Context, reviewerID, assignmentID string, body request.PatchReviewerReportDraftRequest) error {
	if s.reviewerForm == nil {
		return constant.ErrInternalServerError
	}
	a, err := s.reviewRepo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}
	if a == nil {
		return constant.ErrRecordNotFound
	}
	if err := assertReviewerOwnsAssignment(a, reviewerID); err != nil {
		return err
	}
	if a.Status != constant.ReviewAssignmentStatusAccepted {
		return constant.ErrReviewerAssignmentNotAllowed
	}
	existing, err := s.reviewRepo.GetReportByAssignmentID(ctx, assignmentID)
	if err != nil {
		return err
	}
	if existing != nil && existing.Status == constant.ReviewReportStatusSubmitted {
		return constant.ErrReviewAlreadyCompleted
	}
	defaultSV := s.reviewerForm.SchemaVersion
	var merged reviewerStoredReport
	var mergeErr error
	if existing != nil && len(existing.Payload) > 0 {
		merged, mergeErr = mergeReviewerPayload(existing.Payload, body.SchemaVersion, body.Answers, body.Flags, defaultSV)
	} else {
		merged, mergeErr = mergeReviewerPayload(nil, body.SchemaVersion, body.Answers, body.Flags, defaultSV)
	}
	if mergeErr != nil {
		return constant.ErrInvalidReviewerReport
	}
	if err := validateReviewerAnswerSizes(merged.Answers); err != nil {
		return constant.ErrInvalidReviewerReport
	}
	if err := s.reviewerForm.ValidateAnswers(merged.SchemaVersion, merged.Answers, false); err != nil {
		return constant.ErrInvalidReviewerReport
	}
	raw, err := marshalReviewerPayload(merged)
	if err != nil {
		return constant.ErrInvalidReviewerReport
	}
	row := &entity.ReviewAssignmentReport{
		ReviewAssignmentID: assignmentID,
		Status:             constant.ReviewReportStatusDraft,
		SchemaVersion:      merged.SchemaVersion,
		Payload:            raw,
	}
	if existing != nil {
		row.ID = existing.ID
		row.CreatedAt = existing.CreatedAt
	}
	return s.reviewRepo.UpsertReviewReport(ctx, row)
}

// submitReviewerReviewWithOptionalReport completes the assignment and persists the structured report when applicable.
func (s *Service) submitReviewerReviewWithOptionalReport(ctx context.Context, reviewerID, assignmentID string, body *request.SubmitReviewRequest) error {
	if body == nil {
		body = &request.SubmitReviewRequest{}
	}
	a, err := s.reviewRepo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}
	if a == nil {
		return constant.ErrRecordNotFound
	}
	if err := assertReviewerOwnsAssignment(a, reviewerID); err != nil {
		return err
	}
	if a.Status == constant.ReviewAssignmentStatusWithdrawn {
		return constant.ErrReviewerWithdrawn
	}
	if a.Status == constant.ReviewAssignmentStatusCompleted {
		return constant.ErrReviewAlreadyCompleted
	}
	if a.Status != constant.ReviewAssignmentStatusAccepted {
		return constant.ErrReviewerAssignmentNotAllowed
	}

	recTrim := strings.TrimSpace(body.Recommendation)
	var recPtr *string
	switch recTrim {
	case "":
		recPtr = nil
	case "ACCEPT", "REJECT", "MAJOR_REVISION", "MINOR_REVISION":
		r := recTrim
		recPtr = &r
	default:
		return constant.ErrValidationFailed
	}

	var cmt *string
	if t := strings.TrimSpace(body.Comments); t != "" {
		cmt = &t
	}
	now := time.Now()
	updates := map[string]any{
		"recommendation": recPtr,
		"comments":       cmt,
		"status":         constant.ReviewAssignmentStatusCompleted,
		"completed_at":   now,
	}

	existing, err := s.reviewRepo.GetReportByAssignmentID(ctx, assignmentID)
	if err != nil {
		return err
	}
	if existing != nil && existing.Status == constant.ReviewReportStatusSubmitted {
		return constant.ErrReviewAlreadyCompleted
	}

	var repEntity *entity.ReviewAssignmentReport

	switch {
	case s.reviewerForm != nil:
		defaultSV := s.reviewerForm.SchemaVersion
		var basePayload []byte
		if existing != nil {
			basePayload = existing.Payload
		}
		merged, mErr := mergeReviewerPayload(basePayload, 0, body.Answers, body.Flags, defaultSV)
		if mErr != nil {
			return constant.ErrInvalidReviewerReport
		}
		if body.Report != nil {
			raw, mErr := marshalReviewerPayload(merged)
			if mErr != nil {
				return constant.ErrInvalidReviewerReport
			}
			merged, mErr = mergeReviewerPayload(raw, body.Report.SchemaVersion, body.Report.Answers, body.Report.Flags, defaultSV)
			if mErr != nil {
				return constant.ErrInvalidReviewerReport
			}
		}
		if err := validateReviewerAnswerSizes(merged.Answers); err != nil {
			return constant.ErrInvalidReviewerReport
		}
		if issues := s.reviewerForm.CollectAnswerValidationIssues(merged.SchemaVersion, merged.Answers, true); len(issues) > 0 {
			return invalidReviewerReportWithIssues(issues)
		}
		repEntity, err = reviewerSubmittedReportEntity(assignmentID, existing, merged)
		if err != nil {
			return constant.ErrInvalidReviewerReport
		}

	case body.Report != nil:
		return constant.ErrInternalServerError

	default:
		// Legacy: no embedded form — recommendation + comments only.
	}

	return infrastructure.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.ReviewAssignment{}).Where("id = ?", a.ID).Updates(updates).Error; err != nil {
			return err
		}
		if repEntity == nil {
			return nil
		}
		return s.reviewRepo.UpsertReviewReportDB(ctx, tx, repEntity)
	})
}

func reviewerSubmittedReportEntity(assignmentID string, existing *entity.ReviewAssignmentReport, merged reviewerStoredReport) (*entity.ReviewAssignmentReport, error) {
	raw, err := marshalReviewerPayload(merged)
	if err != nil {
		return nil, err
	}
	rep := &entity.ReviewAssignmentReport{
		ReviewAssignmentID: assignmentID,
		Status:             constant.ReviewReportStatusSubmitted,
		SchemaVersion:      merged.SchemaVersion,
		Payload:            raw,
	}
	if existing != nil {
		rep.ID = existing.ID
		rep.CreatedAt = existing.CreatedAt
	}
	return rep, nil
}

// WithdrawReviewerAssignment sets assignment status to WITHDRAWN when still active.
func (s *Service) WithdrawReviewerAssignment(ctx context.Context, reviewerID, assignmentID string) error {
	a, err := s.reviewRepo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}
	if a == nil {
		return constant.ErrRecordNotFound
	}
	if err := assertReviewerOwnsAssignment(a, reviewerID); err != nil {
		return err
	}
	if a.Status != constant.ReviewAssignmentStatusAccepted {
		return constant.ErrReviewerAssignmentNotAllowed
	}
	now := time.Now()
	return s.reviewRepo.UpdateAssignment(ctx, a.ID, map[string]any{
		"status":     constant.ReviewAssignmentStatusWithdrawn,
		"updated_at": now,
	})
}

// RequestReviewerExtension creates a pending extension request for editors to approve later.
func (s *Service) RequestReviewerExtension(ctx context.Context, reviewerID, assignmentID string, body request.ReviewerExtensionRequestBody) error {
	a, err := s.reviewRepo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}
	if a == nil {
		return constant.ErrRecordNotFound
	}
	if err := assertReviewerOwnsAssignment(a, reviewerID); err != nil {
		return err
	}
	if a.Status != constant.ReviewAssignmentStatusAccepted {
		return constant.ErrReviewerAssignmentNotAllowed
	}
	n, err := s.reviewRepo.CountPendingExtensionRequestsByAssignment(ctx, assignmentID)
	if err != nil {
		return err
	}
	if n > 0 {
		return constant.ErrExtensionRequestPending
	}
	if !body.RequestedDue.UTC().After(a.DueDate.UTC()) {
		return constant.ErrValidationFailed
	}
	now := time.Now()
	row := &entity.ReviewExtensionRequest{
		ID:                 uuid.New().String(),
		ReviewAssignmentID: assignmentID,
		RequestedDue:       body.RequestedDue,
		Reason:             strings.TrimSpace(body.Reason),
		Status:             constant.ReviewExtensionStatusPending,
		CreatedAt:          now,
	}
	return s.reviewRepo.CreateExtensionRequest(ctx, row)
}

// ListEditorExtensionRequests lists extension requests for manuscripts owned by the editor.
func (s *Service) ListEditorExtensionRequests(ctx context.Context, editorID, status string, pg *pagination.Pagination) ([]response.EditorExtensionRequestListItemResponse, int64, error) {
	rows, total, err := s.reviewRepo.ListExtensionRequestsForEditor(ctx, editorID, status, pg)
	if err != nil {
		return nil, 0, err
	}
	out := make([]response.EditorExtensionRequestListItemResponse, 0, len(rows))
	for _, r := range rows {
		reviewerName := strings.TrimSpace(strings.Join([]string{
			derefString(r.ReviewerFirstName),
			derefString(r.ReviewerLastName),
		}, " "))
		out = append(out, response.EditorExtensionRequestListItemResponse{
			ID:                r.RequestID,
			ReviewAssignmentID: r.ReviewAssignmentID,
			RequestedDue:       r.RequestedDue,
			Reason:            derefStringPtr(r.RequestReason),
			Status:            r.RequestStatus,
			CreatedAt:         r.RequestCreatedAt,

			AssignmentDueDate: r.AssignmentDueDate,
			AssignmentStatus:  r.AssignmentStatus,
			ReviewRoundID:     r.ReviewRoundID,

			ManuscriptID:    r.ManuscriptID,
			ManuscriptTitle: r.ManuscriptTitle,
			ReferenceNumber: r.ReferenceNumber,

			ReviewerID:    derefStringPtr(r.ReviewerID),
			ReviewerEmail: derefStringPtr(r.ReviewerEmail),
			ReviewerName:  reviewerName,
		})
	}
	return out, total, nil
}

// DecideExtensionRequest approves/rejects an extension request for manuscripts owned by the editor.
// Approve updates assignment due_date to requested_due (per product decision).
func (s *Service) DecideExtensionRequest(ctx context.Context, editorID, requestID string, decision request.EditorDecideExtensionRequestBody) error {
	row, err := s.reviewRepo.GetExtensionRequestForEditorByID(ctx, editorID, requestID)
	if err != nil {
		return err
	}
	if row == nil {
		// default: do not leak existence of non-owned requests
		return constant.ErrExtensionRequestNotFound
	}
	if row.RequestStatus != string(constant.ReviewExtensionStatusPending) {
		return constant.ErrExtensionRequestAlreadyDecided
	}

	newStatus := constant.ReviewExtensionStatusRejected
	if strings.EqualFold(strings.TrimSpace(decision.Decision), "APPROVE") {
		newStatus = constant.ReviewExtensionStatusApproved
	}

	now := time.Now()
	return infrastructure.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		extUpdates := map[string]any{
			"status":     newStatus,
			"decided_by": editorID,
			"decided_at": now,
			"updated_at": now,
		}
		if err := tx.WithContext(ctx).Model(&entity.ReviewExtensionRequest{}).
			Where("id = ?", row.RequestID).Updates(extUpdates).Error; err != nil {
			return err
		}
		if newStatus == constant.ReviewExtensionStatusApproved {
			if err := tx.WithContext(ctx).Model(&entity.ReviewAssignment{}).
				Where("id = ?", row.ReviewAssignmentID).
				Updates(map[string]any{"due_date": row.RequestedDue, "updated_at": now}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UploadReviewerAssignmentPDF stores a reviewer PDF attachment for the assignment.
func (s *Service) UploadReviewerAssignmentPDF(ctx context.Context, reviewerID, assignmentID string, fh *multipart.FileHeader) (*entity.ReviewFile, error) {
	if fh == nil {
		return nil, constant.ErrValidationFailed
	}
	a, err := s.reviewRepo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, constant.ErrRecordNotFound
	}
	if err := assertReviewerOwnsAssignment(a, reviewerID); err != nil {
		return nil, err
	}
	if a.Status != constant.ReviewAssignmentStatusAccepted {
		return nil, constant.ErrReviewerAssignmentNotAllowed
	}
	allowed := []string{"application/pdf"}
	if err := util.ValidateFile(fh, 25*1024*1024, allowed); err != nil {
		return nil, constant.ErrValidationFailed
	}
	f, err := fh.Open()
	if err != nil {
		return nil, constant.ErrValidationFailed
	}
	defer f.Close()
	body, err := io.ReadAll(io.LimitReader(f, 25*1024*1024+1))
	if err != nil {
		return nil, constant.ErrValidationFailed
	}
	if int64(len(body)) > 25*1024*1024 {
		return nil, constant.ErrValidationFailed
	}
	safeName := path.Base(strings.TrimSpace(fh.Filename))
	if safeName == "" || safeName == "." {
		safeName = "review.pdf"
	}
	storagePath := path.Join("reviews", assignmentID, uuid.New().String()+"_"+safeName)
	publicURL, err := s.UploadReviewFileToStorage(ctx, body, storagePath, "application/pdf")
	if err != nil {
		return nil, err
	}

	aid := assignmentID
	rf := &entity.ReviewFile{
		ID:                 uuid.New().String(),
		ReviewAssignmentID: &aid,
		ReviewRoundID:      a.ReviewRoundID,
		UploadedBy:         reviewerID,
		FileType:           string(constant.ReviewFileTypeReviewerPDF),
		FilePath:           publicURL,
		Filename:           safeName,
		MimeType:           "application/pdf",
		SizeBytes:          int64(len(body)),
		UploadedAt:         time.Now(),
	}
	if err := s.reviewRepo.CreateReviewFile(ctx, rf); err != nil {
		return nil, err
	}
	return rf, nil
}

// AcceptReviewerAssignment marks an INVITED assignment as ACCEPTED for a logged-in reviewer.
func (s *Service) AcceptReviewerAssignment(ctx context.Context, reviewerID, assignmentID string) error {
	a, err := s.reviewRepo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}
	if a == nil {
		return constant.ErrRecordNotFound
	}
	if err := assertReviewerOwnsAssignment(a, reviewerID); err != nil {
		return err
	}
	if a.Status != constant.ReviewAssignmentStatusInvited {
		return constant.ErrReviewerAssignmentNotAllowed
	}

	now := time.Now().UTC()
	updates := map[string]any{
		"status":                 constant.ReviewAssignmentStatusAccepted,
		"invitation_accepted_at": now,
		"updated_at":             now,
	}
	return s.reviewRepo.UpdateAssignment(ctx, a.ID, updates)
}

// DeclineReviewerAssignment marks an INVITED assignment as DECLINED for a logged-in reviewer.
func (s *Service) DeclineReviewerAssignment(ctx context.Context, reviewerID, assignmentID, reason string) error {
	a, err := s.reviewRepo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}
	if a == nil {
		return constant.ErrRecordNotFound
	}
	if err := assertReviewerOwnsAssignment(a, reviewerID); err != nil {
		return err
	}
	if a.Status != constant.ReviewAssignmentStatusInvited {
		return constant.ErrReviewerAssignmentNotAllowed
	}

	now := time.Now().UTC()
	updates := map[string]any{
		"status":     constant.ReviewAssignmentStatusDeclined,
		"comments":   strings.TrimSpace(reason),
		"updated_at": now,
	}
	return s.reviewRepo.UpdateAssignment(ctx, a.ID, updates)
}

func assertReviewerOwnsAssignment(a *entity.ReviewAssignment, reviewerID string) error {
	if a == nil || a.ReviewerID == nil || *a.ReviewerID != reviewerID {
		return constant.ErrForbidden
	}
	return nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func derefStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
