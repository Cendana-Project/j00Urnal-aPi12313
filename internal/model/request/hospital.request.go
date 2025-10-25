package request

type CreateHospitalRequest struct {
	Code        string   `json:"code" binding:"required,uppercase,alphanumdash,min=3,max=40"`
	Name        string   `json:"name" binding:"required,max=160"`
	Address     string   `json:"address" binding:"required"`
	City        string   `json:"city" binding:"required,max=100"`
	Province    string   `json:"province" binding:"required,max=100"`
	Country     string   `json:"country" binding:"omitempty,max=100"`
	Latitude    *float64 `json:"latitude" binding:"omitempty"`
	Longitude   *float64 `json:"longitude" binding:"omitempty"`
	Phone       string   `json:"phone" binding:"required,max=50"`
	Description string   `json:"description" binding:"omitempty,max=200"`
	Facilities  any      `json:"facilities" binding:"omitempty"` // JSON (obj/array)
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

type CreateHospitalAdminRequest struct {
	HospitalID string  `json:"hospital_id" binding:"required"`
	Email      string  `json:"email"       binding:"required,email,unique_db=users:email"`
	Username   string  `json:"username"    binding:"required,min=3,max=64,unique_db=users:username"`
	Phone      *string `json:"phone"       binding:"omitempty"`
	Password   string  `json:"password"    binding:"required,validate_password"`
	FirstName  *string `json:"first_name"  binding:"omitempty,max=100"`
	LastName   *string `json:"last_name"   binding:"omitempty,max=100"`
	DOB        *string `json:"dob"         binding:"omitempty,datetime=2006-01-02"`
	Address    *string `json:"address"     binding:"omitempty"`
	Gender     *string `json:"gender"      binding:"omitempty,oneof=L P"`
	NIK        *string `json:"nik"         binding:"omitempty,len=16,numeric"`
}

type CreateHospitalStaffRequest struct {
	HospitalID string  `json:"hospital_id" binding:"required"`
	Role       string  `json:"role"       binding:"required,oneof=doctor nurse receptionist bod admin"`
	Email      string  `json:"email"      binding:"required,email,unique_db=users:email"`
	Username   string  `json:"username"   binding:"required,min=3,max=64,unique_db=users:username"`
	Phone      *string `json:"phone"      binding:"omitempty"`
	Password   string  `json:"password"   binding:"required,validate_password"`
	FirstName  *string `json:"first_name" binding:"omitempty,max=100"`
	LastName   *string `json:"last_name"  binding:"omitempty,max=100"`
	DOB        *string `json:"dob"        binding:"omitempty,datetime=2006-01-02"`
	Address    *string `json:"address"    binding:"omitempty"`
	Gender     *string `json:"gender"     binding:"omitempty,oneof=L P"`
	NIK        *string `json:"nik"        binding:"omitempty,len=16,numeric"`
}
