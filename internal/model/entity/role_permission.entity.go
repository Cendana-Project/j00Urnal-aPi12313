package entity

import "time"

type RolePermission struct {
	ID           string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RoleID       string    `json:"role_id" gorm:"type:uuid;not null;index:idx_rp_unique,unique"`
	PermissionID string    `json:"permission_id" gorm:"type:uuid;not null;index:idx_rp_unique,unique"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
}
