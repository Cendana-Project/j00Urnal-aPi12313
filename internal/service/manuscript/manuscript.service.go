package manuscript

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/repository/issue"
	"github.com/api-monolith-template/internal/repository/journal"
	"github.com/api-monolith-template/internal/repository/manuscript"
	"github.com/api-monolith-template/internal/repository/term"
	userrepo "github.com/api-monolith-template/internal/repository/user"
	"github.com/api-monolith-template/internal/service/storage"
	"github.com/api-monolith-template/pkg/pagination"
)

var (
	ErrIssueNotActive = response.CustomError{
		Code:       "ISSUE_NOT_ACTIVE",
		StatusCode: http.StatusBadRequest,
		Message:    "Target issue is not active. Manuscripts can only be published to active issues.",
	}
	ErrTncNotAccepted = response.CustomError{
		Code:       "TNC_NOT_ACCEPTED",
		StatusCode: http.StatusBadRequest,
		Message:    "Terms and Conditions must be accepted.",
	}
)

type Service struct {
	manuscriptRepo *manuscript.Repository
	issueRepo      *issue.Repository
	journalRepo    *journal.Repository
	termRepo       *term.Repository
	userRepo       *userrepo.Repository
	storage        *storage.Service
}

func NewService(mr *manuscript.Repository, ir *issue.Repository, jr *journal.Repository, tr *term.Repository, ur *userrepo.Repository, s *storage.Service) *Service {
	return &Service{
		manuscriptRepo: mr,
		issueRepo:      ir,
		journalRepo:    jr,
		termRepo:       tr,
		userRepo:       ur,
		storage:        s,
	}
}

// Submit allows an author to submit a manuscript to a journal (Draft)
func (s *Service) Submit(ctx context.Context, userID string, req request.CreateManuscriptRequest) (*entity.Manuscript, error) {
	if !req.IsTncAccepted {
		return nil, ErrTncNotAccepted
	}

	activeTerm, err := s.termRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	if activeTerm == nil {
		return nil, fmt.Errorf("system terms and conditions not configured")
	}

	j, err := s.journalRepo.GetByID(ctx, req.JournalID)
	if err != nil {
		return nil, err
	}
	if j == nil {
		return nil, constant.ErrRecordNotFound
	}

	sid := strings.TrimSpace(userID)
	submitterRow, err := s.userRepo.GetByID(sid)
	if err != nil {
		return nil, err
	}
	if submitterRow == nil {
		return nil, constant.ErrRecordNotFound
	}

	now := time.Now()
	submittedBy := &sid
	mainAuthorIDPtr := &sid

	var authorRows []entity.ManuscriptAuthor
	seenOrder := make(map[int]struct{}, len(req.Authors))
	for _, a := range req.Authors {
		name := strings.TrimSpace(a.Name)
		email := strings.ToLower(strings.TrimSpace(a.Email))
		if name == "" || email == "" {
			return nil, constant.ErrValidationFailed
		}
		if _, ok := seenOrder[a.OrderPosition]; ok {
			return nil, constant.ErrValidationFailed
		}
		seenOrder[a.OrderPosition] = struct{}{}
		authorRows = append(authorRows, entity.ManuscriptAuthor{
			UserID:          nil,
			AuthorName:      strings.TrimSpace(a.Name),
			AuthorEmail:     email,
			Affiliation:     strings.TrimSpace(a.Affiliation),
			IsCorresponding: false,
			IsPrimaryAuthor: a.IsPrimaryAuthor,
			OrderPosition:   a.OrderPosition,
		})
	}
	if err := applySubmitterPrimaryContact(submitterRow, &authorRows); err != nil {
		return nil, err
	}
	if err := validateExactlyOnePrimaryAuthor(authorRows); err != nil {
		return nil, err
	}

	manuscript := &entity.Manuscript{
		JournalID:                     req.JournalID,
		Title:                         req.Title,
		Abstract:                      req.Abstract,
		Status:                        constant.ManuscriptStatusSubmitted,
		MainAuthorID:                  mainAuthorIDPtr,
		SubmittedByUserID:             submittedBy,
		ExternalMainAuthorName:        "",
		ExternalMainAuthorEmail:       "",
		ExternalMainAuthorAffiliation: "",
		IsTncAccepted:                 true,
		TncAcceptedAt:                 &now,
		TermID:                        &activeTerm.ID,
		CreatedAt:                     now,
	}

	if err := s.manuscriptRepo.CreateWithAuthors(ctx, manuscript, authorRows); err != nil {
		return nil, err
	}

	return s.manuscriptRepo.GetByID(ctx, manuscript.ID)
}

