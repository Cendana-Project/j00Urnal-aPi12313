package request

type AssignEditorRequest struct {
	ManuscriptID string `json:"manuscript_id" binding:"required,uuid"`
	EditorID     string `json:"editor_id" binding:"required,uuid"`
}

type InviteReviewerRequest struct {
	RoundID    string `json:"round_id" binding:"required,uuid"`
	ReviewerID string `json:"reviewer_id" binding:"required,uuid"`
	DueDate    string `json:"due_date" binding:"required"` // RFC3339
}

type SubmitReviewRequest struct {
	Recommendation string `json:"recommendation" binding:"required,oneof=ACCEPT REJECT MAJOR_REVISION MINOR_REVISION"`
	Comments       string `json:"comments"`
}

type RoundDecisionRequest struct {
	RoundID  string `json:"round_id" binding:"required,uuid"`
	Decision string `json:"decision" binding:"required,oneof=ACCEPT REJECT REVISION_REQUIRED"`
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
