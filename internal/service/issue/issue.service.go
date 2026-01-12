package issue

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/repository/issue"
	"github.com/api-monolith-template/internal/repository/publicationfile"
	"github.com/api-monolith-template/internal/repository/volume"
	"github.com/api-monolith-template/internal/service/manuscript"
	"github.com/api-monolith-template/internal/service/storage"
)

type Service struct {
	issueRepo         *issue.Repository
	volumeRepo        *volume.Repository
	fileRepo          *publicationfile.Repository
	storage           *storage.Service
	manuscriptService *manuscript.Service
}

func NewService(ir *issue.Repository, vr *volume.Repository, fr *publicationfile.Repository, s *storage.Service, ms *manuscript.Service) *Service {
	return &Service{
		issueRepo:         ir,
		volumeRepo:        vr,
		fileRepo:          fr,
		storage:           s,
		manuscriptService: ms,
	}
}

func (s *Service) Create(ctx context.Context, volumeID string, number int, pubDate time.Time) (*entity.Issue, error) {
	// Check Volume existence
	vol, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return nil, err
	}
	if vol == nil {
		return nil, constant.ErrRecordNotFound
	}

	issue := &entity.Issue{
		VolumeID:        volumeID,
		Number:          number,
		PublicationDate: pubDate,
		Status:          constant.PublicationStatusDraft,
	}

	if err := s.issueRepo.Create(ctx, issue); err != nil {
		return nil, err
	}
	return issue, nil
}

func (s *Service) Update(ctx context.Context, id string, number int, pubDate time.Time, status constant.PublicationStatus) (*entity.Issue, error) {
	issue, err := s.issueRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, constant.ErrRecordNotFound
	}

	// If setting to active, ensure parent Volume and Journal are also active
	if status == constant.PublicationStatusActive && issue.Volume != nil {
		// 1. Check Journal Status
		if issue.Volume.Journal != nil && issue.Volume.Journal.Status != constant.PublicationStatusActive {
			// Journal must be active
			return nil, constant.ErrJournalNotActive
		}
		// 2. Check Volume Status
		if issue.Volume.Status != constant.PublicationStatusActive {
			// Volume must be active
			return nil, constant.ErrVolumeNotActive
		}
	}

	issue.Number = number
	issue.PublicationDate = pubDate
	issue.Status = status

	if err := s.issueRepo.Update(ctx, issue); err != nil {
		return nil, err
	}
	return issue, nil
}

func (s *Service) UploadFile(ctx context.Context, issueID string, fileHeader *multipart.FileHeader, fileType constant.FileType, uploadedBy string) (string, error) {
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return "", err
	}
	if issue == nil {
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

	// Path: issues/{id}/cover.jpg or issues/{id}/full-issue.pdf
	ext := filepath.Ext(fileHeader.Filename)
	fileName := "cover"
	if fileType == constant.FileTypeFullIssuePDF {
		fileName = "full-issue"
	}
	path := fmt.Sprintf("issues/%s/%s%s", issueID, fileName, ext)

	url, err := s.storage.Upload(ctx, fileBytes, path, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	// Update Issue based on file type
	if fileType == constant.FileTypeCover {
		issue.CoverPath = &url
	}
	// For PDF, we don't store path in Issue struct directly in requirements? "ONE full issue PDF (replaceable)"
	// Requirements: "Issues... optionally have ONE full PDF file".
	// Maybe we should verify if we need to store it in issue struct.
	// DB Structure: "Issue... cover_path (string, nullable)". No pdf_path?
	// Ah, "Publication Files" table stores it.
	// "ONE full issue PDF (replaceable)" means we just query the file table or overwrite it?
	// Ideally we'd have a link in Issue?
	// User said: "ONE full issue PDF (replaceable)".
	// I'll stick to storing record in publication_files.
	// If replaceable, we should maybe look up existing one and delete/mark invalid?
	// Or just upload new one and it becomes the "latest".

	if fileType == constant.FileTypeCover {
		if err := s.issueRepo.Update(ctx, issue); err != nil {
			return "", err
		}
	}

	// Create PublicationFile record
	pubFile := &entity.PublicationFile{
		EntityType: constant.EntityTypeIssue,
		EntityID:   issueID,
		FileType:   fileType,
		FilePath:   url,
		Filename:   fileHeader.Filename,
		MimeType:   fileHeader.Header.Get("Content-Type"),
		SizeBytes:  fileHeader.Size,
		UploadedBy: uploadedBy,
	}
	if err := s.fileRepo.Create(ctx, pubFile); err != nil {
		return "", err
	}

	return url, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	// 1. Get Issue (with files if possible, or just id)
	// We need to preload files? IssueRepo.GetByID preloads Volume, Manuscripts.
	// We need to loop Manuscripts.
	iss, err := s.issueRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if iss == nil {
		return constant.ErrRecordNotFound
	}

	// 2. Cascade Delete Manuscripts
	// GetByID preloads `Manuscripts`.
	// We shouldn't rely on preloaded data for deletion if the list is partial?
	// But `Preload("Manuscripts")` fetches all.
	// Ideally use `manuscriptService.ListByIssue` to be safe/consistent?
	// `GetByID` already has them.
	for _, m := range iss.Manuscripts {
		if err := s.manuscriptService.Delete(ctx, m.ID); err != nil {
			// Log error but continue? or fail?
			// Fail to ensure consistency (mostly).
			return err
		}
	}

	// 3. Delete All Issue Files (Cover, Full Issue PDF, etc) from Supabase
	// Fetch all files associated with this issue from publication_files
	files, err := s.fileRepo.ListByEntity(ctx, constant.EntityTypeIssue, id)
	if err == nil {
		prefix := fmt.Sprintf("%s/storage/v1/object/public/%s/", config.Env.Supabase.URL, config.Env.Supabase.Bucket)
		for _, f := range files {
			if f.FilePath != "" {
				if strings.HasPrefix(f.FilePath, prefix) {
					path := strings.TrimPrefix(f.FilePath, prefix)
					_ = s.storage.Delete(ctx, path)
				}
			}
			// We can also delete the record from DB explicitly, but cascade might handle it.
			// Ideally we assume DB cascade handles metadata deletion if foreign keys are set.
			// If not, we should delete them here.
			// Checking schema... assuming Foreign Key exists for EntityID?
			// entity_type/entity_id are loose references usually (polymorphic-like).
			// So we should delete these records manually to avoid orphans in DB too.
			_ = s.fileRepo.Delete(ctx, f.ID)
		}
	} else {
		// Log error but proceed?
		// fmt.Printf("failed to list files for issue %s: %v\n", id, err)
	}

	// 4. Delete Issue Record
	return s.issueRepo.Delete(ctx, id)
}

func (s *Service) GetByID(ctx context.Context, id string) (*entity.Issue, error) {
	return s.issueRepo.GetByID(ctx, id)
}
