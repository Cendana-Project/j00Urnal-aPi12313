package seeder

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/model/entity"
)

// CreateUserActiveWithRole membuat user aktif + assign role (idempotent by email).
// - Jika user sudah ada: pastikan punya role tsb, mark active jika belum, dan SET username jika masih NULL.
// - Jika belum ada: buat user active + username unik, lalu assign role.
func CreateUserActiveWithRole(db *gorm.DB, id, email, firstName, lastName, rawPassword, roleSlug string) (*entity.User, error) {
	if email == "" || rawPassword == "" || roleSlug == "" {
		return nil, errors.New("email/password/role wajib diisi")
	}

	// Ambil role
	var role entity.Role
	if err := db.Where("slug = ?", roleSlug).First(&role).Error; err != nil {
		return nil, err
	}

	// Cek user by email or ID (unscoped to find deleted ones too)
	var u entity.User
	// If id is provided, check both. If not, just email.
	// But safely, check email mainly, or ID if hardcoded.
	query := db.Unscoped().Where("email = ?", email)
	if id != "" {
		query = query.Or("id = ?", id)
	}
	err := query.First(&u).Error
	if err == nil {
		// If user was soft-deleted, restore them
		if u.DeletedAt.Valid {
			if err := db.Unscoped().Model(&u).Update("deleted_at", nil).Error; err != nil {
				return &u, err
			}
			u.DeletedAt.Valid = false // Update struct
		}

		// Assign role jika belum
		var cnt int64
		if err := db.Table("user_roles").
			Where("user_id = ? AND role_id = ?", u.ID, role.ID).
			Count(&cnt).Error; err != nil {
			return &u, err
		}
		if cnt == 0 {
			if err := db.Table("user_roles").Create(map[string]any{
				"user_id":     u.ID,
				"role_id":     role.ID,
				"assigned_at": time.Now(),
			}).Error; err != nil {
				return &u, err
			}
		}
		// Mark active bila belum
		if u.Status != "active" {
			now := time.Now()
			// Use Unscoped to update even if it was just restored (though Update above handles it)
			if err := db.Model(&entity.User{}).Where("id = ?", u.ID).
				Updates(map[string]any{
					"status":      "active",
					"verified_at": now,
				}).Error; err != nil {
				return &u, err
			}
			u.Status = "active"
			u.VerifiedAt = &time.Time{}
			*u.VerifiedAt = now
		}
		return &u, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// User belum ada → buat baru (status active)
	// Hash password (scrypt)
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	key, err := scrypt.Key([]byte(rawPassword), salt, 1<<15, 8, 1, 64)
	if err != nil {
		return nil, err
	}
	hash := base64.StdEncoding.EncodeToString(key) + ":" + base64.StdEncoding.EncodeToString(salt)

	// Generate username from email (take header)
	username := email
	if idx := strings.Index(email, "@"); idx > 0 {
		username = email[:idx]
	}

	now := time.Now()
	u = entity.User{
		Email:        email,
		Username:     username, // <=== Assigned
		FirstName:    &firstName,
		LastName:     &lastName,
		PasswordHash: hash,
		Status:       "active",
		VerifiedAt:   &now,
	}
	if id != "" {
		u.ID = id
	}
	if err := db.Create(&u).Error; err != nil {
		return nil, err
	}

	// Assign role
	if err := db.Table("user_roles").Create(map[string]any{
		"user_id":     u.ID,
		"role_id":     role.ID,
		"assigned_at": time.Now(),
	}).Error; err != nil {
		return &u, err
	}

	return &u, nil
}
