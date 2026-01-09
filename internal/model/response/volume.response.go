package response

import (
	"time"
)

type VolumeResponse struct {
	ID        string     `json:"id"`
	JournalID string     `json:"journal_id"`
	Year      int        `json:"year"`
	Number    int        `json:"number"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
