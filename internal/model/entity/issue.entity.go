package entity

import (
	"time"

	"github.com/api-monolith-template/internal/constant"
	"gorm.io/gorm"
)

type Issue struct {
	ID              string                     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	VolumeID        string                     `json:"volume_id" gorm:"type:uuid;not null"`
	Number          int                        `json:"number" gorm:"type:int;not null"`
	PublicationDate time.Time                  `json:"publication_date" gorm:"type:date;not null"`
	Status          constant.PublicationStatus `json:"status" gorm:"type:varchar(20);not null;default:'DRAFT'"`
	CoverPath       *string                    `json:"cover_path" gorm:"type:text"`
	CreatedAt       time.Time                  `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt       *time.Time                 `json:"updated_at" gorm:"type:timestamp"`
	DeletedAt       gorm.DeletedAt             `json:"-" gorm:"index;type:timestamp"`

	// Relationships
	Volume      *Volume      `json:"volume,omitempty" gorm:"foreignKey:VolumeID;references:ID"`
	Manuscripts []Manuscript `json:"manuscripts,omitempty" gorm:"foreignKey:IssueID;references:ID"`
}
