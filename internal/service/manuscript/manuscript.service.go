package manuscript

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/repository/issue"
	"github.com/api-monolith-template/internal/repository/manuscript"
	"github.com/api-monolith-template/internal/service/storage"
)

var (
	ErrIssueNotActive = response.CustomError{
		Code:       "ISSUE_NOT_ACTIVE",
		StatusCode: http.StatusBadRequest,
		Message:    "Target issue is not active. Manuscripts can only be published to active issues.",
	}
)

type Service struct {
	manuscriptRepo *manuscript.Repository
	issueRepo      *issue.Repository
	storage        *storage.Service
}

func NewService(mr *manuscript.Repository, ir *issue.Repository, s *storage.Service) *Service {
	return &Service{
		manuscriptRepo: mr,
		issueRepo:      ir,
		storage:        s,
	}
}

func (s *Service) Create(ctx context.Context, mainAuthorID string, issueID string, title, abstract string) (*entity.Manuscript, error) {
	// Validate Issue
	iss, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}
	if iss == nil {
		return nil, constant.ErrRecordNotFound
	}
	if iss.Status != constant.PublicationStatusActive {
		return nil, ErrIssueNotActive
	}

	manuscript := &entity.Manuscript{
		IssueID:      issueID,
		Title:        title,
		Abstract:     abstract,
		Status:       constant.PublicationStatusPublished,
		MainAuthorID: mainAuthorID,
		PublishedAt:  time.Now(),
	}

	if err := s.manuscriptRepo.Create(ctx, manuscript); err != nil {
		return nil, err
	}

	return manuscript, nil
}

func (s *Service) Update(ctx context.Context, id string, title, abstract string) (*entity.Manuscript, error) {
	manuscript, err := s.manuscriptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if manuscript == nil {
		return nil, constant.ErrRecordNotFound
	}

	manuscript.Title = title
	manuscript.Abstract = abstract

	if err := s.manuscriptRepo.Update(ctx, manuscript); err != nil {
		return nil, err
	}

	return manuscript, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.manuscriptRepo.Delete(ctx, id)
}

func (s *Service) GetByID(ctx context.Context, id string) (*entity.Manuscript, error) {
	return s.manuscriptRepo.GetByID(ctx, id)
}

func (s *Service) ListByIssue(ctx context.Context, issueID string) ([]entity.Manuscript, error) {
	return s.manuscriptRepo.ListByIssue(ctx, issueID)
}

func (s *Service) UpdateAuthors(ctx context.Context, manuscriptID string, authors []entity.ManuscriptAuthor) error {
	// Check if manuscript exists
	m, err := s.manuscriptRepo.GetByID(ctx, manuscriptID)
	if err != nil {
		return err
	}
	if m == nil {
		return constant.ErrRecordNotFound
	}

	return s.manuscriptRepo.UpdateAuthors(ctx, manuscriptID, authors)
}

func (s *Service) UploadFile(ctx context.Context, manuscriptID string, fileType constant.FileType, fileHeader *multipart.FileHeader) (*entity.ManuscriptFile, error) {
	m, err := s.manuscriptRepo.GetByID(ctx, manuscriptID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, constant.ErrRecordNotFound
	}

	// Prepare storage path
	// Path: manuscripts/{journal_id}/{issue_id}/{manuscript_id}/{type}/v{version}.pdf
	journalID := m.Issue.Volume.Journal.ID
	issueID := m.IssueID

	version := 1
	if fileType == constant.FileTypeMain {
		v, err := s.manuscriptRepo.GetLatestMainFileVersion(ctx, manuscriptID)
		if err != nil {
			return nil, err
		}
		version = v + 1
	}

	var path string
	ext := filepath.Ext(fileHeader.Filename)
	switch fileType {
	case constant.FileTypeMain:
		path = fmt.Sprintf("manuscripts/%s/%s/%s/main/v%d%s", journalID, issueID, manuscriptID, version, ext)
	case constant.FileTypeFigure:
		path = fmt.Sprintf("manuscripts/%s/%s/%s/figures/%s", journalID, issueID, manuscriptID, fileHeader.Filename)
	case constant.FileTypeTable:
		path = fmt.Sprintf("manuscripts/%s/%s/%s/tables/%s", journalID, issueID, manuscriptID, fileHeader.Filename)
	case constant.FileTypeSupplement:
		path = fmt.Sprintf("manuscripts/%s/%s/%s/supplements/%s", journalID, issueID, manuscriptID, fileHeader.Filename)
	default:
		return nil, fmt.Errorf("invalid file type")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	// Upload to Supabase
	url, err := s.storage.Upload(ctx, fileBytes, path, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	// Record in manuscript_files
	file := &entity.ManuscriptFile{
		ManuscriptID: manuscriptID,
		FileType:     fileType,
		FilePath:     url,
		Filename:     fileHeader.Filename,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		SizeBytes:    fileHeader.Size,
		Version:      version,
		UploadedAt:   time.Now(),
	}

	if err := s.manuscriptRepo.AddFile(ctx, file); err != nil {
		return nil, err
	}

	return file, nil
}
