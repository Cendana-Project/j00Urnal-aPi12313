package entity

import (
	"time"

	"github.com/api-monolith-template/internal/constant"
	"gorm.io/gorm"
)

type Volume struct {
	ID        string                     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	JournalID string                     `json:"journal_id" gorm:"type:uuid;not null;uniqueIndex:idx_vol_journal_year_num"`
	Year      int                        `json:"year" gorm:"type:int;not null;uniqueIndex:idx_vol_journal_year_num"`
	Number    int                        `json:"number" gorm:"type:int;not null;uniqueIndex:idx_vol_journal_year_num"`
	Status    constant.PublicationStatus `json:"status" gorm:"type:varchar(20);not null;default:'DRAFT'"`
	CreatedAt time.Time                  `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt *time.Time                 `json:"updated_at" gorm:"type:timestamp"`
	DeletedAt gorm.DeletedAt             `json:"-" gorm:"index;type:timestamp"`

	// Relationships
	Journal *Journal `json:"journal,omitempty" gorm:"foreignKey:JournalID;references:ID"`
	Issues  []Issue  `json:"issues,omitempty" gorm:"foreignKey:VolumeID;references:ID"`
}
