package entity

import (
	"time"

	"github.com/api-monolith-template/internal/constant"
)

type ReviewAssignment struct {
	ID                   string                           `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ReviewRoundID        string                           `json:"review_round_id" gorm:"type:uuid;not null"`
	ReviewerID           string                           `json:"reviewer_id" gorm:"type:uuid;not null"`
	AssignedBy           string                           `json:"assigned_by" gorm:"type:uuid;not null"`
	Status               constant.ReviewAssignmentStatus  `json:"status" gorm:"type:varchar(20);not null;default:'INVITED'"`
	InvitationToken      *string                          `json:"invitation_token" gorm:"type:varchar(128);uniqueIndex"`
	InvitationExpiresAt  time.Time                        `json:"invitation_expires_at" gorm:"type:timestamp;not null"`
	InvitationAcceptedAt *time.Time                       `json:"invitation_accepted_at" gorm:"type:timestamp"`
	DueDate              time.Time                        `json:"due_date" gorm:"type:timestamp;not null"`
	Recommendation       *string                          `json:"recommendation" gorm:"type:varchar(20)"`
	Comments             *string                          `json:"comments" gorm:"type:text"`
	CompletedAt          *time.Time                       `json:"completed_at" gorm:"type:timestamp"`
	CreatedAt            time.Time                        `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt            *time.Time                       `json:"updated_at" gorm:"type:timestamp"`

	// Relationships
	ReviewRound *ReviewRound `json:"review_round,omitempty" gorm:"foreignKey:ReviewRoundID;references:ID"`
	Reviewer    *User        `json:"reviewer,omitempty" gorm:"foreignKey:ReviewerID;references:ID"`
	Assigner    *User        `json:"assigner,omitempty" gorm:"foreignKey:AssignedBy;references:ID"`
	Files       []ReviewFile `json:"files,omitempty" gorm:"foreignKey:ReviewAssignmentID;references:ID"`
}

func (ReviewAssignment) TableName() string {
	return "review_assignments"
}
