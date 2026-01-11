package seeder

import (
	"time"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
)

// SeedRoles memastikan 7 role tersedia (idempotent).
func SeedRoles(db *gorm.DB) error {
	now := time.Now()
	// User requested pattern: 550e8400-e29b-41d4-a716-44665544000x
	roles := []entity.Role{
		{ID: "550e8400-e29b-41d4-a716-446655440001", Name: "Super Admin", Slug: constant.RoleSuperAdmin, Active: true, CreatedAt: now},
		{ID: "550e8400-e29b-41d4-a716-446655440002", Name: "Editor", Slug: constant.RoleEditor, Active: true, CreatedAt: now},
		{ID: "550e8400-e29b-41d4-a716-446655440003", Name: "Chief Editor", Slug: constant.RoleChiefEditor, Active: true, CreatedAt: now},
		{ID: "550e8400-e29b-41d4-a716-446655440004", Name: "Author", Slug: constant.RoleAuthor, Active: true, CreatedAt: now},
		{ID: "550e8400-e29b-41d4-a716-446655440005", Name: "Reviewer", Slug: constant.RoleReviewer, Active: true, CreatedAt: now},
	}
	for _, r := range roles {
		var cnt int64
		if err := db.Model(&entity.Role{}).Where("slug = ?", r.Slug).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt == 0 {
			if err := db.Create(&r).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
