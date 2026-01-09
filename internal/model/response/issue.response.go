package response

import (
	"time"
)

type IssueResponse struct {
	ID              string     `json:"id"`
	VolumeID        string     `json:"volume_id"`
	Number          int        `json:"number"`
	PublicationDate string     `json:"publication_date"`
	Status          string     `json:"status"`
	CoverPath       *string    `json:"cover_path"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}
