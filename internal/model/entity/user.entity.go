package entity

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           string         `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string         `json:"email" gorm:"type:varchar(190);uniqueIndex;not null"`
	FirstName    string         `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName     string         `json:"last_name" gorm:"type:varchar(100);not null"`
	Phone        *string        `json:"phone" gorm:"type:varchar(32)"`
	DOB          *time.Time     `json:"dob"`
	Address      *string        `json:"address" gorm:"type:text"`
	Gender       *string        `json:"gender" gorm:"type:varchar(1)"` // L | P
	NIK          *string        `json:"nik" gorm:"type:varchar(16);uniqueIndex"`
	PasswordHash string         `json:"-" gorm:"type:text;not null"`
	Status       string         `json:"status" gorm:"type:varchar(16);not null;default:'pending'"`
	VerifiedAt   *time.Time     `json:"verified_at"`
	CreatedAt    time.Time      `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt    *time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
