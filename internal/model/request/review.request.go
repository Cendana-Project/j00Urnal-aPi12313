package request

import "time"

type AssignEditorRequest struct {
	ManuscriptID string `json:"manuscript_id" binding:"required,uuid"`
	EditorID     string `json:"editor_id" binding:"required,uuid"`
}

type InviteReviewerRequest struct {
	ManuscriptID string    `json:"manuscript_id" binding:"required,uuid"`
	Email        string    `json:"email" binding:"required,email"`
	DueDate      time.Time `json:"due_date" binding:"required"`
}

// InviteRegisteredReviewerRequest assigns an existing user (REVIEWER role) to a review round with a due date.
type InviteRegisteredReviewerRequest struct {
	RoundID    string    `json:"round_id" binding:"required,uuid"`
	ReviewerID string    `json:"reviewer_id" binding:"required,uuid"`
	DueDate    time.Time `json:"due_date" binding:"required"`
}

// CompleteReviewerInvitationRequest sets password and optional profile fields after opening the email link.
type CompleteReviewerInvitationRequest struct {
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// DeclineReviewerInvitationRequest optional reason when declining via token.
type DeclineReviewerInvitationRequest struct {
	Reason string `json:"reason"`
}

// ReviewerReportPayload is optional structured answers for independent review (keys match form schema field ids).
// SchemaVersion is optional; when omitted or 0 the server uses the embedded form schema version.
type ReviewerReportPayload struct {
	SchemaVersion int            `json:"schema_version,omitempty"`
	Answers       map[string]any `json:"answers"`
	Flags         map[string]any `json:"flags,omitempty"`
}

// SubmitReviewRequest is POST .../assignments/:id/submit. Body may be empty: the server finalizes the saved draft
// and may leave recommendation/comments null. When recommendation is sent it must be a valid enum.
// Answers/Flags merge into the stored draft before validation (same as PATCH .../draft) so clients can fix missing
// fields and resubmit without an extra save request.
type SubmitReviewRequest struct {
	Recommendation string                 `json:"recommendation,omitempty" binding:"omitempty,oneof=ACCEPT REJECT MAJOR_REVISION MINOR_REVISION"`
	Comments       string                 `json:"comments,omitempty"`
	Answers        map[string]any         `json:"answers,omitempty"`
	Flags          map[string]any         `json:"flags,omitempty"`
	Report         *ReviewerReportPayload `json:"report,omitempty"`
}

// PatchReviewerReportDraftRequest merges answers/flags into the saved draft for an assignment.
// SchemaVersion is optional; when omitted or 0 the server uses the embedded form schema version.
type PatchReviewerReportDraftRequest struct {
	SchemaVersion int            `json:"schema_version,omitempty"`
	Answers       map[string]any `json:"answers"`
	Flags         map[string]any `json:"flags,omitempty"`
}

// ReviewerExtensionRequestBody creates a pending extension request for editors to approve later.
type ReviewerExtensionRequestBody struct {
	RequestedDue time.Time `json:"requested_due" binding:"required"`
	Reason       string    `json:"reason"`
}

// EditorDecideExtensionRequestBody is POST /v1/editor/extension-requests/:id/decision.
// Reason is optional and may be stored for audit purposes.
type EditorDecideExtensionRequestBody struct {
	Decision string `json:"decision" binding:"required,oneof=APPROVE REJECT"`
	Reason   string `json:"reason"`
}

type RoundDecisionRequest struct {
	RoundID  string `json:"round_id"`
	// Alias: FE sends review_round_id on MakeRoundDecision
	ReviewRoundID string `json:"review_round_id"`
	Decision string `json:"decision" binding:"required,oneof=ACCEPT REJECT REVISION_REQUIRED SKIP_ACCEPT"`
	Comments string `json:"comments"`
}

type AcceptManuscriptRequest struct {
	ManuscriptID string `json:"manuscript_id" binding:"required,uuid"`
}

type DeclineManuscriptRequest struct {
	ManuscriptID string `json:"manuscript_id" binding:"required,uuid"`
	Reason       string `json:"reason" binding:"required"`
}

type RevisionRequest struct {
	ManuscriptID string `json:"manuscript_id" binding:"required,uuid"`
	Comments     string `json:"comments" binding:"required"`
}
