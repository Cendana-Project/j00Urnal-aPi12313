package response

import (
	"time"
)

type ManuscriptResponse struct {
	ID                 string                     `json:"id"`
	IssueID            *string                    `json:"issue_id"`
	Title              string                     `json:"title"`
	Abstract           string                     `json:"abstract"`
	Status             string                     `json:"status"`
	MainAuthorID       *string                    `json:"main_author_id,omitempty"`
	SubmittedByUserID  *string                    `json:"submitted_by_user_id,omitempty"`
	Submitter          *ManuscriptSubmitterBrief  `json:"submitter,omitempty"`
	AssignedEditorID   *string                    `json:"assigned_editor_id,omitempty"`
	AssignedEditorName *string                    `json:"assigned_editor_name,omitempty"`
	PublishedAt        time.Time                  `json:"published_at"`
	CreatedAt          time.Time                  `json:"created_at"`
	UpdatedAt          *time.Time                 `json:"updated_at"`
	Authors            []ManuscriptAuthorResponse `json:"authors,omitempty"`
	AuthorsSorted      []string                   `json:"authors_sorted,omitempty"`
	Files              []ManuscriptFileResponse   `json:"files,omitempty"`
	File               *ManuscriptFileResponse    `json:"file,omitempty"`
}

// ManuscriptSubmitterBrief is the account that submitted the manuscript (notifikasi & ownership).
type ManuscriptSubmitterBrief struct {
	UserID      string  `json:"user_id"`
	Email       string  `json:"email"`
	Name        string  `json:"name"`
	Affiliation *string `json:"affiliation,omitempty"`
}

type ManuscriptAuthorResponse struct {
	ID               string  `json:"id"`
	UserID           *string `json:"user_id,omitempty"`
	AuthorName       string  `json:"author_name"`
	AuthorEmail      string  `json:"author_email"`
	Affiliation      string  `json:"affiliation"`
	OrderPosition    int     `json:"order_position"`
	IsPrimaryContact bool    `json:"is_primary_contact"`
	IsPrimaryAuthor  bool    `json:"is_primary_author"`
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
