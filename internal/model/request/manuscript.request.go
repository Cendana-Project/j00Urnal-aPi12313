package request

import "github.com/api-monolith-template/internal/constant"

type CreateManuscriptRequest struct {
	Title         string `json:"title" binding:"required"`
	Abstract      string `json:"abstract" binding:"required"`
	JournalID     string `json:"journal_id" binding:"required,uuid"`
	MainAuthorID string `json:"main_author_id" binding:"omitempty,uuid"`
	IsTncAccepted bool   `json:"is_tnc_accepted" binding:"required,eq=true"` // Must be true
}

type UpdateManuscriptRequest struct {
	Title    string `json:"title" binding:"required"`
	Abstract string `json:"abstract" binding:"required"`
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

type AuthorManuscriptFilterRequest struct {
	Statuses     []constant.ManuscriptStatus `json:"statuses"`
	StartDate    *string                     `json:"start_date"`
	EndDate      *string                     `json:"end_date"`
	SearchTitle  string                      `json:"search_title"`
	SearchAuthor string                      `json:"search_author"`
	Page         int                         `json:"page" binding:"required"`
	PageSize     int                         `json:"page_size" binding:"required"`
}
