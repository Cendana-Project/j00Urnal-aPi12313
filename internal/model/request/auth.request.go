package request

type RegisterRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Phone       string `json:"phone"`
	Password    string `json:"password"`
	Affiliation string `json:"affiliation"`
}

// OTP/PIN
type VerifyPINRequest struct {
	Email string `json:"email"`
	PIN   string `json:"pin"`
}
type ResendPINRequest struct {
	Email string `json:"email"`
}

// Login & Refresh (public)
type LoginRequest struct {
	Identity string `json:"identity"` // email atau username
	Password string `json:"password"`
}
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Password
type PasswordForgotRequest struct {
	Email string `json:"email"`
}

type PasswordResetRequest struct {
	Email       string `json:"email"`
	PIN         string `json:"pin"`
	NewPassword string `json:"new_password"`
}

type PasswordChangeRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"` // optional
}
