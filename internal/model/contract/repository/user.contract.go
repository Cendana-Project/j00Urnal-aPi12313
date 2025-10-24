package repository

import "github.com/api-monolith-template/internal/model/entity"

type UserRepository interface {
	Create(u *entity.User) error
	FindByEmail(email string) (*entity.User, error)
	MarkVerified(userID string) error
}
