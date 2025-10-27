package request

import "encoding/json"

// Step-1 register (lite)
type RegisterLiteRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
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

// Login — Hospital (opsi A: di body, tanpa header)
type LoginHospitalRequest struct {
	Identifier   string  `json:"identifier"` // email/username/phone (sesuai implementasi service)
	Password     string  `json:"password"`
	HospitalID   *string `json:"hospital_id"`
	HospitalCode *string `json:"hospital_code"`
}

// Choose role (lama; masih dipakai di endpoint terpisah)
type ChooseRoleRequest struct {
	Role string `json:"role"`
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

// ====== SetProfile (gabungan pilih role + set profil) ======

type SetProfileRequest struct {
	Role    string           `json:"role"`    // "PATIENT" | "DOCTOR" (UPPERCASE)
	Profile *json.RawMessage `json:"profile"` // payload profil mentah; akan diparse sesuai role di service
}

// ====== Profil masing-masing role (dipakai Update* & SetProfile) ======

type PatientProfileRequest struct {
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	NIK            *string `json:"nik,omitempty"`
	DOB            *string `json:"dob,omitempty"` // YYYY-MM-DD
	Address        *string `json:"address,omitempty"`
	Gender         *string `json:"gender,omitempty"` // L|P
	HeightCM       *int    `json:"height_cm,omitempty"`
	WeightKG       *int    `json:"weight_kg,omitempty"`
	Allergies      *string `json:"allergies,omitempty"`
	MedicalHistory *string `json:"medical_history,omitempty"`
}

type DoctorProfileRequest struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Address   *string `json:"address,omitempty"`
	Gender    *string `json:"gender,omitempty"` // L|P
	SIPNumber *string `json:"sip_number,omitempty"`
	Specialty *string `json:"specialty,omitempty"`
}
