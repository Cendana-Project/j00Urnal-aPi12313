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
	ReviewerID           string     `json:"reviewer_id"`
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
