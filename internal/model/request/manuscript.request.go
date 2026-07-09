package request

import "github.com/api-monolith-template/internal/constant"

// SubmitManuscriptAuthorInput: daftar bibliografi saja. Submitter tidak wajib termasuk; jika email submitter ada di salah satu baris, baris itu ditandai is_primary_contact di response. Wajib tepat satu is_primary_author.
type SubmitManuscriptAuthorInput struct {
	Name             string `json:"name" binding:"required"`
	Email            string `json:"email" binding:"required,email"`
	Affiliation      string `json:"affiliation"`
	OrderPosition    int    `json:"order_position" binding:"required,gte=1"`
	IsPrimaryAuthor  bool   `json:"is_primary_author"`
}

type CreateManuscriptRequest struct {
	Title         string                        `json:"title" binding:"required"`
	Abstract      string                        `json:"abstract" binding:"required"`
	JournalID     string                        `json:"journal_id" binding:"required,uuid"`
	Authors       []SubmitManuscriptAuthorInput `json:"authors" binding:"required,min=1,dive"`
	IsTncAccepted bool                          `json:"is_tnc_accepted" binding:"required,eq=true"`
}

type UpdateManuscriptRequest struct {
	Title    string `json:"title" binding:"required"`
	Abstract string `json:"abstract" binding:"required"`
}

type UpdateManuscriptAuthorsRequest struct {
	Authors []ManuscriptAuthorRequest `json:"authors" binding:"required,min=1,dive"`
}

type ManuscriptAuthorRequest struct {
	UserID          *string `json:"user_id"`
	AuthorName      string  `json:"author_name" binding:"required"`
	AuthorEmail     string  `json:"author_email" binding:"required,email"`
	Affiliation     string  `json:"affiliation"`
	OrderPosition   int     `json:"order_position" binding:"required,gte=1"`
	IsPrimaryAuthor bool    `json:"is_primary_author"`
}

type PublishManuscriptRequest struct {
	IssueID string `json:"issue_id"`
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
