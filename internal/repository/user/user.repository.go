package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/model/entity"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) FindByEmail(email string) (*entity.User, error) {
	var u entity.User
	err := r.db.Where("LOWER(email)=LOWER(?)", strings.ToLower(email)).
		First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *Repository) FindByUsername(uname string) (*entity.User, error) {
	var u entity.User
	err := r.db.Where("LOWER(username)=LOWER(?)", strings.ToLower(uname)).
		First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *Repository) ExistsUsername(uname string) (bool, error) {
	var cnt int64
	if err := r.db.Model(&entity.User{}).
		Where("LOWER(username)=LOWER(?)", strings.ToLower(uname)).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (r *Repository) Create(u *entity.User) error { return r.db.Create(u).Error }

func (r *Repository) MarkVerified(id string) error {
	now := time.Now()
	return r.db.Model(&entity.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":      "active",
			"verified_at": now,
			"updated_at":  now,
		}).Error
}

func (r *Repository) UpdateByID(id string, fields map[string]any) error {
	fields["updated_at"] = time.Now()
	return r.db.Model(&entity.User{}).Where("id = ?", id).Updates(fields).Error
}

// Profiles (UPSERT)
func (r *Repository) UpsertPatientProfile(p map[string]any) error {
	return r.db.Exec(`
		INSERT INTO patient_profiles (user_id, height_cm, weight_kg, allergies, medical_hist, created_at, updated_at)
		VALUES (@user_id, @height_cm, @weight_kg, @allergies, @medical_hist, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
		  height_cm   = EXCLUDED.height_cm,
		  weight_kg   = EXCLUDED.weight_kg,
		  allergies   = EXCLUDED.allergies,
		  medical_hist= EXCLUDED.medical_hist,
		  updated_at  = NOW();
	`, p).Error
}

func (r *Repository) UpsertDoctorProfile(p map[string]any) error {
	return r.db.Exec(`
		INSERT INTO doctor_profiles (user_id, sip_number, specialty, created_at, updated_at)
		VALUES (@user_id, @sip_number, @specialty, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
		  sip_number = EXCLUDED.sip_number,
		  specialty  = EXCLUDED.specialty,
		  updated_at = NOW();
	`, p).Error
}

func (r *Repository) GetByID(id string) (*entity.User, error) {
	var u entity.User
	if err := r.db.First(&u, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetByEmail(email string) (*entity.User, error) { // <=== added
	var u entity.User
	err := r.db.First(&u, "email = ? AND deleted_at IS NULL", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *Repository) Update(u *entity.User) error { // <=== added
	return r.db.Save(u).Error
}

type InsertUser struct {
	ID           string
	Email        string
	Username     string
	Phone        string
	FirstName    string
	LastName     string
	PasswordHash string
	VerifiedAt   time.Time
}

func (r *Repository) InsertActive(ctx context.Context, in InsertUser) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO users (id, email, username, phone, first_name, last_name, password_hash, status, verified_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'active', ?, NOW())
	`, in.ID, in.Email, in.Username, in.Phone, in.FirstName, in.LastName, in.PasswordHash, in.VerifiedAt).Error
}
func (r *Repository) ExistsNIK(nik string) (bool, error) {
	if strings.TrimSpace(nik) == "" {
		return false, nil
	}
	var cnt int64
	if err := r.db.Model(&entity.User{}).
		Where("nik = ? AND deleted_at IS NULL", nik).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

type InsertUserFull struct {
	ID            string
	Email         string
	Username      *string
	Phone         *string
	FirstName     *string
	LastName      *string
	DOB           *time.Time
	Address       *string
	Gender        *string // "L" | "P"
	NIK           *string
	PasswordPlain string // input (akan di-hash di service)
	PasswordHash  string // hasil scrypt "hash:salt"
	Status        string // "active"
	VerifiedAt    *time.Time
}

// InsertActiveFull: membuat user aktif dengan field lengkap (tanpa verifikasi)
func (r *Repository) InsertActiveFull(ctx context.Context, in InsertUserFull) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO users (
			id, email, username, first_name, last_name, phone, dob, address, gender, nik,
			password_hash, status, verified_at, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active', ?, NOW())
	`, in.ID, in.Email, in.Username, in.FirstName, in.LastName, in.Phone, in.DOB, in.Address, in.Gender, in.NIK,
		in.PasswordHash, in.VerifiedAt).Error
}

func (r *Repository) GetPatientProfileByUserID(userID string) (heightCM, weightKG *int, allergies, medicalHist *string, err error) { // <=== added
	type row struct {
		HeightCM    *int
		WeightKG    *int
		Allergies   *string
		MedicalHist *string
	}
	var out row
	err = r.db.Raw(`
		SELECT height_cm, weight_kg, allergies, medical_hist
		FROM patient_profiles
		WHERE user_id = ?
		LIMIT 1
	`, userID).Scan(&out).Error
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return out.HeightCM, out.WeightKG, out.Allergies, out.MedicalHist, nil
}

// GetDoctorProfileByUserID mengambil data doctor_profiles untuk user tertentu.
func (r *Repository) GetDoctorProfileByUserID(userID string) (sipNumber, specialty *string, err error) { // <=== added
	type row struct {
		SIPNumber *string
		Specialty *string
	}
	var out row
	err = r.db.Raw(`
		SELECT sip_number, specialty
		FROM doctor_profiles
		WHERE user_id = ?
		LIMIT 1
	`, userID).Scan(&out).Error
	if err != nil {
		return nil, nil, err
	}
	return out.SIPNumber, out.Specialty, nil
}

// ExistsPatientProfile returns true if a patient profile already exists for user_id.
func (r *Repository) ExistsPatientProfile(userID string) (bool, error) { // <=== added
	var exists bool
	err := r.db.Raw(`SELECT EXISTS(SELECT 1 FROM patient_profiles WHERE user_id = ?)`, userID).
		Scan(&exists).Error
	return exists, err
}

// ExistsDoctorProfile returns true if a doctor profile already exists for user_id.
func (r *Repository) ExistsDoctorProfile(userID string) (bool, error) { // <=== added
	var exists bool
	err := r.db.Raw(`SELECT EXISTS(SELECT 1 FROM doctor_profiles WHERE user_id = ?)`, userID).
		Scan(&exists).Error
	return exists, err
}
