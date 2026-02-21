package mapper

import (
	"fmt"
	"strings"

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
		ReviewerID:           a.ReviewerID,
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

	if a.Reviewer != nil {
		resp.ReviewerName = formatUserName(a.Reviewer)
		resp.ReviewerEmail = a.Reviewer.Email
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
	resp := response.SubmissionListItemResponse{
		ID:           m.ID,
		Title:        m.Title,
		Abstract:     m.Abstract,
		Status:       string(m.Status),
		MainAuthorID: m.MainAuthorID,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
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

// ====== Invitation Info ======

// ====== Review Detail ======

func ToReviewDetailResponse(m *entity.Manuscript, rounds []entity.ReviewRound) response.ReviewDetailResponse {
	return response.ReviewDetailResponse{
		ManuscriptID: m.ID,
		Title:        m.Title,
		Status:       string(m.Status),
		Rounds:       ToReviewRoundListResponse(rounds),
	}
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
