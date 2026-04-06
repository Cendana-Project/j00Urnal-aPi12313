package response

import "time"

type ReviewRoundResponse struct {
	ID               string                     `json:"id"`
	ManuscriptID     string                     `json:"manuscript_id"`
	RoundNumber      int                        `json:"round_number"`
	Status           string                     `json:"status"`
	EditorDecision   *string                    `json:"editor_decision"`
	DecisionComments *string                    `json:"decision_comments"`
	DecisionAt       *time.Time                 `json:"decision_at"`
	CreatedBy        string                     `json:"created_by"`
	CreatedAt        time.Time                  `json:"created_at"`
	Assignments      []ReviewAssignmentResponse `json:"assignments,omitempty"`
}

type ReviewAssignmentResponse struct {
	ID                   string     `json:"id"`
	ReviewRoundID        string     `json:"review_round_id"`
	ReviewerID           string     `json:"reviewer_id,omitempty"`
	InvitedEmail         string     `json:"invited_email,omitempty"`
	ReviewerName         string     `json:"reviewer_name"`
	ReviewerEmail        string     `json:"reviewer_email"`
	AssignedBy           string     `json:"assigned_by"`
	Status               string     `json:"status"`
	InvitationExpiresAt  time.Time  `json:"invitation_expires_at"`
	InvitationAcceptedAt *time.Time `json:"invitation_accepted_at,omitempty"`
	DueDate              time.Time  `json:"due_date"`
	Recommendation       *string    `json:"recommendation,omitempty"`
	Comments             *string    `json:"comments,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

type ReviewerCandidateResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Affiliation string `json:"affiliation,omitempty"`
	ActiveCount int    `json:"active_count"`
	DoneCount   int    `json:"done_count"`
}

type EditorCandidateResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type SubmissionListItemResponse struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Abstract           string     `json:"abstract"`
	Status             string     `json:"status"`
	JournalName        string     `json:"journal_name,omitempty"`
	MainAuthorID       string     `json:"main_author_id"`
	MainAuthorName     string     `json:"main_author_name"`
	MainAuthorEmail    string     `json:"main_author_email"`
	AssignedEditorID   *string    `json:"assigned_editor_id,omitempty"`
	AssignedEditorName *string    `json:"assigned_editor_name,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at"`
}

type ReviewDetailResponse struct {
	ManuscriptID string                `json:"manuscript_id"`
	Title        string                `json:"title"`
	Status       string                `json:"status"`
	Rounds       []ReviewRoundResponse `json:"rounds"`
}

// ReviewerInvitationPreviewResponse is a public, non-sensitive view for the invitation landing page.
type ReviewerInvitationPreviewResponse struct {
	Valid               bool      `json:"valid"`
	InvitationExpiresAt time.Time `json:"invitation_expires_at"`
	DueDate             time.Time `json:"due_date"`
	ManuscriptTitle     string    `json:"manuscript_title"`
	JournalName         string    `json:"journal_name,omitempty"`
	EditorName          string    `json:"editor_name,omitempty"`
	InvitedEmailMasked  string    `json:"invited_email_masked"`
	RequiresPassword    bool      `json:"requires_password"`
	Section             string    `json:"section"`
}

// ReviewerWorkflowStep matches the reviewer portal progress stepper (initial validation through final decision).
type ReviewerWorkflowStep struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Done    bool   `json:"done"`
	Current bool   `json:"current"`
}

// ReviewerAssignmentListItemResponse is one row in GET /v1/reviewer/assignments.
type ReviewerAssignmentListItemResponse struct {
	AssignmentID      string     `json:"assignment_id"`
	ReviewRoundID     string     `json:"review_round_id"`
	ManuscriptID      string     `json:"manuscript_id"`
	ManuscriptTitle   string     `json:"manuscript_title"`
	ManuscriptSection string     `json:"manuscript_section"`
	ReferenceNumber   *int64     `json:"reference_number,omitempty"`
	RoundNumber       int        `json:"round_number"`
	RoundStatus       string     `json:"round_status"`
	AssignmentStatus  string     `json:"assignment_status"`
	DueDate           time.Time  `json:"due_date"`
	IsDelayed         bool       `json:"is_delayed"`
	CreatedAt         time.Time  `json:"created_at"`
}

// ReviewerAuthorBrief is a co-author line on the reviewer workspace.
type ReviewerAuthorBrief struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	IsCorresponding   bool   `json:"is_corresponding"`
	OrderPosition     int    `json:"order_position"`
}

// ReviewerAssignmentFileResponse is metadata for a file linked to the assignment.
type ReviewerAssignmentFileResponse struct {
	ID          string    `json:"id"`
	FileType    string    `json:"file_type"`
	Filename    string    `json:"filename"`
	MimeType    string    `json:"mime_type"`
	SizeBytes   int64     `json:"size_bytes"`
	UploadedAt  time.Time `json:"uploaded_at"`
	UploadedBy  string    `json:"uploaded_by"`
	UploaderName string   `json:"uploader_name,omitempty"`
}

// ReviewerSavedReportResponse is structured answers loaded from review_assignment_reports.
type ReviewerSavedReportResponse struct {
	Status         string         `json:"status"`
	SchemaVersion  int            `json:"schema_version"`
	Answers        map[string]any `json:"answers"`
	Flags          map[string]any `json:"flags,omitempty"`
}

// ReviewerWorkspaceAssignment is the logged-in reviewer's task on this round (no invitation token noise).
type ReviewerWorkspaceAssignment struct {
	ID                   string     `json:"id"`
	Status               string     `json:"status"`
	DueDate              time.Time  `json:"due_date"`
	IsDelayed            bool       `json:"is_delayed"`
	InvitationAcceptedAt *time.Time `json:"invitation_accepted_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	Recommendation       *string    `json:"recommendation,omitempty"`
	Comments             *string    `json:"comments,omitempty"`
}

