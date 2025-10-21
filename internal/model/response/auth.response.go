package response

type RegisterResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

type VerifyEmailResponse struct {
	Email  string `json:"email"`
	Status string `json:"status"`
}
