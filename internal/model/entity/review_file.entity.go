package entity

import (
	"time"
)

type ReviewFile struct {
	ID                 string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ReviewAssignmentID *string   `json:"review_assignment_id" gorm:"type:uuid"`
	ReviewRoundID      string    `json:"review_round_id" gorm:"type:uuid;not null"`
	UploadedBy         string    `json:"uploaded_by" gorm:"type:uuid;not null"`
	FileType           string    `json:"file_type" gorm:"type:varchar(20);not null"`
	FilePath           string    `json:"file_path" gorm:"type:text;not null"`
	Filename           string    `json:"filename" gorm:"type:text;not null"`
	MimeType           string    `json:"mime_type" gorm:"type:varchar(100);not null"`
	SizeBytes          int64     `json:"size_bytes" gorm:"not null"`
	UploadedAt         time.Time `json:"uploaded_at" gorm:"type:timestamp;not null;default:now()"`

	// Relationships
	ReviewAssignment *ReviewAssignment `json:"review_assignment,omitempty" gorm:"foreignKey:ReviewAssignmentID;references:ID"`
	ReviewRound      *ReviewRound      `json:"review_round,omitempty" gorm:"foreignKey:ReviewRoundID;references:ID"`
	Uploader         *User             `json:"uploader,omitempty" gorm:"foreignKey:UploadedBy;references:ID"`
}

func (ReviewFile) TableName() string {
	return "review_files"
}
