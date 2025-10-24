package seeder

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
)

// SeedRoles memastikan 7 role tersedia (idempotent).
func SeedRoles(db *gorm.DB) error {
	now := time.Now()
	roles := []entity.Role{
		{ID: uuid.NewString(), Name: "Super Admin", Slug: constant.RoleSuperAdmin, Active: true, CreatedAt: now},
		{ID: uuid.NewString(), Name: "Admin (Hospital)", Slug: constant.RoleAdmin, Active: true, CreatedAt: now},
		{ID: uuid.NewString(), Name: "Patient", Slug: constant.RolePatient, Active: true, CreatedAt: now},
		{ID: uuid.NewString(), Name: "Doctor", Slug: constant.RoleDoctor, Active: true, CreatedAt: now},
		{ID: uuid.NewString(), Name: "Nurse", Slug: constant.RoleNurse, Active: true, CreatedAt: now},
		{ID: uuid.NewString(), Name: "Receptionist", Slug: constant.RoleReceptionist, Active: true, CreatedAt: now},
		{ID: uuid.NewString(), Name: "BOD", Slug: constant.RoleBOD, Active: true, CreatedAt: now},
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
