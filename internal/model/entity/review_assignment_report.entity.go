package entity

import (
	"time"

	"github.com/api-monolith-template/internal/constant"
)

// ReviewAssignmentReport stores structured reviewer answers (draft or submitted) for one assignment.
type ReviewAssignmentReport struct {
	ID                 string                     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ReviewAssignmentID string                     `json:"review_assignment_id" gorm:"type:uuid;not null;uniqueIndex"`
	Status             constant.ReviewReportStatus `json:"status" gorm:"type:varchar(20);not null;default:'DRAFT'"`
	SchemaVersion      int                        `json:"schema_version" gorm:"not null;default:1"`
	Payload            []byte                     `json:"payload" gorm:"type:jsonb;not null"`
	CreatedAt          time.Time                  `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt          *time.Time                 `json:"updated_at" gorm:"type:timestamp"`

	ReviewAssignment *ReviewAssignment `json:"review_assignment,omitempty" gorm:"foreignKey:ReviewAssignmentID;references:ID"`
}

func (ReviewAssignmentReport) TableName() string {
	return "review_assignment_reports"
}
