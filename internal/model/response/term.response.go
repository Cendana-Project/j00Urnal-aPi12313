package response

import "time"

type TermResponse struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Version   int       `json:"version"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
