package request

type CreateHospitalRequest struct {
	Code        *string  `json:"code,omitempty" validate:"omitempty,max=40"`
	Name        string   `json:"name" validate:"required,min=3,max=160"`
	Address     *string  `json:"address,omitempty" validate:"omitempty,max=1000"`
	City        *string  `json:"city,omitempty" validate:"omitempty,max=100"`
	Province    *string  `json:"province,omitempty" validate:"omitempty,max=100"`
	Country     *string  `json:"country,omitempty" validate:"omitempty,max=100"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	Phone       *string  `json:"phone,omitempty" validate:"omitempty,max=50"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=200"`
	Facilities  any      `json:"facilities,omitempty"`
}

type AssignAdminRequest struct {
	Email     string  `json:"email" validate:"required,email"`
	FirstName string  `json:"first_name" validate:"required"`
	LastName  string  `json:"last_name" validate:"required"`
	Phone     *string `json:"phone,omitempty"`
	Password  string  `json:"password" validate:"required,min=8"`
}

type TenantLoginRequest struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required"`
	HospitalCode string `json:"hospital_code" validate:"required"`
}

type RegisterStaffRequest struct {
	Email     string  `json:"email" validate:"required,email"`
	FirstName string  `json:"first_name" validate:"required"`
	LastName  string  `json:"last_name" validate:"required"`
	Role      string  `json:"role" validate:"required,oneof=doctor"` // untuk saat ini hanya doctor
	Phone     *string `json:"phone,omitempty"`
	Password  string  `json:"password" validate:"required,min=8"`
}
