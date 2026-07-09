package entity

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
	ID          string         `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Slug        string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	Description string         `gorm:"type:text"`
	IsActive    bool           `gorm:"type:boolean;not null;default:true"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;not null;default:now()"`
	DeletedAt   gorm.DeletedAt

	Roles []Role `gorm:"many2many:role_permissions;"`
}
