package response

import (
	"time"
)

type ManuscriptResponse struct {
	ID            string                     `json:"id"`
	IssueID       *string                    `json:"issue_id"`
	Title         string                     `json:"title"`
	Abstract      string                     `json:"abstract"`
	Status        string                     `json:"status"`
	CurrentStep   int                        `json:"current_step"`
	MainAuthorID  string                     `json:"main_author_id"`
	PublishedAt   time.Time                  `json:"published_at"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     *time.Time                 `json:"updated_at"`
	Authors       []ManuscriptAuthorResponse `json:"authors,omitempty"`
	AuthorsSorted []string                   `json:"authors_sorted,omitempty"`
	Files         []ManuscriptFileResponse   `json:"files,omitempty"`
}

type ManuscriptAuthorResponse struct {
	ID              string  `json:"id"`
	UserID          *string `json:"user_id,omitempty"`
	AuthorName      string  `json:"author_name"`
	AuthorEmail     string  `json:"author_email"`
	Affiliation     string  `json:"affiliation"`
	IsCorresponding bool    `json:"is_corresponding"`
	OrderPosition   int     `json:"order_position"`
}

type ManuscriptFileResponse struct {
	ID         string    `json:"id"`
	FileType   string    `json:"file_type"`
	FilePath   string    `json:"file_path"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type"`
	SizeBytes  int64     `json:"size_bytes"`
	Version    int       `json:"version"`
	UploadedAt time.Time `json:"uploaded_at"`
}
