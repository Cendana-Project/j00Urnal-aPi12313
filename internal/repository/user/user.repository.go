package user

import (
	"errors"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(u *entity.User) error {
	return r.db.Create(u).Error
}

func (r *Repository) FindByEmail(email string) (*entity.User, error) {
	var u entity.User
	err := r.db.Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *Repository) MarkVerified(userID string) error {
	return r.db.Model(&entity.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"status":      "active",
			"verified_at": gorm.Expr("NOW()"),
		}).Error
}
