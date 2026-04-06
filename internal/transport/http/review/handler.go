package review

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/mapper"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/reviewerform"
	reviewSvc "github.com/api-monolith-template/internal/service/review"
	"github.com/api-monolith-template/internal/util"
	"github.com/api-monolith-template/pkg/pagination"
)

type Controller struct {
	svc *reviewSvc.Service
}

func NewController(svc *reviewSvc.Service) *Controller {
	return &Controller{svc: svc}
}

// ====================================================================
// Chief Editor Endpoints
// ====================================================================

// ListChiefEditorSubmissions lists manuscripts that need editor assignment.
func (c *Controller) ListChiefEditorSubmissions(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	pg := pagination.New(page, pageSize)

	statuses := constant.StatusQueue
	statuses = append(statuses, constant.ManuscriptStatusAccepted, constant.ManuscriptStatusRejected)

	manuscripts, total, err := c.svc.ListSubmissions(ctx.Request.Context(), statuses, pg)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = gin.H{
		"items":     mapper.ToSubmissionListResponse(manuscripts),
		"total":     total,
		"page":      pg.Page,
		"page_size": pg.PageSize,
	}
	util.HandleResponse(ctx, res, nil)
}

// ListEditorCandidates returns users with EDITOR role for assignment.
func (c *Controller) ListEditorCandidates(ctx *gin.Context) {
	search := ctx.Query("search")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	pg := pagination.New(page, pageSize)

	candidates, total, err := c.svc.ListEditorCandidates(ctx.Request.Context(), search, pg)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = gin.H{
		"items":     mapper.ToEditorCandidateListResponse(candidates),
		"total":     total,
		"page":      pg.Page,
		"page_size": pg.PageSize,
	}
	util.HandleResponse(ctx, res, nil)
}

// AssignEditor assigns an editor to a manuscript (Chief Editor action).
func (c *Controller) AssignEditor(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}

	var req request.AssignEditorRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	if err := c.svc.AssignEditor(ctx.Request.Context(), req.ManuscriptID, req.EditorID, userID); err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Message = "Editor assigned successfully"
	util.HandleResponse(ctx, res, nil)
}

// ====================================================================
// Editor Endpoints
// ====================================================================

// ListEditorSubmissions lists manuscripts assigned to the current editor.
func (c *Controller) ListEditorSubmissions(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}

	tab := ctx.DefaultQuery("tab", "all")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	pg := pagination.New(page, pageSize)

	manuscripts, total, err := c.svc.ListEditorSubmissions(ctx.Request.Context(), userID, tab, pg)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = gin.H{
		"items":     mapper.ToSubmissionListResponse(manuscripts),
		"total":     total,
		"page":      pg.Page,
		"page_size": pg.PageSize,
	}
	util.HandleResponse(ctx, res, nil)
}

// ListEditorExtensionRequests lists extension requests for manuscripts owned by the current editor.
func (c *Controller) ListEditorExtensionRequests(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}
	status := strings.TrimSpace(ctx.DefaultQuery("status", string(constant.ReviewExtensionStatusPending)))
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	pg := pagination.New(page, pageSize)

	items, total, err := c.svc.ListEditorExtensionRequests(ctx.Request.Context(), userID, status, pg)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Data = gin.H{
		"items":     items,
		"total":     total,
		"page":      pg.Page,
		"page_size": pg.PageSize,
	}
	util.HandleResponse(ctx, res, nil)
}

// DecideExtensionRequest approves/rejects a pending extension request.
func (c *Controller) DecideExtensionRequest(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}
	reqID := strings.TrimSpace(ctx.Param("id"))
	if reqID == "" {
		util.HandleError(ctx, constant.ErrInvalidUUIDFormat)
		return
	}
	var body request.EditorDecideExtensionRequestBody
	if err := util.BindAndValidate(ctx, &body); err != nil {
		util.HandleError(ctx, err)
		return
	}
	if err := c.svc.DecideExtensionRequest(ctx.Request.Context(), userID, reqID, body); err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Message = "Extension request decided"
	util.HandleResponse(ctx, res, nil)
}

