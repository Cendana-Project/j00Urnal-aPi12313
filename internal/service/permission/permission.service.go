package permission

import (
	"context"

	"github.com/api-monolith-template/internal/model/contract/repository"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
)

type Service struct {
	permissionRepo repository.PermissionContract
}

func NewService(permissionRepo repository.PermissionContract) *Service {
	return &Service{
		permissionRepo: permissionRepo,
	}
}

func (s *Service) CreatePermission(ctx context.Context, req *request.CreatePermissionReq) (*response.BaseResponse, error) {
	return s.permissionRepo.CreatePermission(ctx, req)
}

func (s *Service) GetAllPermissions(ctx context.Context, req *request.GetAllPermissionsReq) (*response.BaseResponse, error) {
	return s.permissionRepo.GetAllPermissions(ctx, req)
}

func (s *Service) GetPermissionByID(ctx context.Context, req *request.GetPermissionByIDReq) (*response.BaseResponse, error) {
	return s.permissionRepo.GetPermissionByID(ctx, req)
}

func (s *Service) UpdatePermission(ctx context.Context, req *request.UpdatePermissionReq) (*response.BaseResponse, error) {
	return s.permissionRepo.UpdatePermission(ctx, req)
}

func (s *Service) DeletePermission(ctx context.Context, req *request.DeletePermissionReq) (*response.BaseResponse, error) {
	return s.permissionRepo.DeletePermission(ctx, req)
}

func (s *Service) AssignPermissionToRole(ctx context.Context, req *request.AssignPermissionToRoleReq) (*response.BaseResponse, error) {
	return s.permissionRepo.AssignPermissionToRole(ctx, req)
}

func (s *Service) RemovePermissionFromRole(ctx context.Context, req *request.RemovePermissionFromRoleReq) (*response.BaseResponse, error) {
	return s.permissionRepo.RemovePermissionFromRole(ctx, req)
}

func (s *Service) GetRolePermissions(ctx context.Context, req *request.GetRolePermissionsReq) (*response.BaseResponse, error) {
	return s.permissionRepo.GetRolePermissions(ctx, req)
}
