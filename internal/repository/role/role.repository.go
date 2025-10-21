package role

import (
	"github.com/api-monolith-template/internal/model/entity"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) FindBySlug(slug string) (*entity.Role, error) {
	var m entity.Role
	if err := r.db.Where("slug = ?", slug).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) Assign(userID, roleID string) error {
	ur := entity.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.Create(&ur).Error
}