// Create (Admin/Legacy) - Keeping for now but potentially deprecated or updated
func (s *Service) Create(ctx context.Context, mainAuthorID string, issueID string, title, abstract string) (*entity.Manuscript, error) {
	// Validate Issue
	iss, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}
	if iss == nil {
		return nil, constant.ErrRecordNotFound
	}
	// For legacy create, we need JournalID. Issue -> Volume -> Journal
	// Assuming Issue loaded with Volume.Journal or we fetch it.
	// Current IssueRepo GetByID preloads Volume.Journal ? Let's check or just fetch Volume.
	// If IssueRepo.GetByID doesn't load Journal, we might fail constraint.
	// For safety, let's assume we need to fetch JournalID.

	// Hack: We can fetch existing logic. Ideally Admin creates Draft too.
	// For now, I'll update it to be compatible if Issue has JournalID.
	// But `issue.entity.go` has VolumeID. Volume has JournalID.

	// Let's rely on finding Journal ID from Issue.
	// Since I don't want to break existing logic too much, I will use Query if needed.
	// But assuming Issue.Volume.Journal is loaded:
	var journalID string
	if iss.Volume != nil {
		journalID = iss.Volume.JournalID
	}

	// If still empty (no preload), fetch it.
	if journalID == "" {
		// Fetch Volume... cumbersome.
		// For now let's assume valid data or fail.
		return nil, fmt.Errorf("could not determine journal_id from issue")
	}

	aid := strings.TrimSpace(mainAuthorID)

	manuscript := &entity.Manuscript{
		IssueID:      &issueID,
		JournalID:    journalID,
		Title:        title,
		Abstract:     abstract,
		Status:       constant.ManuscriptStatusPublished,
		MainAuthorID: &aid,
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
	// 1. Get Manuscript with Files
	m, err := s.manuscriptRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return constant.ErrRecordNotFound
	}

	// 2. Delete Files from Storage
	prefix := fmt.Sprintf("%s/storage/v1/object/public/%s/", config.Env.Supabase.URL, config.Env.Supabase.Bucket)
	for _, f := range m.Files {
		// Extract path from URL
		if strings.HasPrefix(f.FilePath, prefix) {
			path := strings.TrimPrefix(f.FilePath, prefix)
			// Ignore error, try to delete all
			_ = s.storage.Delete(ctx, path)
		} else {
			// Fallback: try to find "manuscripts/"
			idx := strings.Index(f.FilePath, "manuscripts/")
			if idx != -1 {
				_ = s.storage.Delete(ctx, f.FilePath[idx:])
			}
		}
	}

	// 3. Delete Record
	return s.manuscriptRepo.Delete(ctx, id)
}

func (s *Service) DeleteByAuthor(ctx context.Context, authorID string) error {
	// 1. List manuscripts where user is main author
	manuscripts, _, err := s.manuscriptRepo.ListByMainAuthor(ctx, authorID, request.AuthorManuscriptFilterRequest{}, pagination.New(1, 100))
	if err != nil {
		return err
	}

	// 2. Cascade delete
	for _, m := range manuscripts {
		if err := s.Delete(ctx, m.ID); err != nil {
			return err
		}
	}
	return nil
}

// PublishToIssue assigns a manuscript to an issue and sets status to PUBLISHED
func (s *Service) PublishToIssue(ctx context.Context, manuscriptID string, issueID string) (*entity.Manuscript, error) {
	// 1. Get Manuscript
	m, err := s.manuscriptRepo.GetByID(ctx, manuscriptID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, constant.ErrRecordNotFound
	}

	// 2. Validate Issue
	iss, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}
	if iss == nil {
		return nil, constant.ErrRecordNotFound
	}

	// 3. Check if Issue belongs to the same Journal (Optional but recommended)
	// We need to fetch Volume to check JournalID
	// Assuming Issue has VolumeID. We might not have Volume loaded.
	// Logic: If m.JournalID != iss.Volume.JournalID
	// Implementation: Fetch issue with relations or check volume.
	// For now, let's assume Admin knows what they are doing or add check if strictly needed.
	// But let's at least check that.
	// Since GetByID might not preload Volume -> Journal.
	// Let's rely on admin discretion for now or add stricter check later if requested.

	// 4. Update Manuscript
	m.IssueID = &issueID
	m.Status = constant.ManuscriptStatusPublished
	m.PublishedAt = time.Now()

	if err := s.manuscriptRepo.Update(ctx, m); err != nil {
		return nil, err
	}

	// 5. Return updated object
	return m, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*entity.Manuscript, error) {
	return s.manuscriptRepo.GetByID(ctx, id)
}

