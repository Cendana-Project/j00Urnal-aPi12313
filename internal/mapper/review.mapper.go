package mapper

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/repository/review"
)

// ====== Review Round ======

func ToReviewRoundResponse(r *entity.ReviewRound) response.ReviewRoundResponse {
	resp := response.ReviewRoundResponse{
		ID:               r.ID,
		ManuscriptID:     r.ManuscriptID,
		RoundNumber:      r.RoundNumber,
		Status:           string(r.Status),
		EditorDecision:   r.EditorDecision,
		DecisionComments: r.DecisionComments,
		DecisionAt:       r.DecisionAt,
		CreatedBy:        r.CreatedBy,
		CreatedAt:        r.CreatedAt,
	}

	if len(r.Assignments) > 0 {
		resp.Assignments = make([]response.ReviewAssignmentResponse, len(r.Assignments))
		for i, a := range r.Assignments {
			resp.Assignments[i] = ToReviewAssignmentResponse(&a)
		}
	}

	return resp
}

func ToReviewRoundListResponse(rounds []entity.ReviewRound) []response.ReviewRoundResponse {
	res := make([]response.ReviewRoundResponse, len(rounds))
	for i, r := range rounds {
		res[i] = ToReviewRoundResponse(&r)
	}
	return res
}

// ====== Review Assignment ======

func ToReviewAssignmentResponse(a *entity.ReviewAssignment) response.ReviewAssignmentResponse {
	resp := response.ReviewAssignmentResponse{
		ID:                   a.ID,
		ReviewRoundID:        a.ReviewRoundID,
		InvitedEmail:         a.InvitedEmail,
		AssignedBy:           a.AssignedBy,
		Status:               string(a.Status),
		InvitationExpiresAt:  a.InvitationExpiresAt,
		InvitationAcceptedAt: a.InvitationAcceptedAt,
		DueDate:              a.DueDate,
		Recommendation:       a.Recommendation,
		Comments:             a.Comments,
		CompletedAt:          a.CompletedAt,
		CreatedAt:            a.CreatedAt,
	}

	if a.ReviewerID != nil {
		resp.ReviewerID = *a.ReviewerID
	}

	if a.Reviewer != nil {
		resp.ReviewerName = formatUserName(a.Reviewer)
		resp.ReviewerEmail = a.Reviewer.Email
	}

	if a.Report != nil {
		resp.Report = ToReviewerSavedReportResponse(a.Report)
	}

	if len(a.Files) > 0 {
		resp.Files = ToReviewerAssignmentFileResponses(a.Files)
	}

	return resp
}

// ====== Candidates ======

func ToReviewerCandidateResponse(c review.ReviewerCandidate) response.ReviewerCandidateResponse {
	return response.ReviewerCandidateResponse{
		ID:          c.ID,
		Name:        formatName(c.FirstName, c.LastName),
		Email:       c.Email,
		ActiveCount: c.ActiveCount,
		DoneCount:   c.DoneCount,
	}
}

func ToReviewerCandidateListResponse(candidates []review.ReviewerCandidate) []response.ReviewerCandidateResponse {
	res := make([]response.ReviewerCandidateResponse, len(candidates))
	for i, c := range candidates {
		res[i] = ToReviewerCandidateResponse(c)
	}
	return res
}

func ToEditorCandidateResponse(c review.EditorCandidate) response.EditorCandidateResponse {
	return response.EditorCandidateResponse{
		ID:    c.ID,
		Name:  formatName(c.FirstName, c.LastName),
		Email: c.Email,
	}
}

func ToEditorCandidateListResponse(candidates []review.EditorCandidate) []response.EditorCandidateResponse {
	res := make([]response.EditorCandidateResponse, len(candidates))
	for i, c := range candidates {
		res[i] = ToEditorCandidateResponse(c)
	}
	return res
}

// ====== Submission List Item ======

