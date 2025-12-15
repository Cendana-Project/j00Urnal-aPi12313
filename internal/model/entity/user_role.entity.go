package entity

import "time"

type UserRole struct {
	UserID     string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	RoleID     string    `gorm:"type:uuid;primaryKey" json:"role_id"`
	AssignedAt time.Time `gorm:"not null;default:now()" json:"assigned_at"`
	CreatedBy  *string   `gorm:"type:uuid" json:"created_by,omitempty"`
}
