package request

import "github.com/google/uuid"

type CreatePermissionReq struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
}

type GetAllPermissionsReq struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"page_size" binding:"required,min=1,max=100"`
}

type GetPermissionByIDReq struct {
	ID uuid.UUID `uri:"id" binding:"required"`
}

type UpdatePermissionReq struct {
	ID          uuid.UUID `json:"-"`
	Name        string    `json:"name" binding:"required"`
	Slug        string    `json:"slug" binding:"required"`
	Description string    `json:"description"`
}

type DeletePermissionReq struct {
	ID uuid.UUID `uri:"id" binding:"required"`
}

type AssignPermissionToRoleReq struct {
	RoleID       uuid.UUID `json:"role_id" binding:"required"`
	PermissionID uuid.UUID `json:"permission_id" binding:"required"`
}

type RemovePermissionFromRoleReq struct {
	RoleID       uuid.UUID `json:"role_id" binding:"required"`
	PermissionID uuid.UUID `json:"permission_id" binding:"required"`
}

type GetRolePermissionsReq struct {
	RoleID   uuid.UUID `uri:"role_id" binding:"required"`
	Page     int       `form:"page" binding:"required,min=1"`
	PageSize int       `form:"page_size" binding:"required,min=1,max=100"`
}
