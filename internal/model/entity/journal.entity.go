package entity

import (
	"time"

	"github.com/api-monolith-template/internal/constant"
	"gorm.io/gorm"
)

type Journal struct {
	ID          string                     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string                     `json:"name" gorm:"type:varchar;not null"`
	Description string                     `json:"description" gorm:"type:text"`
	Status      constant.PublicationStatus `json:"status" gorm:"type:varchar(20);not null;default:'DRAFT'"`
	CoverPath   *string                    `json:"cover_path" gorm:"type:text"`
	CreatedBy   string                     `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time                  `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt   *time.Time                 `json:"updated_at" gorm:"type:timestamp"`
	DeletedAt   gorm.DeletedAt             `json:"-" gorm:"index;type:timestamp"`

	// Relationships
	Creator *User    `json:"creator,omitempty" gorm:"foreignKey:CreatedBy;references:ID"`
	Volumes []Volume `json:"volumes,omitempty" gorm:"foreignKey:JournalID;references:ID"`
}
