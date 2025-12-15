package response

import "time"

// MeResponse is the response for GET /v1/me
type MeResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	FirstName   *string    `json:"first_name"`
	LastName    *string    `json:"last_name"`
	Affiliation *string    `json:"affiliation,omitempty"`
	Phone       *string    `json:"phone,omitempty"`
	Status      string     `json:"status"`
	VerifiedAt  *time.Time `json:"verified_at,omitempty"`
	LastLogin   *time.Time `json:"last_login,omitempty"`
	Role        string     `json:"role"` // single role slug
}
