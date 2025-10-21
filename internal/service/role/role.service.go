package role

import (
	"context"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/contract/service"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/util"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service struct {
	roleRepository service.RoleContract
}

func NewService() *Service {
	return new(Service)
}

func (s *Service) WithRoleRepository(repo service.RoleContract) *Service {
	s.roleRepository = repo
	return s
}

func (s *Service) AssignRoleToUser(ctx context.Context, req *request.AssignRoleToUserReq) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx).WithFields(logrus.Fields{
		"req": util.Dump(req),
	})

	resp, err := s.roleRepository.AssignRoleToUser(ctx, req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return resp, nil
}

func (s *Service) CreateNewRole(ctx context.Context, req *request.CreateRoleReq) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx).WithFields(logrus.Fields{
		"req": util.Dump(req),
	})

	resp, err := s.roleRepository.CreateNewRole(ctx, req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return resp, nil
}

func (s *Service) DeleteRole(ctx context.Context, req *request.DeleteRoleReq) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx).WithFields(logrus.Fields{
		"req": util.Dump(req),
	})
	logger.Info("delete role process started")

	if req.RoleID == uuid.Nil {
		logger.Error("invalid role ID (nil UUID)")
		return nil, constant.ErrInvalidRoleID
	}

	resp, err := s.roleRepository.DeleteRole(ctx, req)
	if err != nil {
		logger.WithError(err).Error("failed to delete role")
		return nil, err
	}

	logger.Info("role deleted successfully")
	return resp, nil
}

func (s *Service) GetAllRoles(ctx context.Context, req *request.GetAllRoles) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx)
	logger.Info("get all roles process started")

	resp, err := s.roleRepository.GetAllRoles(ctx, req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Info("get all roles completed successfully")
	return resp, nil
}

func (s *Service) GetRoleById(ctx context.Context, req *request.GetRoleByIdReq) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx).WithFields(logrus.Fields{
		"req": util.Dump(req),
	})
	logger.Info("get role by id process started")

	resp, err := s.roleRepository.GetRoleById(ctx, req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Info("get role by id completed successfully")
	return resp, nil
}

func (s *Service) RemoveRoleFromUser(ctx context.Context, req *request.RemoveRoleFromUserReq) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx).WithFields(logrus.Fields{
		"req": util.Dump(req),
	})

	resp, err := s.roleRepository.RemoveRoleFromUser(ctx, req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return resp, nil
}

func (s *Service) UpdateRole(ctx context.Context, req *request.UpdateRoleReq) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx).WithFields(logrus.Fields{
		"req": util.Dump(req),
	})
	logger.Info("update role process started")

	resp, err := s.roleRepository.UpdateRole(ctx, req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Info("role updated successfully")
	return resp, nil
}

func (s *Service) GetAllUserByRoles(ctx context.Context, req *request.GetAllUserByRolesReq) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx).WithFields(logrus.Fields{
		"req": util.Dump(req),
	})
	logger.Info("get all user by roles process started")

	resp, err := s.roleRepository.GetAllUserByRoles(ctx, req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Info("get all user by roles completed successfully")
	return resp, nil
}

func (s *Service) UpdateUserRole(ctx context.Context, req *request.UpdateUserRoleReq) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx).WithFields(logrus.Fields{
		"req": util.Dump(req),
	})
	logger.Info("update user role process started")

	resp, err := s.roleRepository.UpdateUserRole(ctx, req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Info("user role updated successfully")
	return resp, nil
}
