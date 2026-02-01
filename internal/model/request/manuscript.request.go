package request

type CreateManuscriptRequest struct {
	Title         string `json:"title" binding:"required"`
	Abstract      string `json:"abstract" binding:"required"`
	JournalID     string `json:"journal_id" binding:"required,uuid"`
	IsTncAccepted bool   `json:"is_tnc_accepted" binding:"required,eq=true"` // Must be true
	CurrentStep   int    `json:"current_step"`
}

type UpdateManuscriptRequest struct {
	Title       string `json:"title"`
	Abstract    string `json:"abstract"`
	CurrentStep int    `json:"current_step"`
}

type UpdateManuscriptAuthorsRequest struct {
	Authors []ManuscriptAuthorRequest `json:"authors" binding:"required,dive"`
}

type ManuscriptAuthorRequest struct {
	UserID          *string `json:"user_id"`
	AuthorName      string  `json:"author_name" binding:"required"`
	AuthorEmail     string  `json:"author_email" binding:"required,email"`
	Affiliation     string  `json:"affiliation"`
	IsCorresponding bool    `json:"is_corresponding"`
	OrderPosition   int     `json:"order_position"`
}

type PublishManuscriptRequest struct {
	IssueID string `json:"issue_id" binding:"required,uuid"`
}