func ToSubmissionListItemResponse(m *entity.Manuscript) response.SubmissionListItemResponse {
	mainAuthorID := ""
	if m.MainAuthorID != nil {
		mainAuthorID = *m.MainAuthorID
	}
	resp := response.SubmissionListItemResponse{
		ID:             m.ID,
		Title:          m.Title,
		Abstract:       m.Abstract,
		Status:         string(m.Status),
		MainAuthorID:   mainAuthorID,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}

	if m.MainAuthor != nil {
		resp.MainAuthorName = formatUserName(m.MainAuthor)
		resp.MainAuthorEmail = m.MainAuthor.Email
	}

	if m.AssignedEditorID != nil {
		resp.AssignedEditorID = m.AssignedEditorID
		if m.AssignedEditor != nil {
			name := formatUserName(m.AssignedEditor)
			resp.AssignedEditorName = &name
		}
	}

	if m.Journal != nil {
		resp.JournalName = m.Journal.Name
	}

	return resp
}

func ToSubmissionListResponse(manuscripts []entity.Manuscript) []response.SubmissionListItemResponse {
	res := make([]response.SubmissionListItemResponse, len(manuscripts))
	for i, m := range manuscripts {
		res[i] = ToSubmissionListItemResponse(&m)
	}
	return res
}

// ====== Invitation / History ======

const defaultManuscriptSection = "Article Text"

// RecommendationLabel maps stored enum to UI label (align with SubmitReviewRequest).
func RecommendationLabel(code *string) string {
	if code == nil {
		return ""
	}
	switch *code {
	case "ACCEPT":
		return "Accepted"
	case "REJECT":
		return "Declined"
	case "MAJOR_REVISION":
		return "Major revision"
	case "MINOR_REVISION":
		return "Minor revision"
	default:
		return *code
	}
}

// EditorRoundDecisionLabel maps round editor_decision to UI label.
func EditorRoundDecisionLabel(code *string) string {
	if code == nil {
		return ""
	}
	switch *code {
	case "ACCEPT":
		return "Accept"
	case "REJECT":
		return "Reject"
	case "REVISION_REQUIRED":
		return "Revision required"
	default:
		return *code
	}
}

func formatReviewerHistoryMMDD(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("01-02")
}

func ToReviewerHistoryItemResponse(row review.ReviewerHistoryRow, displayNumber int) response.ReviewerHistoryItemResponse {
	recLabel := RecommendationLabel(row.Recommendation)
	edLabel := EditorRoundDecisionLabel(row.EditorDecision)
	sec := strings.TrimSpace(row.Section)
	if sec == "" {
		sec = defaultManuscriptSection
	}
	return response.ReviewerHistoryItemResponse{
		ID:                 displayNumber,
		AssignmentID:       row.AssignmentID,
		ManuscriptID:       row.ManuscriptID,
		MMDDAssigned:       formatReviewerHistoryMMDD(row.AssignedAt),
		Sec:                sec,
		Title:              row.Title,
		Review:             recLabel,
		EditorDecision:     edLabel,
		RecommendationCode: row.Recommendation,
		EditorDecisionCode: row.EditorDecision,
		CompletedAt:        row.CompletedAt,
	}
}

// ToReviewerHistoryListResponse maps rows with 1-based id across the paged list.
func ToReviewerHistoryListResponse(rows []review.ReviewerHistoryRow, page, pageSize int) []response.ReviewerHistoryItemResponse {
	offset := 0
	if page > 0 && pageSize > 0 {
		offset = (page - 1) * pageSize
	}
	out := make([]response.ReviewerHistoryItemResponse, len(rows))
	for i := range rows {
		out[i] = ToReviewerHistoryItemResponse(rows[i], offset+i+1)
	}
	return out
}

// ====== Reviewer workspace ======

