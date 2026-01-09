package entity

import (
	"time"

	"github.com/api-monolith-template/internal/constant"
)

type PublicationFile struct {
	ID         string              `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	EntityType constant.EntityType `json:"entity_type" gorm:"type:varchar(20);not null"`
	EntityID   string              `json:"entity_id" gorm:"type:uuid;not null"`
	FileType   constant.FileType   `json:"file_type" gorm:"type:varchar(20);not null"`
	FilePath   string              `json:"file_path" gorm:"type:text;not null"`
	Filename   string              `json:"filename" gorm:"type:varchar(255);not null"`
	MimeType   string              `json:"mime_type" gorm:"type:varchar(100);not null"`
	SizeBytes  int64               `json:"size_bytes" gorm:"type:bigint;not null"`
	UploadedBy string              `json:"uploaded_by" gorm:"type:uuid;not null"`
	UploadedAt time.Time           `json:"uploaded_at" gorm:"type:timestamp;not null;default:now()"`

	// Relationships
	Uploader *User `json:"uploader,omitempty" gorm:"foreignKey:UploadedBy;references:ID"`
}
