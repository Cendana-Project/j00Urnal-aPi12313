package entity

import "time"

type UserRole struct {
	UserID     string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	RoleID     string    `gorm:"type:uuid;primaryKey" json:"role_id"`
	HospitalID *string   `gorm:"type:uuid;default:null" json:"hospital_id"` // <=== added (NULL = global)
	AssignedAt time.Time `gorm:"not null;default:now()" json:"assigned_at"`
	CreatedBy  *string   `gorm:"type:uuid" json:"created_by,omitempty"`
}
