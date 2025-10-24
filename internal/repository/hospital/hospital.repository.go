package hospital

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Hospital struct {
	ID       string  `gorm:"column:id"`
	Code     *string `gorm:"column:code"`
	Name     string  `gorm:"column:name"`
	IsActive bool    `gorm:"column:is_active"`
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, h *Hospital) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO hospitals (id, code, name, address, city, province, country, latitude, longitude, phone, description, facilities, is_active)
		VALUES (gen_random_uuid(), ?, ?, NULL, NULL, NULL, DEFAULT, NULL, NULL, NULL, NULL, '{}'::jsonb, TRUE)
	`, h.Code, h.Name).Error
}

func (r *Repository) FindByID(ctx context.Context, id string) (*Hospital, error) {
	var row Hospital
	if err := r.db.WithContext(ctx).
		Raw(`SELECT id, code, name, is_active FROM hospitals WHERE id = ? AND deleted_at IS NULL LIMIT 1`, id).
		Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *Repository) FindByCode(ctx context.Context, code string) (*Hospital, error) {
	var row Hospital
	if err := r.db.WithContext(ctx).
		Raw(`SELECT id, code, name, is_active FROM hospitals WHERE code = ? AND deleted_at IS NULL LIMIT 1`, code).
		Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

// EnsureMembership: masukkan ke user_hospitals jika belum ada
func (r *Repository) EnsureMembership(ctx context.Context, userID, hospitalID string, setPrimary bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// insert ignore
		if err := tx.Exec(`
			INSERT INTO user_hospitals (user_id, hospital_id, is_active, is_primary)
			VALUES (?, ?, TRUE, ?)
			ON CONFLICT (user_id, hospital_id) DO NOTHING
		`, userID, hospitalID, setPrimary).Error; err != nil {
			return err
		}
		if setPrimary {
			// reset primary lain
			if err := tx.Exec(`
				UPDATE user_hospitals SET is_primary = FALSE WHERE user_id = ? AND hospital_id <> ?
			`, userID, hospitalID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// AssignHospitalRole: assign role di scope hospital
func (r *Repository) AssignHospitalRole(ctx context.Context, userID, hospitalID, roleID string) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO hospital_user_roles (hospital_id, user_id, role_id)
		VALUES (?, ?, ?)
		ON CONFLICT (hospital_id, user_id, role_id) DO NOTHING
	`, hospitalID, userID, roleID).Error
}

// Helper: resolve hospital id dari id / code
func (r *Repository) ResolveHospitalID(ctx context.Context, idOrCode string) (string, error) {
	if len(idOrCode) == 36 {
		h, err := r.FindByID(ctx, idOrCode)
		return retID(h, err)
	}
	h, err := r.FindByCode(ctx, idOrCode)
	return retID(h, err)
}

func retID(h *Hospital, err error) (string, error) {
	if err != nil {
		return "", err
	}
	if h == nil || h.ID == "" || !h.IsActive {
		return "", errors.New("hospital not found or inactive")
	}
	return h.ID, nil
}

func (r *Repository) IsUserLinkedToHospital(ctx context.Context, userID, hospitalID string) (bool, error) {
	var cnt int64
	if err := r.db.WithContext(ctx).
		Table("user_hospitals").
		Where("user_id = ? AND hospital_id = ? AND is_active = TRUE", userID, hospitalID).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}