// SendToReview creates a review round and transitions manuscript to UNDER_REVIEW.
func (c *Controller) SendToReview(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}

	var req struct {
		ManuscriptID string `json:"manuscript_id" binding:"required,uuid"`
	}
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	round, err := c.svc.SendToReview(ctx.Request.Context(), req.ManuscriptID, userID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.StatusCode = http.StatusCreated
	res.Message = "Review round created"
	if round != nil {
		res.Data = mapper.ToReviewRoundResponse(round)
	}
	util.HandleResponse(ctx, res, nil)
}

// AcceptManuscript accepts the manuscript.
func (c *Controller) AcceptManuscript(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}

	var req request.AcceptManuscriptRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	if err := c.svc.AcceptManuscript(ctx.Request.Context(), req.ManuscriptID, userID); err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Message = "Manuscript accepted"
	util.HandleResponse(ctx, res, nil)
}

// DeclineManuscript rejects the manuscript with a reason.
func (c *Controller) DeclineManuscript(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}

	var req request.DeclineManuscriptRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	if err := c.svc.DeclineManuscript(ctx.Request.Context(), req.ManuscriptID, userID, req.Reason); err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Message = "Manuscript declined"
	util.HandleResponse(ctx, res, nil)
}

// RequestRevision requests revisions from the author.
func (c *Controller) RequestRevision(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}

	var req request.RevisionRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	if err := c.svc.RequestRevision(ctx.Request.Context(), req.ManuscriptID, userID, req.Comments); err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Message = "Revision requested"
	util.HandleResponse(ctx, res, nil)
}

// GetReviewDetails returns all review rounds and assignments for a manuscript.
func (c *Controller) GetReviewDetails(ctx *gin.Context) {
	manuscriptID := ctx.Param("id")

	m, rounds, err := c.svc.GetReviewDetails(ctx.Request.Context(), manuscriptID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToReviewDetailResponse(m, rounds)
	util.HandleResponse(ctx, res, nil)
}

// ListReviewerCandidates returns users with REVIEWER role.
func (c *Controller) ListReviewerCandidates(ctx *gin.Context) {
	search := ctx.Query("search")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	pg := pagination.New(page, pageSize)

	candidates, total, err := c.svc.ListReviewerCandidates(ctx.Request.Context(), search, pg)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = gin.H{
		"items":     mapper.ToReviewerCandidateListResponse(candidates),
		"total":     total,
		"page":      pg.Page,
		"page_size": pg.PageSize,
	}
	util.HandleResponse(ctx, res, nil)
}

// InviteReviewer invites a reviewer to a review round.
func (c *Controller) InviteReviewer(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}

	var req request.InviteReviewerRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	assignment, err := c.svc.InviteReviewer(ctx.Request.Context(), req.ManuscriptID, req.Email, userID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.StatusCode = http.StatusCreated
	res.Message = "Reviewer invited"
	if assignment != nil {
		res.Data = mapper.ToReviewAssignmentResponse(assignment)
	}
	util.HandleResponse(ctx, res, nil)
}

// MakeRoundDecision makes a decision for a review round.
func (c *Controller) MakeRoundDecision(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}

	var req request.RoundDecisionRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	if err := c.svc.MakeRoundDecision(ctx.Request.Context(), req.RoundID, userID, req.Decision, req.Comments); err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Message = "Decision recorded"
	util.HandleResponse(ctx, res, nil)
}

// ====================================================================
// Reviewer portal (invitation link + history)
// ====================================================================

// GetInvitationPreview is public: metadata for landing page from email link.
func (c *Controller) GetInvitationPreview(ctx *gin.Context) {
	token := strings.TrimSpace(ctx.Param("token"))
	if token == "" {
		util.HandleError(ctx, constant.ErrInvitationNotFound)
		return
	}
	data, err := c.svc.GetInvitationPreview(ctx.Request.Context(), token)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Data = data
	util.HandleResponse(ctx, res, nil)
}

