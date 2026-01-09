package entity

import (
	"time"

	"github.com/api-monolith-template/internal/constant"
	"gorm.io/gorm"
)

type Manuscript struct {
	ID           string                     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	IssueID      string                     `json:"issue_id" gorm:"type:uuid;not null;column:volume_number_id"`
	Title        string                     `json:"title" gorm:"type:text;not null"`
	Abstract     string                     `json:"abstract" gorm:"type:text;not null"`
	Status       constant.PublicationStatus `json:"status" gorm:"type:varchar(20);not null;default:'PUBLISHED'"`
	MainAuthorID string                     `json:"main_author_id" gorm:"type:uuid;not null"`
	PublishedAt  time.Time                  `json:"published_at" gorm:"type:timestamp;not null;default:now()"`
	CreatedAt    time.Time                  `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt    *time.Time                 `json:"updated_at" gorm:"type:timestamp"`
	DeletedAt    gorm.DeletedAt             `json:"-" gorm:"index;type:timestamp"`

	// Relationships
	Issue      *Issue             `json:"issue,omitempty" gorm:"foreignKey:IssueID;references:ID"`
	MainAuthor *User              `json:"main_author,omitempty" gorm:"foreignKey:MainAuthorID;references:ID"`
	Authors    []ManuscriptAuthor `json:"authors,omitempty" gorm:"foreignKey:ManuscriptID;references:ID"`
	Files      []ManuscriptFile   `json:"files,omitempty" gorm:"foreignKey:ManuscriptID;references:ID"`
}