func ToReviewerAssignmentListResponse(rows []review.ReviewerAssignmentSummaryRow) []response.ReviewerAssignmentListItemResponse {
	now := time.Now().UTC()
	out := make([]response.ReviewerAssignmentListItemResponse, len(rows))
	for i, row := range rows {
		sec := strings.TrimSpace(row.ManuscriptSection)
		if sec == "" {
			sec = defaultManuscriptSection
		}
		delayed := row.AssignmentStatus == string(constant.ReviewAssignmentStatusAccepted) && now.After(row.DueDate.UTC())
		out[i] = response.ReviewerAssignmentListItemResponse{
			AssignmentID:      row.AssignmentID,
			ReviewRoundID:     row.ReviewRoundID,
			ManuscriptID:      row.ManuscriptID,
			ManuscriptTitle:   row.ManuscriptTitle,
			ManuscriptSection: sec,
			ReferenceNumber:   row.ReferenceNumber,
			RoundNumber:       row.RoundNumber,
			RoundStatus:       row.RoundStatus,
			AssignmentStatus:  row.AssignmentStatus,
			DueDate:           row.DueDate,
			IsDelayed:         delayed,
			CreatedAt:         row.CreatedAt,
		}
	}
	return out
}

func ToReviewerSavedReportResponse(rep *entity.ReviewAssignmentReport) *response.ReviewerSavedReportResponse {
	if rep == nil {
		return nil
	}
	var inner struct {
		SchemaVersion int            `json:"schema_version"`
		Answers       map[string]any `json:"answers"`
		Flags         map[string]any `json:"flags"`
	}
	_ = json.Unmarshal(rep.Payload, &inner)
	sv := rep.SchemaVersion
	if inner.SchemaVersion != 0 {
		sv = inner.SchemaVersion
	}
	ans := inner.Answers
	if ans == nil {
		ans = map[string]any{}
	}
	fl := inner.Flags
	if fl == nil {
		fl = map[string]any{}
	}
	return &response.ReviewerSavedReportResponse{
		Status:        string(rep.Status),
		SchemaVersion: sv,
		Answers:       ans,
		Flags:         fl,
	}
}

func ToReviewerAuthorBriefs(authors []entity.ManuscriptAuthor) []response.ReviewerAuthorBrief {
	out := make([]response.ReviewerAuthorBrief, 0, len(authors))
	for _, a := range authors {
		out = append(out, response.ReviewerAuthorBrief{
			Name:            strings.TrimSpace(a.AuthorName),
			Email:           strings.TrimSpace(a.AuthorEmail),
			IsCorresponding: a.IsCorresponding,
			IsPrimaryAuthor: a.IsPrimaryAuthor,
			OrderPosition:   a.OrderPosition,
		})
	}
	return out
}

func ToReviewerAssignmentFileResponses(files []entity.ReviewFile) []response.ReviewerAssignmentFileResponse {
	out := make([]response.ReviewerAssignmentFileResponse, 0, len(files))
	for _, f := range files {
		uploaderName := ""
		if f.Uploader != nil {
			uploaderName = formatUserName(f.Uploader)
		}
		out = append(out, response.ReviewerAssignmentFileResponse{
			ID:           f.ID,
			FileType:     f.FileType,
			Filename:     f.Filename,
			FileURL:      f.FilePath,
			MimeType:     f.MimeType,
			SizeBytes:    f.SizeBytes,
			UploadedAt:   f.UploadedAt,
			UploadedBy:   f.UploadedBy,
			UploaderName: uploaderName,
		})
	}
	return out
}

