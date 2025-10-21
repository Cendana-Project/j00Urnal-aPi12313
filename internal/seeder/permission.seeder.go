package seeder

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
)

func SeedPermissions(db *gorm.DB) error {
	now := time.Now()
	for _, slug := range constant.AllPermissions() {
		var cnt int64
		if err := db.Model(&entity.Permission{}).Where("slug = ?", slug).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt == 0 {
			name := toTitle(slug)
			desc := ""
			p := entity.Permission{
				Name:        name,
				Slug:        slug,
				IsActive:    true,
				CreatedAt:   now,
				Description: desc,
			}
			if err := db.Create(&p).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func SeedRolePermissions(db *gorm.DB) error {
	type roleMap struct {
		RoleSlug string
		Perms    []string
	}
	mapping := []roleMap{
		{constant.RoleSuperAdmin, constant.PermissionsSuperAdmin}, // all perms
		{constant.RoleAdmin, constant.PermissionsAdmin},
		{constant.RolePatient, constant.PermissionsPatient},
		{constant.RoleDoctor, constant.PermissionsDoctor},
		{constant.RoleNurse, constant.PermissionsNurse},
		{constant.RoleReceptionist, constant.PermissionsReceptionist},
		{constant.RoleBOD, constant.PermissionsBOD},
	}

	for _, m := range mapping {
		var r entity.Role
		if err := db.Where("slug = ?", m.RoleSlug).First(&r).Error; err != nil {
			return err
		}
		for _, pslug := range m.Perms {
			var p entity.Permission
			if err := db.Where("slug = ?", pslug).First(&p).Error; err != nil {
				return err
			}
			var cnt int64
			if err := db.Table("role_permissions").
				Where("role_id = ? AND permission_id = ?", r.ID, p.ID).
				Count(&cnt).Error; err != nil {
				return err
			}
			if cnt == 0 {
				if err := db.Table("role_permissions").Create(map[string]any{
					"role_id":       r.ID,
					"permission_id": p.ID,
					"created_at":    time.Now(),
				}).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func toTitle(slug string) string {
	parts := strings.Split(slug, ".")
	for i, p := range parts {
		if strings.ToUpper(p) == "EMR" {
			parts[i] = "EMR"
			continue
		}
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}
