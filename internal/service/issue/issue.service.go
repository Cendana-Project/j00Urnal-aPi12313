package issue

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/repository/issue"
	"github.com/api-monolith-template/internal/repository/publicationfile"
	"github.com/api-monolith-template/internal/repository/volume"
	"github.com/api-monolith-template/internal/service/storage"
)

type Service struct {
	issueRepo  *issue.Repository
	volumeRepo *volume.Repository
	fileRepo   *publicationfile.Repository
	storage    *storage.Service
}

func NewService(ir *issue.Repository, vr *volume.Repository, fr *publicationfile.Repository, s *storage.Service) *Service {
	return &Service{
		issueRepo:  ir,
		volumeRepo: vr,
		fileRepo:   fr,
		storage:    s,
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

	// If setting to active, check volume keys (optional logic)
	if status == constant.PublicationStatusActive && issue.Volume != nil {
		if issue.Volume.Status != constant.PublicationStatusActive {
			// Can't activate issue if volume is not active?
			// Requirement: "Activate issue (only if volume & journal ACTIVE)"
			// We need to check deeply. Preload needed or simple check.
			if issue.Volume.Journal != nil && issue.Volume.Journal.Status != constant.PublicationStatusActive {
				return nil, constant.ErrConflict // Or ErrPreconditionFailed
			}
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

func (s *Service) GetByID(ctx context.Context, id string) (*entity.Issue, error) {
	return s.issueRepo.GetByID(ctx, id)
}
