package service

import (
	"context"

	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
)

type RoleContract interface {
	CreateNewRole(ctx context.Context, req *request.CreateRoleReq) (*response.BaseResponse, error)
	AssignRoleToUser(ctx context.Context, req *request.AssignRoleToUserReq) (*response.BaseResponse, error)
	RemoveRoleFromUser(ctx context.Context, req *request.RemoveRoleFromUserReq) (*response.BaseResponse, error)
	GetAllRoles(ctx context.Context, req *request.GetAllRoles) (*response.BaseResponse, error)
	UpdateUserRole(ctx context.Context, req *request.UpdateUserRoleReq) (*response.BaseResponse, error)
	GetRoleById(ctx context.Context, req *request.GetRoleByIdReq) (*response.BaseResponse, error)
	UpdateRole(ctx context.Context, req *request.UpdateRoleReq) (*response.BaseResponse, error)
	DeleteRole(ctx context.Context, req *request.DeleteRoleReq) (*response.BaseResponse, error)
	GetAllUserByRoles(ctx context.Context, req *request.GetAllUserByRolesReq) (*response.BaseResponse, error)
}