func ToReviewerWorkspaceResponse(
	a *entity.ReviewAssignment,
	m *entity.Manuscript,
	authors []entity.ManuscriptAuthor,
	files []entity.ReviewFile,
	rep *entity.ReviewAssignmentReport,
	isDelayed bool,
	workflow []response.ReviewerWorkflowStep,
	currentWorkflowStep int,
) response.ReviewerWorkspaceResponse {
	out := response.ReviewerWorkspaceResponse{
		Workflow:            workflow,
		CurrentWorkflowStep: currentWorkflowStep,
		Files:               ToReviewerAssignmentFileResponses(files),
		Report:              ToReviewerSavedReportResponse(rep),
	}
	if a == nil {
		return out
	}
	out.Assignment = response.ReviewerWorkspaceAssignment{
		ID:                   a.ID,
		Status:               string(a.Status),
		DueDate:              a.DueDate,
		IsDelayed:            isDelayed,
		InvitationAcceptedAt: a.InvitationAcceptedAt,
		CompletedAt:          a.CompletedAt,
		Recommendation:       a.Recommendation,
		Comments:             a.Comments,
	}
	if a.ReviewRound != nil {
		r := a.ReviewRound
		out.Round = response.ReviewerWorkspaceRound{
			ID:             r.ID,
			RoundNumber:    r.RoundNumber,
			Status:         string(r.Status),
			EditorDecision: r.EditorDecision,
			CreatedAt:      r.CreatedAt,
		}
	}
	if m != nil {
		sec := strings.TrimSpace(m.Section)
		if sec == "" {
			sec = defaultManuscriptSection
		}
		jn := ""
		if m.Journal != nil {
			jn = m.Journal.Name
		}
		editorName := ""
		if m.AssignedEditor != nil {
			editorName = formatUserName(m.AssignedEditor)
		}
		var mFile *response.ManuscriptFileResponse
		if len(m.Files) > 0 {
			f := m.Files[0] // absolute latest based on repo ordering
			mFile = &response.ManuscriptFileResponse{
				ID:         f.ID,
				FileType:   string(f.FileType),
				FilePath:   f.FilePath,
				Filename:   f.Filename,
				MimeType:   f.MimeType,
				SizeBytes:  f.SizeBytes,
				Version:    f.Version,
				UploadedAt: f.UploadedAt,
			}
		}
		out.Manuscript = response.ReviewerWorkspaceManuscript{
			ID:              m.ID,
			Title:           m.Title,
			Status:          string(m.Status),
			Section:         sec,
			Abstract:        m.Abstract,
			ReferenceNumber: m.ReferenceNumber,
			JournalName:     jn,
			ReceivedAt:      m.CreatedAt,
			HandlingEditor:  editorName,
			Authors:         ToReviewerAuthorBriefs(authors),
			File:            mFile,
		}
	} else {
		out.Manuscript = response.ReviewerWorkspaceManuscript{
			Authors: ToReviewerAuthorBriefs(authors),
		}
	}
	return out
}

// ====== Review Detail ======

func ToReviewDetailResponse(m *entity.Manuscript, rounds []entity.ReviewRound) response.ReviewDetailResponse {
	resp := response.ReviewDetailResponse{
		ManuscriptID: m.ID,
		Title:        m.Title,
		Status:       string(m.Status),
		Rounds:       ToReviewRoundListResponse(rounds),
		File:         nil,
	}

	if len(m.Files) > 0 {
		f := m.Files[0] // absolute latest
		resp.File = &response.ManuscriptFileResponse{
			ID:         f.ID,
			FileType:   string(f.FileType),
			FilePath:   f.FilePath,
			Filename:   f.Filename,
			MimeType:   f.MimeType,
			SizeBytes:  f.SizeBytes,
			Version:    f.Version,
			UploadedAt: f.UploadedAt,
		}
	}

	return resp
}

// ====== Helpers ======

func formatUserName(u *entity.User) string {
	firstName := ""
	if u.FirstName != nil {
		firstName = *u.FirstName
	}
	lastName := ""
	if u.LastName != nil {
		lastName = *u.LastName
	}
	return strings.TrimSpace(fmt.Sprintf("%s %s", firstName, lastName))
}

func formatName(firstName, lastName *string) string {
	fn := ""
	if firstName != nil {
		fn = *firstName
	}
	ln := ""
	if lastName != nil {
		ln = *lastName
	}
	return strings.TrimSpace(fmt.Sprintf("%s %s", fn, ln))
}
