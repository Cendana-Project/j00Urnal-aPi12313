package user

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"strings"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/contract/repository"
	"github.com/api-monolith-template/internal/model/contract/service"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/util"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Service struct {
	userRepository  repository.UserContract
	cacheRepository repository.CacheContract
}

var _ service.UserContract = (*Service)(nil)

func NewService() *Service {
	return new(Service)
}

func (s *Service) WithUserRepository(repo repository.UserContract) *Service {
	s.userRepository = repo
	return s
}

func (s *Service) WithCacheRepository(repo repository.CacheContract) *Service {
	s.cacheRepository = repo
	return s
}

func (s *Service) GetByIdentifier(ctx context.Context, identifier string) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx)
	logger.Info("find by identifier process started")

	user, err := s.userRepository.GetByIdentifier(ctx, identifier)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if user == nil {
		logger.Warn("user not found with provided identifier")
		resp := constant.ErrUserNotFound.ToResponse()
		return &resp, nil
	}

	mapped := response.MapUser(*user)

	logger.Info("get user by identifier completed successfully")
	return &response.BaseResponse{
		Message: response.MessageOK,
		Data:    mapped,
	}, nil
}

func (s *Service) GetByIdentifierSimple(ctx context.Context, identifier string) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx)
	logger.Info("get user by identifier simple started")

	user, err := s.userRepository.GetByIdentifierByRole(ctx, identifier, constant.RoleClientSlug)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if user == nil {
		resp := constant.ErrUserNotFound.ToResponse()
		return &resp, nil
	}

	mapped := response.MapUserToSimple(*user)
	return &response.BaseResponse{
		Message: response.MessageOK,
		Data:    mapped,
	}, nil
}

func (s *Service) Upsert(ctx context.Context, user *entity.User) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx)
	logger.Info("upsert user process started")

	err := s.userRepository.Upsert(ctx, user)
	if err != nil {
		logger.Error(err)

		if strings.Contains(err.Error(), "duplicated key not allowed") {
			resp := constant.ErrDuplicateUsernameOrEmail.ToResponse()
			return &resp, nil
		}

		return nil, err
	}

	logger.Info("upsert user completed successfully")
	return &response.BaseResponse{
		Message: response.MessageOK,
		Data:    user,
	}, nil
}

func (s *Service) GetAllUsers(ctx context.Context) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx)
	logger.Info("get all users process started")

	users, err := s.userRepository.GetAllUsers(ctx)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(users) == 0 {
		logger.Info("no users found")
	} else {
		logger.WithField("count", len(users)).Info("found users")
	}

	var userResponses []response.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, response.MapUser(*user))
	}

	logger.Info("get all users completed successfully")
	return &response.BaseResponse{
		Message: response.MessageOK,
		Data:    userResponses,
	}, nil
}

func (s *Service) GetAllUsersSimple(ctx context.Context) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx)
	logger.Info("get all users minimal process started")

	users, err := s.userRepository.GetAllUsersByRole(ctx, constant.RoleClientSlug)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(users) == 0 {
		logger.Info("no client users found")
	} else {
		logger.WithField("count", len(users)).Info("found client users")
	}

	var userSimpleResponses []response.UserSimpleResponse
	for _, user := range users {
		userSimpleResponses = append(userSimpleResponses, response.MapUserToSimple(*user))
	}

	responseData := response.UserSimpleListResponse{
		Users: userSimpleResponses,
		Total: int64(len(userSimpleResponses)),
	}

	logger.Info("get all users simple completed successfully")
	return &response.BaseResponse{
		Message: response.MessageOK,
		Data:    responseData,
	}, nil
}

func (s *Service) Delete(ctx context.Context, userID string) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx)
	logger.Info("delete user process started")

	err := s.userRepository.Delete(ctx, userID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Info("delete user completed successfully")
	return &response.BaseResponse{
		Message: "User deleted successfully",
	}, nil
}

func (s *Service) GetByIDs(ctx context.Context, ids []uuid.UUID) (*response.BaseResponse, error) {
	logger := util.NewDefaultLogger(ctx)
	logger.WithField("count", len(ids)).Info("get users by IDs started")

	if len(ids) == 0 {
		return &response.BaseResponse{
			Message: response.MessageOK,
			Data:    map[uuid.UUID]response.UserResponse{},
		}, nil
	}

	users, err := s.userRepository.GetByIDs(ctx, ids) // <=== repo method baru
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	out := make(map[uuid.UUID]response.UserResponse, len(users))
	for _, u := range users {
		if u == nil {
			continue
		}
		out[u.ID] = response.MapUser(*u)
	}

	logger.Info("get users by IDs completed successfully")
	return &response.BaseResponse{
		Message: response.MessageOK,
		Data:    out,
	}, nil
}
