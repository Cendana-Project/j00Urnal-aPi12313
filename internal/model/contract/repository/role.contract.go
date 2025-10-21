package repository

import "github.com/api-monolith-template/internal/model/entity"

type RoleRepository interface {
	FindBySlug(slug string) (*entity.Role, error)
	Assign(userID, roleID string) error
}