// ReviewerWorkspaceRound is the review round that owns this assignment.
type ReviewerWorkspaceRound struct {
	ID             string     `json:"id"`
	RoundNumber    int        `json:"round_number"`
	Status         string     `json:"status"`
	EditorDecision *string    `json:"editor_decision,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ReviewerWorkspaceManuscript is manuscript + journal + handling editor context for the workspace page.
type ReviewerWorkspaceManuscript struct {
	ID              string               `json:"id"`
	Title           string               `json:"title"`
	Status          string               `json:"status"`
	Section         string               `json:"section"`
	Abstract        string               `json:"abstract"`
	ReferenceNumber *int64               `json:"reference_number,omitempty"`
	JournalName     string               `json:"journal_name,omitempty"`
	ReceivedAt      time.Time            `json:"received_at"`
	HandlingEditor  string               `json:"handling_editor_name,omitempty"`
	Authors         []ReviewerAuthorBrief `json:"authors"`
}

// ReviewerWorkspaceResponse is GET /v1/reviewer/assignments/:id.
// Form copy + field definitions: GET /v1/reviewer/review-form-schema (hindari duplikasi payload besar).
type ReviewerWorkspaceResponse struct {
	Assignment          ReviewerWorkspaceAssignment      `json:"assignment"`
	Round               ReviewerWorkspaceRound           `json:"round"`
	Manuscript          ReviewerWorkspaceManuscript      `json:"manuscript"`
	Report              *ReviewerSavedReportResponse     `json:"report,omitempty"`
	Files               []ReviewerAssignmentFileResponse `json:"files"`
	Workflow            []ReviewerWorkflowStep           `json:"workflow"`
	CurrentWorkflowStep int                              `json:"current_workflow_step"`
}

// EditorExtensionRequestListItemResponse is one row in GET /v1/editor/extension-requests.
type EditorExtensionRequestListItemResponse struct {
	ID                string     `json:"id"`
	ReviewAssignmentID string    `json:"review_assignment_id"`
	RequestedDue       time.Time `json:"requested_due"`
	Reason            *string    `json:"reason,omitempty"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`

	AssignmentDueDate time.Time `json:"assignment_due_date"`
	AssignmentStatus  string    `json:"assignment_status"`
	ReviewRoundID     string    `json:"review_round_id"`

	ManuscriptID    string `json:"manuscript_id"`
	ManuscriptTitle string `json:"manuscript_title"`
	ReferenceNumber *int64 `json:"reference_number,omitempty"`

	ReviewerID    *string `json:"reviewer_id,omitempty"`
	ReviewerName  string  `json:"reviewer_name,omitempty"`
	ReviewerEmail *string `json:"reviewer_email,omitempty"`
}

// ReviewerHistoryItemResponse matches reviewer History table columns:
// ID, MM-DD ASSIGNED, SEC, TITLE, REVIEW, EDITOR DECISION (+ ids for navigation).
type ReviewerHistoryItemResponse struct {
	ID                   int        `json:"id"` // 1-based row number (table ID column)
	AssignmentID         string     `json:"assignment_id"`
	ManuscriptID         string     `json:"manuscript_id"`
	MMDDAssigned         string     `json:"mm_dd_assigned"` // e.g. "05-24" (UTC)
	Sec                  string     `json:"sec"`
	Title                string     `json:"title"`
	Review               string     `json:"review"`            // reviewer recommendation (display)
	EditorDecision       string     `json:"editor_decision"`   // editor decision (display)
	RecommendationCode   *string    `json:"recommendation_code,omitempty"`
	EditorDecisionCode   *string    `json:"editor_decision_code,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
}
