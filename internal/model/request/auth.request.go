package request

// Step-1 register (lite)
type RegisterLiteRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=32"`
	Phone    string `json:"phone"    binding:"required,min=8,max=20"`
	Password string `json:"password" binding:"required,min=8"`
}

// OTP/PIN
type VerifyPINRequest struct {
	Email string `json:"email" binding:"required,email"`
	PIN   string `json:"pin"   binding:"required,len=6"`
}
type ResendPINRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Login & Refresh (public)
type LoginRequest struct {
	Identity string `json:"identity" binding:"required"` // email atau username
	Password string `json:"password" binding:"required"`
}
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Login — Hospital (opsi A: di body, tanpa header)
type LoginHospitalRequest struct {
	Identifier   string  `json:"identifier"    binding:"required"` // email/username/phone (sesuai implementasi service)
	Password     string  `json:"password"      binding:"required,min=8"`
	HospitalID   *string `json:"hospital_id"   binding:"omitempty,uuid4"`
	HospitalCode *string `json:"hospital_code" binding:"omitempty,max=40"` // boleh huruf/angka/dash; relaks dulu
}

// Role choose
type ChooseRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=patient doctor"`
}

// Profiles
type PatientProfileRequest struct {
	FirstName string  `json:"first_name" binding:"required"`
	LastName  string  `json:"last_name"  binding:"required"`
	NIK       *string `json:"nik"        binding:"omitempty,len=16"`
	DOB       *string `json:"dob"        binding:"omitempty"` // YYYY-MM-DD
	Address   *string `json:"address"    binding:"omitempty"`
	Gender    *string `json:"gender"     binding:"omitempty,oneof=L P"`

	HeightCM  *int    `json:"height_cm"  binding:"omitempty,gte=0,lte=300"`
	WeightKG  *int    `json:"weight_kg"  binding:"omitempty,gte=0,lte=500"`
	Allergies *string `json:"allergies"  binding:"omitempty"`
	History   *string `json:"medical_history" binding:"omitempty"`
}

type DoctorProfileRequest struct {
	FirstName string  `json:"first_name" binding:"required"`
	LastName  string  `json:"last_name"  binding:"required"`
	Address   *string `json:"address"    binding:"omitempty"`
	Gender    *string `json:"gender"     binding:"omitempty,oneof=L P"`

	SIP       string `json:"sip_number" binding:"required"`
	Specialty string `json:"specialty"  binding:"required"`
}