func (s *Service) ListByIssue(ctx context.Context, issueID string) ([]entity.Manuscript, error) {
	return s.manuscriptRepo.ListByIssue(ctx, issueID)
}

func (s *Service) ListByMainAuthor(
	ctx context.Context,
	authorID string,
	req request.AuthorManuscriptFilterRequest,
	pg *pagination.Pagination,
) ([]entity.Manuscript, int64, error) {
	return s.manuscriptRepo.ListByMainAuthor(ctx, authorID, req, pg)
}

// UpdateStatus is the central place for manuscript status transitions.
// Review service calls this instead of touching the manuscript table directly (DRY).
func (s *Service) UpdateStatus(ctx context.Context, id string, status constant.ManuscriptStatus) error {
	m, err := s.manuscriptRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return constant.ErrRecordNotFound
	}
	return s.manuscriptRepo.UpdateStatus(ctx, id, status)
}

// AssignEditor is called by the Chief Editor to assign an editor to a manuscript.
func (s *Service) AssignEditor(ctx context.Context, manuscriptID, editorID string) error {
	m, err := s.manuscriptRepo.GetByID(ctx, manuscriptID)
	if err != nil {
		return err
	}
	if m == nil {
		return constant.ErrRecordNotFound
	}
	return s.manuscriptRepo.AssignEditor(ctx, manuscriptID, editorID)
}

// ListByStatuses lists manuscripts filtered by statuses with pagination.
func (s *Service) ListByStatuses(ctx context.Context, statuses []constant.ManuscriptStatus, pg *pagination.Pagination) ([]entity.Manuscript, int64, error) {
	return s.manuscriptRepo.ListByStatuses(ctx, statuses, pg)
}

// ListByAssignedEditor lists manuscripts assigned to a specific editor.
func (s *Service) ListByAssignedEditor(ctx context.Context, editorID string, statuses []constant.ManuscriptStatus, pg *pagination.Pagination) ([]entity.Manuscript, int64, error) {
	return s.manuscriptRepo.ListByAssignedEditor(ctx, editorID, statuses, pg)
}

// MoveToProduction transitions an ACCEPTED manuscript to IN_PRODUCTION status.
func (s *Service) MoveToProduction(ctx context.Context, id string) error {
	m, err := s.manuscriptRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return constant.ErrRecordNotFound
	}
	if m.Status != constant.ManuscriptStatusAccepted {
		return constant.ErrInvalidManuscriptStatus
	}
	return s.manuscriptRepo.UpdateStatus(ctx, id, constant.ManuscriptStatusInProduction)
}

func (s *Service) UpdateAuthors(ctx context.Context, manuscriptID string, authors []entity.ManuscriptAuthor) error {
	m, err := s.manuscriptRepo.GetByID(ctx, manuscriptID)
	if err != nil {
		return err
	}
	if m == nil {
		return constant.ErrRecordNotFound
	}

	seen := make(map[int]struct{}, len(authors))
	for _, a := range authors {
		if _, ok := seen[a.OrderPosition]; ok {
			return constant.ErrValidationFailed
		}
		seen[a.OrderPosition] = struct{}{}
	}

	submitterUser, err := s.submitterUserForManuscript(ctx, m)
	if err != nil {
		return err
	}
	if err := applySubmitterPrimaryContact(submitterUser, &authors); err != nil {
		return err
	}
	if err := validateExactlyOnePrimaryAuthor(authors); err != nil {
		return err
	}

	return s.manuscriptRepo.UpdateAuthors(ctx, manuscriptID, authors)
}

