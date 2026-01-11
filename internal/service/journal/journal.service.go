package journal

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/repository/journal"
	"github.com/api-monolith-template/internal/repository/publicationfile"
	"github.com/api-monolith-template/internal/service/storage"
	"github.com/api-monolith-template/internal/service/volume"
)

type Service struct {
	journalRepo   *journal.Repository
	fileRepo      *publicationfile.Repository
	storage       *storage.Service
	volumeService *volume.Service
}

func NewService(jr *journal.Repository, fr *publicationfile.Repository, s *storage.Service, vs *volume.Service) *Service {
	return &Service{
		journalRepo:   jr,
		fileRepo:      fr,
		storage:       s,
		volumeService: vs,
	}
}

func (s *Service) Create(ctx context.Context, userID, name, description string) (*entity.Journal, error) {
	journal := &entity.Journal{
		Name:        name,
		Description: description,
		Status:      constant.PublicationStatusDraft,
		CreatedBy:   userID,
	}
	if err := s.journalRepo.Create(ctx, journal); err != nil {
		return nil, err
	}
	return journal, nil
}

func (s *Service) Update(ctx context.Context, id, name, description string) (*entity.Journal, error) {
	journal, err := s.journalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if journal == nil {
		return nil, constant.ErrRecordNotFound
	}
	// TODO: Check permission (done in handler/middleware usually, but logic here helps)

	journal.Name = name
	journal.Description = description
	if err := s.journalRepo.Update(ctx, journal); err != nil {
		return nil, err
	}
	return journal, nil
}

func (s *Service) SetStatus(ctx context.Context, id string, status constant.PublicationStatus) (*entity.Journal, error) {
	journal, err := s.journalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if journal == nil {
		return nil, constant.ErrRecordNotFound
	}

	journal.Status = status
	if err := s.journalRepo.Update(ctx, journal); err != nil {
		return nil, err
	}
	return journal, nil
}

func (s *Service) UploadCover(ctx context.Context, journalID string, fileHeader *multipart.FileHeader, uploadedBy string) (string, error) {
	journal, err := s.journalRepo.GetByID(ctx, journalID)
	if err != nil {
		return "", err
	}
	if journal == nil {
		return "", constant.ErrRecordNotFound
	}

	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	// Upload to Supabase
	// Path: journals/{id}/cover.jpg
	ext := filepath.Ext(fileHeader.Filename)
	path := fmt.Sprintf("journals/%s/cover%s", journalID, ext)
	url, err := s.storage.Upload(ctx, fileBytes, path, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	// Update Journal
	journal.CoverPath = &url
	if err := s.journalRepo.Update(ctx, journal); err != nil {
		return "", err
	}

	// Record in publication_files
	pubFile := &entity.PublicationFile{
		EntityType: constant.EntityTypeJournal,
		EntityID:   journalID,
		FileType:   constant.FileTypeCover,
		FilePath:   url,
		Filename:   fileHeader.Filename,
		MimeType:   fileHeader.Header.Get("Content-Type"),
		SizeBytes:  fileHeader.Size,
		UploadedBy: uploadedBy,
	}
	if err := s.fileRepo.Create(ctx, pubFile); err != nil {
		// Log error but don't fail the request as the main action succeeded?
		// Or fail? Let's fail for consistency.
		return "", err
	}

	return url, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	// 1. Get Journal (With Volumes)
	// JournalRepo.GetByID preloads Volumes and Volumes.Issues
	j, err := s.journalRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if j == nil {
		return constant.ErrRecordNotFound
	}

	// 2. Cascade Delete Volumes
	for _, v := range j.Volumes {
		if err := s.volumeService.Delete(ctx, v.ID); err != nil {
			return err
		}
	}

	// 3. Delete Journal Cover
	prefix := fmt.Sprintf("%s/storage/v1/object/public/%s/", config.Env.Supabase.URL, config.Env.Supabase.Bucket)
	if j.CoverPath != nil && *j.CoverPath != "" {
		if strings.HasPrefix(*j.CoverPath, prefix) {
			path := strings.TrimPrefix(*j.CoverPath, prefix)
			_ = s.storage.Delete(ctx, path)
		}
	}

	// 4. Delete Journal Record
	return s.journalRepo.Delete(ctx, id)
}

func (s *Service) GetByID(ctx context.Context, id string) (*entity.Journal, error) {
	return s.journalRepo.GetByID(ctx, id)
}

func (s *Service) GetAll(ctx context.Context, status *constant.PublicationStatus, page, limit int) ([]entity.Journal, int64, error) {
	offset := (page - 1) * limit
	return s.journalRepo.GetAll(ctx, status, offset, limit)
}
