package entity

import (
	"time"

	"github.com/api-monolith-template/internal/constant"
	"gorm.io/gorm"
)

type Manuscript struct {
	ID            string                    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	JournalID     string                    `json:"journal_id" gorm:"type:uuid;not null"`
	IssueID       *string                   `json:"issue_id" gorm:"type:uuid;column:volume_number_id"` // Nullable now
	Title         string                    `json:"title" gorm:"type:text;not null"`
	Abstract      string                    `json:"abstract" gorm:"type:text;not null"`
	Status        constant.ManuscriptStatus `json:"status" gorm:"type:varchar(20);not null;default:'DRAFT'"`
	MainAuthorID  string                    `json:"main_author_id" gorm:"type:uuid;not null"`
	IsTncAccepted bool                      `json:"is_tnc_accepted" gorm:"type:boolean;default:false"`
	TncAcceptedAt *time.Time                `json:"tnc_accepted_at" gorm:"type:timestamp"`
	TermID        *string                   `json:"term_id" gorm:"type:uuid;column:term_id"` // Link to specific T&C version
	CurrentStep   int                       `json:"current_step" gorm:"type:integer;not null;default:1"`

	PublishedAt time.Time      `json:"published_at" gorm:"type:timestamp;not null;default:now()"`
	CreatedAt   time.Time      `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt   *time.Time     `json:"updated_at" gorm:"type:timestamp"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index;type:timestamp"`

	// Relationships
	Journal    *Journal           `json:"journal,omitempty" gorm:"foreignKey:JournalID;references:ID"`
	Issue      *Issue             `json:"issue,omitempty" gorm:"foreignKey:IssueID;references:ID"`
	MainAuthor *User              `json:"main_author,omitempty" gorm:"foreignKey:MainAuthorID;references:ID"`
	Term       *PublicationTerm   `json:"term,omitempty" gorm:"foreignKey:TermID;references:ID"`
	Authors    []ManuscriptAuthor `json:"authors,omitempty" gorm:"foreignKey:ManuscriptID;references:ID"`
	Files      []ManuscriptFile   `json:"files,omitempty" gorm:"foreignKey:ManuscriptID;references:ID"`
}
