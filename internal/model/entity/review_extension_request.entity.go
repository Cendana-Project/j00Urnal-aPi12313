package entity

import (
	"time"

	"github.com/api-monolith-template/internal/constant"
)

type ReviewExtensionRequest struct {
	ID                 string                               `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ReviewAssignmentID string                               `json:"review_assignment_id" gorm:"type:uuid;not null"`
	RequestedDue       time.Time                            `json:"requested_due" gorm:"type:timestamp;not null"`
	Reason             string                               `json:"reason" gorm:"type:text"`
	Status             constant.ReviewExtensionRequestStatus `json:"status" gorm:"type:varchar(20);not null;default:'PENDING'"`
	DecidedBy          *string                              `json:"decided_by,omitempty" gorm:"type:uuid"`
	DecidedAt          *time.Time                           `json:"decided_at,omitempty" gorm:"type:timestamp"`
	CreatedAt          time.Time                            `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt          *time.Time                           `json:"updated_at" gorm:"type:timestamp"`

	ReviewAssignment *ReviewAssignment `json:"review_assignment,omitempty" gorm:"foreignKey:ReviewAssignmentID;references:ID"`
}

func (ReviewExtensionRequest) TableName() string {
	return "review_extension_requests"
}
