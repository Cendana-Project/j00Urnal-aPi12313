package service

import (
	"context"
	"github.com/google/uuid"

	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
)

type UserContract interface {
	GetByIdentifier(ctx context.Context, identifier string) (*response.BaseResponse, error)
	Upsert(ctx context.Context, user *entity.User) (*response.BaseResponse, error)
	GetAllUsers(ctx context.Context) (*response.BaseResponse, error)
	GetAllUsersSimple(ctx context.Context) (*response.BaseResponse, error)
	GetByIdentifierSimple(ctx context.Context, identifier string) (*response.BaseResponse, error)
	Delete(ctx context.Context, userID string) (*response.BaseResponse, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) (*response.BaseResponse, error) // Data: map[uuid.UUID]response.UserResponse
}
