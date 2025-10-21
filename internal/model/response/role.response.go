package response

import "github.com/google/uuid"

type RoleResp struct {
	RoleID      uuid.UUID `json:"role_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
}

func NewRoleResp() *RoleResp {
	return &RoleResp{
		IsActive: true, //default role status is active
	}
}
