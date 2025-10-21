package request

type RegisterRequest struct {
	Email       string  `json:"email" validate:"required,email"`
	Password    string  `json:"password" validate:"required,min=8"` // validasi kombo di service (huruf+angka)
	FirstName   string  `json:"first_name" validate:"required"`
	LastName    string  `json:"last_name" validate:"required"`
	Phone       *string `json:"phone" validate:"omitempty,numeric"`
	DOB         *string `json:"dob" validate:"omitempty"` // "YYYY-MM-DD"
	Address     *string `json:"address" validate:"omitempty"`
	Gender      *string `json:"gender" validate:"omitempty,oneof=L P"`
	NIK         *string `json:"nik" validate:"omitempty,len=16,numeric"`
	AccountRole string  `json:"account_role" validate:"required,oneof=patient doctor"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type VerifyPINRequest struct {
	Email string `json:"email" binding:"required,email"`
	PIN   string `json:"pin"   binding:"required,len=6"`
}
