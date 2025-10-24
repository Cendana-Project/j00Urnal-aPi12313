package request

import "github.com/google/uuid"

type CreateRoleReq struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description" binding:"required"`
	IsActive    bool   `json:"is_active"`
}

func NewCreateRoleReq() *CreateRoleReq {
	return &CreateRoleReq{
		IsActive: true, //default role status is active
	}
}

type AssignRoleToUserReq struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}

type RemoveRoleFromUserReq struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}

type GetAllRoles struct {
	Page     int `form:"page" binding:"required"`
	PageSize int `form:"page_size" binding:"required"`
}

type UpdateUserRoleReq struct {
	UserID    uuid.UUID `json:"user_id" binding:"required"`
	OldRoleID uuid.UUID `json:"old_role_id" binding:"required"`
	NewRoleID uuid.UUID `json:"new_role_id" binding:"required"`
}

type GetRoleByIdReq struct {
	RoleID uuid.UUID `uri:"id" binding:"required"`
}

type UpdateRoleReq struct {
	RoleID      uuid.UUID `uri:"id" swaggerignore:"true"`
	Name        string    `json:"name" binding:"required"`
	Slug        string    `json:"slug" binding:"required"`
	Description string    `json:"description" binding:"required"`
	IsActive    bool      `json:"is_active"`
}

type DeleteRoleReq struct {
	RoleID uuid.UUID `uri:"id" binding:"required"`
}

type GetAllUserByRolesReq struct {
	RoleID   uuid.UUID `json:"role_id" binding:"required"`
	Page     int       `json:"page" binding:"required"`
	PageSize int       `json:"page_size" binding:"required"`
}
