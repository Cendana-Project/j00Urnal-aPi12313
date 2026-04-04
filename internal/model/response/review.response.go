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
