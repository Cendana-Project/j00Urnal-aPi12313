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

type RoleBrief struct { // <=== added
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type LoginResponse struct { // <=== changed
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	Role                  string `json:"role"`                  // hanya slug
	AccessTokenExpiredAt  string `json:"accessTokenExpiredAt"`  // RFC3339 UTC
	RefreshTokenExpiredAt string `json:"refreshTokenExpiredAt"` // RFC3339 UTC
}

// LoginHospitalResponse sekarang menyertakan waktu kadaluarsa token
type LoginHospitalResponse struct { // <=== changed
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	ExpiresIn             int64  `json:"expires_in"`
	TokenType             string `json:"token_type"`
	HospitalID            string `json:"hospital_id"`
	Role                  string `json:"role"`                  // hanya slug
	AccessTokenExpiredAt  string `json:"accessTokenExpiredAt"`  // RFC3339 UTC
	RefreshTokenExpiredAt string `json:"refreshTokenExpiredAt"` // RFC3339 UTC
}
