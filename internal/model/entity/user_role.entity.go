package entity

import "time"

type UserRole struct {
	ID        string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	RoleID    string    `json:"role_id" gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
}
