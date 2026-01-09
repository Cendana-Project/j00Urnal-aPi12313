package request

type CreateManuscriptRequest struct {
	IssueID  string `json:"issue_id" binding:"required,uuid"`
	Title    string `json:"title" binding:"required"`
	Abstract string `json:"abstract" binding:"required"`
}

type UpdateManuscriptRequest struct {
	Title    string `json:"title" binding:"required"`
	Abstract string `json:"abstract" binding:"required"`
}

type ManuscriptAuthorRequest struct {
	UserID          *string `json:"user_id,omitempty" binding:"omitempty,uuid"`
	AuthorName      string  `json:"author_name" binding:"required"`
	AuthorEmail     string  `json:"author_email" binding:"required,email"`
	Affiliation     string  `json:"affiliation"`
	IsCorresponding bool    `json:"is_corresponding"`
	OrderPosition   int     `json:"order_position" binding:"required"`
}

type UpdateManuscriptAuthorsRequest struct {
	Authors []ManuscriptAuthorRequest `json:"authors" binding:"required,dive"`
}