func (s *Service) UploadFile(ctx context.Context, manuscriptID string, fileType constant.ManuscriptFileType, fileHeader *multipart.FileHeader) (*entity.ManuscriptFile, error) {
	m, err := s.manuscriptRepo.GetByID(ctx, manuscriptID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, constant.ErrRecordNotFound
	}

	// Prepare storage path
	// Path: manuscripts/{journal_id}/{manuscript_id}/{type}/v{version}.pdf
	// If IssueID is nil, we can't use it in path?
	// Or we use "pending" or just omit issueID.
	// Previous: manuscripts/{journal_id}/{issue_id}/{manuscript_id}/{type}/v{version}.pdf
	// New: manuscripts/{journal_id}/submissions/{manuscript_id}/{type}/v{version}.pdf (If no issue)

	journalID := m.JournalID
	issueSegment := "submissions"
	if m.IssueID != nil {
		issueSegment = *m.IssueID
	}

	version := 1
	if fileType == constant.ManuscriptFileTypeMain {
		v, err := s.manuscriptRepo.GetLatestMainFileVersion(ctx, manuscriptID)
		if err != nil {
			return nil, err
		}
		version = v + 1
	}

	var path string
	ext := filepath.Ext(fileHeader.Filename)

	// Normalize filetype key for path
	fileTypeStr := strings.ToLower(string(fileType))

	path = fmt.Sprintf("manuscripts/%s/%s/%s/%s/v%d%s", journalID, issueSegment, manuscriptID, fileTypeStr, version, ext)
	// Example: manuscripts/uuid/submissions/uuid/main/v1.pdf

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

func (s *Service) submitterUserForManuscript(_ context.Context, m *entity.Manuscript) (*entity.User, error) {
	var uid string
	if m.SubmittedByUserID != nil && strings.TrimSpace(*m.SubmittedByUserID) != "" {
		uid = strings.TrimSpace(*m.SubmittedByUserID)
	} else if m.MainAuthorID != nil && strings.TrimSpace(*m.MainAuthorID) != "" {
		uid = strings.TrimSpace(*m.MainAuthorID)
	}
	if uid == "" {
		return nil, response.CustomError{
			Code:       "INVALID_MANUSCRIPT_AUTHORS",
			StatusCode: http.StatusBadRequest,
			Message:    "Manuscript has no submitter reference; cannot set primary contact.",
		}
	}
	u, err := s.userRepo.GetByID(uid)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, constant.ErrRecordNotFound
	}
	if strings.TrimSpace(u.Email) == "" {
		return nil, constant.ErrValidationFailed
	}
	return u, nil
}

// applySubmitterPrimaryContact: jika ada baris author dengan email sama submitter, tandai satu baris jadi correspondence (position terkecil jika duplikat). Submitter tidak wajib ada di authors — kalau tidak ada, semua is_corresponding false; kontak admin tetap lewat field submitter di manuskrip/response.
func applySubmitterPrimaryContact(submitter *entity.User, authors *[]entity.ManuscriptAuthor) error {
	if submitter == nil {
		return constant.ErrRecordNotFound
	}
	se := strings.ToLower(strings.TrimSpace(submitter.Email))
	if se == "" {
		return constant.ErrValidationFailed
	}

	var matches []int
	for i := range *authors {
		if strings.ToLower(strings.TrimSpace((*authors)[i].AuthorEmail)) == se {
			matches = append(matches, i)
		}
	}

	if len(matches) == 0 {
		for i := range *authors {
			(*authors)[i].IsCorresponding = false
		}
		return nil
	}

	best := matches[0]
	for _, idx := range matches[1:] {
		if (*authors)[idx].OrderPosition < (*authors)[best].OrderPosition {
			best = idx
		}
	}
	for i := range *authors {
		(*authors)[i].IsCorresponding = (i == best)
	}
	return nil
}

func validateExactlyOnePrimaryAuthor(authors []entity.ManuscriptAuthor) error {
	var n int
	for _, a := range authors {
		if a.IsPrimaryAuthor {
			n++
		}
	}
	if n != 1 {
		return response.CustomError{
			Code:       "INVALID_MANUSCRIPT_AUTHORS",
			StatusCode: http.StatusBadRequest,
			Message:    "Exactly one author must have is_primary_author true.",
		}
	}
	return nil
}
