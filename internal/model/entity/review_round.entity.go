package entity

import (
	"time"

	"github.com/api-monolith-template/internal/constant"
)

type ReviewRound struct {
	ID               string                      `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ManuscriptID     string                      `json:"manuscript_id" gorm:"type:uuid;not null"`
	RoundNumber      int                         `json:"round_number" gorm:"not null;default:1"`
	Status           constant.ReviewRoundStatus  `json:"status" gorm:"type:varchar(30);not null;default:'PENDING'"`
	EditorDecision   *string                     `json:"editor_decision" gorm:"type:varchar(30)"`
	DecisionComments *string                     `json:"decision_comments" gorm:"type:text"`
	DecisionAt       *time.Time                  `json:"decision_at" gorm:"type:timestamp"`
	CreatedBy        string                      `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt        time.Time                   `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt        *time.Time                  `json:"updated_at" gorm:"type:timestamp"`

	// Relationships
	Manuscript  *Manuscript        `json:"manuscript,omitempty" gorm:"foreignKey:ManuscriptID;references:ID"`
	Creator     *User              `json:"creator,omitempty" gorm:"foreignKey:CreatedBy;references:ID"`
	Assignments []ReviewAssignment `json:"assignments,omitempty" gorm:"foreignKey:ReviewRoundID;references:ID"`
	Files       []ReviewFile       `json:"files,omitempty" gorm:"foreignKey:ReviewRoundID;references:ID"`
}

func (ReviewRound) TableName() string {
	return "review_rounds"
}
