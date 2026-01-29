package entity

import "time"

type PublicationTerm struct {
	ID        string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Version   int       `json:"version" gorm:"type:int;not null;default:1"`
	IsActive  bool      `json:"is_active" gorm:"type:boolean;not null;default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
}
