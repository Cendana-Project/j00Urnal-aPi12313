package response

import (
	"time"
)

type JournalResponse struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Status      string           `json:"status"`
	CoverPath   *string          `json:"cover_path"`
	CreatedBy   string           `json:"created_by"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   *time.Time       `json:"updated_at,omitempty"`
	Volumes     []VolumeResponse `json:"volumes,omitempty"`
}
