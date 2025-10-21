package service

import (
	"context"

	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
)

type PermissionContract interface {
	// Permission management
	CreatePermission(ctx context.Context, req *request.CreatePermissionReq) (*response.BaseResponse, error)
	GetAllPermissions(ctx context.Context, req *request.GetAllPermissionsReq) (*response.BaseResponse, error)
	GetPermissionByID(ctx context.Context, req *request.GetPermissionByIDReq) (*response.BaseResponse, error)
	UpdatePermission(ctx context.Context, req *request.UpdatePermissionReq) (*response.BaseResponse, error)
	DeletePermission(ctx context.Context, req *request.DeletePermissionReq) (*response.BaseResponse, error)

	// Role permission management
	AssignPermissionToRole(ctx context.Context, req *request.AssignPermissionToRoleReq) (*response.BaseResponse, error)
	RemovePermissionFromRole(ctx context.Context, req *request.RemovePermissionFromRoleReq) (*response.BaseResponse, error)
	GetRolePermissions(ctx context.Context, req *request.GetRolePermissionsReq) (*response.BaseResponse, error)
}
