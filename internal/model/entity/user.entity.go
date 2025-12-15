package entity

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username     string     `json:"username" gorm:"type:varchar;uniqueIndex;not null"`
	FirstName    *string    `json:"first_name" gorm:"type:varchar"`
	LastName     *string    `json:"last_name" gorm:"type:varchar"`
	Email        string     `json:"email" gorm:"type:varchar;uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"type:varchar;not null"`
	Affiliation  *string    `json:"affiliation" gorm:"type:varchar(300)"`
	Phone        *string    `json:"phone" gorm:"type:varchar"`
	LastLogin    *time.Time `json:"last_login" gorm:"type:timestamp"`

	// Auth flow fields (not in ERD but needed for registration/verification)
	Status     string     `json:"status" gorm:"type:varchar(16);not null;default:'pending'"`
	VerifiedAt *time.Time `json:"verified_at" gorm:"type:timestamp"`

	UpdatedAt *time.Time     `json:"updated_at" gorm:"type:timestamp"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;type:timestamp"`
	CreatedAt time.Time      `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
}