// CompleteReviewerInvitation is public: token + password, auto REVIEWER role.
func (c *Controller) CompleteReviewerInvitation(ctx *gin.Context) {
	token := strings.TrimSpace(ctx.Param("token"))
	if token == "" {
		util.HandleError(ctx, constant.ErrInvitationNotFound)
		return
	}
	var req request.CompleteReviewerInvitationRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}
	if err := c.svc.CompleteReviewerInvitation(ctx.Request.Context(), token, req.Password, req.FirstName, req.LastName); err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Message = "Invitation accepted. You can sign in with your email and password."
	util.HandleResponse(ctx, res, nil)
}

// DeclineReviewerInvitation is public.
func (c *Controller) DeclineReviewerInvitation(ctx *gin.Context) {
	token := strings.TrimSpace(ctx.Param("token"))
	if token == "" {
		util.HandleError(ctx, constant.ErrInvitationNotFound)
		return
	}
	var req request.DeclineReviewerInvitationRequest
	_ = ctx.ShouldBindJSON(&req)
	if err := c.svc.DeclineReviewerInvitation(ctx.Request.Context(), token, req.Reason); err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Message = "Invitation declined"
	util.HandleResponse(ctx, res, nil)
}

// ListReviewerHistory lists completed reviews for the logged-in reviewer.
func (c *Controller) ListReviewerHistory(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	page, pageSize = pagination.NormalizeQuery(page, pageSize)
	pg := pagination.New(page, pageSize)
	search := strings.TrimSpace(ctx.Query("search"))
	recFilter := strings.TrimSpace(ctx.Query("recommendation"))
	edFilter := strings.TrimSpace(ctx.Query("editor_decision"))

	if recFilter != "" {
		switch recFilter {
		case "ACCEPT", "REJECT", "MAJOR_REVISION", "MINOR_REVISION":
		default:
			util.HandleError(ctx, constant.ErrValidationFailed)
			return
		}
	}
	if edFilter != "" {
		switch edFilter {
		case "ACCEPT", "REJECT", "REVISION_REQUIRED":
		default:
			util.HandleError(ctx, constant.ErrValidationFailed)
			return
		}
	}

	rows, total, err := c.svc.ListReviewerHistory(ctx.Request.Context(), userID, search, recFilter, edFilter, pg)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Data = gin.H{
		"items":     mapper.ToReviewerHistoryListResponse(rows, pg.Page, pg.PageSize),
		"total":     total,
		"page":      pg.Page,
		"page_size": pg.PageSize,
	}
	util.HandleResponse(ctx, res, nil)
}

// SubmitReviewerReview completes the assignment. Body may be empty; optional fields: recommendation, comments,
// answers, flags (merged into draft like PATCH .../draft), report (second merge). No extra save call needed to fix missing fields.
func (c *Controller) SubmitReviewerReview(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}
	assignmentID := strings.TrimSpace(ctx.Param("id"))
	if assignmentID == "" {
		util.HandleError(ctx, constant.ErrInvalidUUIDFormat)
		return
	}
	var req request.SubmitReviewRequest
	if err := util.BindJSONOrEmpty(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}
	if err := c.svc.SubmitReviewerReview(ctx.Request.Context(), userID, assignmentID, req); err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Message = "Review submitted"
	util.HandleResponse(ctx, res, nil)
}

// ListReviewerAssignments lists pending or completed assignments for the logged-in reviewer.
func (c *Controller) ListReviewerAssignments(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	page, pageSize = pagination.NormalizeQuery(page, pageSize)
	pg := pagination.New(page, pageSize)

	statusStr := strings.ToUpper(strings.TrimSpace(ctx.DefaultQuery("status", "ACCEPTED")))
	var st constant.ReviewAssignmentStatus
	switch statusStr {
	case "ACCEPTED":
		st = constant.ReviewAssignmentStatusAccepted
	case "COMPLETED":
		st = constant.ReviewAssignmentStatusCompleted
	default:
		util.HandleError(ctx, constant.ErrValidationFailed)
		return
	}

	items, total, err := c.svc.ListReviewerAssignments(ctx.Request.Context(), userID, st, pg)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Data = gin.H{
		"items":     items,
		"total":     total,
		"page":      pg.Page,
		"page_size": pg.PageSize,
	}
	util.HandleResponse(ctx, res, nil)
}

