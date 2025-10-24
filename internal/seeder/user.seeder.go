package seeder

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/model/entity"
)

// CreateUserActiveWithRole membuat user aktif + assign role (idempotent by email).
// - Jika user sudah ada: pastikan punya role tsb, mark active jika belum, dan SET username jika masih NULL.
// - Jika belum ada: buat user active + username unik, lalu assign role.
func CreateUserActiveWithRole(db *gorm.DB, email, firstName, lastName, rawPassword, roleSlug string) (*entity.User, error) {
	if email == "" || rawPassword == "" || roleSlug == "" {
		return nil, errors.New("email/password/role wajib diisi")
	}

	// Ambil role
	var role entity.Role
	if err := db.Where("slug = ?", roleSlug).First(&role).Error; err != nil {
		return nil, err
	}

	// Cek user by email
	var u entity.User
	err := db.Where("email = ?", email).First(&u).Error
	if err == nil {
		// Assign role jika belum
		var cnt int64
		if err := db.Table("user_roles").
			Where("user_id = ? AND role_id = ?", u.ID, role.ID).
			Count(&cnt).Error; err != nil {
			return &u, err
		}
		if cnt == 0 {
			if err := db.Table("user_roles").Create(map[string]any{
				"user_id":    u.ID,
				"role_id":    role.ID,
				"created_at": time.Now(),
			}).Error; err != nil {
				return &u, err
			}
		}
		// Mark active bila belum
		if u.Status != "active" {
			now := time.Now()
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

	now := time.Now()
	u = entity.User{
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		PasswordHash: hash,
		Status:       "active",
		VerifiedAt:   &now,
	}
	if err := db.Create(&u).Error; err != nil {
		return nil, err
	}

	// Assign role
	if err := db.Table("user_roles").Create(map[string]any{
		"user_id":    u.ID,
		"role_id":    role.ID,
		"created_at": time.Now(),
	}).Error; err != nil {
		return &u, err
	}

	return &u, nil
}
