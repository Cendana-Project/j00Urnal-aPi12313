package entity

import (
	"time"

	"github.com/api-monolith-template/internal/constant"
)

type ManuscriptFile struct {
	ID             string            `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ManuscriptID   string            `json:"manuscript_id" gorm:"type:uuid;not null"`
	FileType       constant.FileType `json:"file_type" gorm:"type:varchar(20);not null"`
	FilePath       string            `json:"file_path" gorm:"type:text;not null"`
	Filename       string            `json:"filename" gorm:"type:text;not null"`
	MimeType       string            `json:"mime_type" gorm:"type:varchar(100);not null"`
	SizeBytes      int64             `json:"size_bytes" gorm:"type:bigint;not null"`
	ChecksumSHA256 string            `json:"checksum_sha256" gorm:"type:varchar(64)"`
	Version        int               `json:"version" gorm:"type:int;not null;default:1"`
	UploadedAt     time.Time         `json:"uploaded_at" gorm:"type:timestamp;not null;default:now()"`
}
