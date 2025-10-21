package entity

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID          string         `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string         `json:"name" gorm:"type:varchar(64);not null;unique"`
	Slug        string         `json:"slug" gorm:"type:varchar(64);not null;uniqueIndex"`
	Description *string        `json:"description"`
	Active      bool           `json:"active" gorm:"not null;default:true"`
	CreatedAt   time.Time      `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt   *time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}