// GetReviewerAssignmentWorkspace returns nested assignment, round, manuscript, report, files, and workflow stepper.
// Form schema: GET /v1/reviewer/review-form-schema.
func (c *Controller) GetReviewerAssignmentWorkspace(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}
	assignmentID := strings.TrimSpace(ctx.Param("id"))
	if assignmentID == "" {
		util.HandleError(ctx, constant.ErrInvalidUUIDFormat)
		return
	}
	data, err := c.svc.GetReviewerAssignmentWorkspace(ctx.Request.Context(), userID, assignmentID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Data = data
	util.HandleResponse(ctx, res, nil)
}

// GetReviewerFormSchema returns schema_version and sections whose fields only expose id, label, and options (checklist flattened).
func (c *Controller) GetReviewerFormSchema(ctx *gin.Context) {
	s, err := reviewerform.Load()
	if err != nil {
		util.HandleError(ctx, constant.ErrInternalServerError)
		return
	}
	res := response.NewResponseOK()
	res.Data = s.ToPublic()
	util.HandleResponse(ctx, res, nil)
}

// PatchReviewerReportDraft saves or merges a structured draft for an assignment.
func (c *Controller) PatchReviewerReportDraft(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}
	assignmentID := strings.TrimSpace(ctx.Param("id"))
	if assignmentID == "" {
		util.HandleError(ctx, constant.ErrInvalidUUIDFormat)
		return
	}
	var reqBody request.PatchReviewerReportDraftRequest
	if err := util.BindAndValidate(ctx, &reqBody); err != nil {
		util.HandleError(ctx, err)
		return
	}
	if err := c.svc.PatchReviewerReportDraft(ctx.Request.Context(), userID, assignmentID, reqBody); err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Message = "Draft saved"
	util.HandleResponse(ctx, res, nil)
}

// UploadReviewerPDF accepts a PDF file for the assignment (multipart field "file").
func (c *Controller) UploadReviewerPDF(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}
	assignmentID := strings.TrimSpace(ctx.Param("id"))
	if assignmentID == "" {
		util.HandleError(ctx, constant.ErrInvalidUUIDFormat)
		return
	}
	fh, err := ctx.FormFile("file")
	if err != nil {
		util.HandleError(ctx, constant.ErrValidationError)
		return
	}
	rf, err := c.svc.UploadReviewerAssignmentPDF(ctx.Request.Context(), userID, assignmentID, fh)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Data = gin.H{
		"id":          rf.ID,
		"file_type":   rf.FileType,
		"filename":    rf.Filename,
		"size_bytes":  rf.SizeBytes,
		"uploaded_at": rf.UploadedAt,
	}
	util.HandleResponse(ctx, res, nil)
}

// WithdrawReviewerAssignment withdraws an accepted assignment before completion.
func (c *Controller) WithdrawReviewerAssignment(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}
	assignmentID := strings.TrimSpace(ctx.Param("id"))
	if assignmentID == "" {
		util.HandleError(ctx, constant.ErrInvalidUUIDFormat)
		return
	}
	if err := c.svc.WithdrawReviewerAssignment(ctx.Request.Context(), userID, assignmentID); err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Message = "Assignment withdrawn"
	util.HandleResponse(ctx, res, nil)
}

// RequestReviewerExtension creates a pending due-date extension request.
func (c *Controller) RequestReviewerExtension(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		util.HandleError(ctx, constant.ErrUnauthorized)
		return
	}
	assignmentID := strings.TrimSpace(ctx.Param("id"))
	if assignmentID == "" {
		util.HandleError(ctx, constant.ErrInvalidUUIDFormat)
		return
	}
	var reqBody request.ReviewerExtensionRequestBody
	if err := util.BindAndValidate(ctx, &reqBody); err != nil {
		util.HandleError(ctx, err)
		return
	}
	if err := c.svc.RequestReviewerExtension(ctx.Request.Context(), userID, assignmentID, reqBody); err != nil {
		util.HandleError(ctx, err)
		return
	}
	res := response.NewResponseOK()
	res.Message = "Extension request submitted"
	util.HandleResponse(ctx, res, nil)
}
