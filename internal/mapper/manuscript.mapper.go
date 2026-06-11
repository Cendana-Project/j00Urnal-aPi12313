package mapper

import (
	"fmt"
	"sort"
	"strings"

	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
)

func ToManuscriptResponse(m *entity.Manuscript) response.ManuscriptResponse {
	sort.Slice(m.Authors, func(i, j int) bool {
		return m.Authors[i].OrderPosition < m.Authors[j].OrderPosition
	})

	allAuthors := make([]response.ManuscriptAuthorResponse, 0, len(m.Authors)+1)
	allAuthorsSorted := make([]string, 0, len(m.Authors)+1)

	if len(m.Authors) > 0 {
		for _, a := range m.Authors {
			allAuthors = append(allAuthors, response.ManuscriptAuthorResponse{
				ID:               a.ID,
				UserID:           a.UserID,
				AuthorName:       a.AuthorName,
				AuthorEmail:      a.AuthorEmail,
				Affiliation:      a.Affiliation,
				OrderPosition:    a.OrderPosition,
				IsPrimaryContact: a.IsCorresponding,
				IsPrimaryAuthor:  a.IsPrimaryAuthor,
			})
			allAuthorsSorted = append(allAuthorsSorted, a.AuthorName)
		}
	} else {
		// Legacy manuscripts without manuscript_authors rows
		external := strings.TrimSpace(m.ExternalMainAuthorEmail) != "" || strings.TrimSpace(m.ExternalMainAuthorName) != ""
		if external {
			aff := strings.TrimSpace(m.ExternalMainAuthorAffiliation)
			allAuthors = append(allAuthors, response.ManuscriptAuthorResponse{
				ID:               "",
				UserID:           nil,
				AuthorName:       strings.TrimSpace(m.ExternalMainAuthorName),
				AuthorEmail:      strings.TrimSpace(strings.ToLower(m.ExternalMainAuthorEmail)),
				Affiliation:      aff,
				OrderPosition:    0,
				IsPrimaryContact: true,
				IsPrimaryAuthor:  true,
			})
			allAuthorsSorted = append(allAuthorsSorted, strings.TrimSpace(m.ExternalMainAuthorName))
		} else if m.MainAuthor != nil {
			firstName := ""
			if m.MainAuthor.FirstName != nil {
				firstName = *m.MainAuthor.FirstName
			}
			lastName := ""
			if m.MainAuthor.LastName != nil {
				lastName = *m.MainAuthor.LastName
			}
			mainAuthorName := strings.TrimSpace(fmt.Sprintf("%s %s", firstName, lastName))
			if mainAuthorName == "" {
				mainAuthorName = m.MainAuthor.Email
			}
			uid := m.MainAuthor.ID
			allAuthors = append(allAuthors, response.ManuscriptAuthorResponse{
				ID:               m.MainAuthor.ID,
				UserID:           &uid,
				AuthorName:       mainAuthorName,
				AuthorEmail:      m.MainAuthor.Email,
				Affiliation:      "",
				OrderPosition:    0,
				IsPrimaryContact: true,
				IsPrimaryAuthor:  true,
			})
			allAuthorsSorted = append(allAuthorsSorted, mainAuthorName)
		}
	}

	var mFile *response.ManuscriptFileResponse
	if len(m.Files) > 0 {
		f := m.Files[0] // absolute latest
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

	var assignedEditorID *string
	var assignedEditorName *string
	if m.AssignedEditorID != nil {
		assignedEditorID = m.AssignedEditorID
		if m.AssignedEditor != nil {
			firstName := ""
			if m.AssignedEditor.FirstName != nil {
				firstName = *m.AssignedEditor.FirstName
			}
			lastName := ""
			if m.AssignedEditor.LastName != nil {
				lastName = *m.AssignedEditor.LastName
			}
			name := fmt.Sprintf("%s %s", firstName, lastName)
			assignedEditorName = &name
		}
	}

	return response.ManuscriptResponse{
		ID:                 m.ID,
		IssueID:            m.IssueID,
		Title:              m.Title,
		Abstract:           m.Abstract,
		Status:             string(m.Status),
		MainAuthorID:       m.MainAuthorID,
		SubmittedByUserID:  m.SubmittedByUserID,
		Submitter:          submitterBrief(m.SubmittedBy),
		AssignedEditorID:   assignedEditorID,
		AssignedEditorName: assignedEditorName,
		PublishedAt:        m.PublishedAt,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
		Authors:            allAuthors,
		AuthorsSorted:      allAuthorsSorted,
		File:               mFile,
	}
}

func ToManuscriptListResponse(ms []entity.Manuscript) []response.ManuscriptResponse {
	res := make([]response.ManuscriptResponse, len(ms))
	for i, m := range ms {
		res[i] = ToManuscriptResponse(&m)
	}
	return res
}

func submitterBrief(u *entity.User) *response.ManuscriptSubmitterBrief {
	if u == nil {
		return nil
	}
	fn, ln := "", ""
	if u.FirstName != nil {
		fn = *u.FirstName
	}
	if u.LastName != nil {
		ln = *u.LastName
	}
	name := strings.TrimSpace(fmt.Sprintf("%s %s", fn, ln))
	if name == "" {
		name = u.Email
	}
	var aff *string
	if u.Affiliation != nil && strings.TrimSpace(*u.Affiliation) != "" {
		s := strings.TrimSpace(*u.Affiliation)
		aff = &s
	}
	return &response.ManuscriptSubmitterBrief{
		UserID:      u.ID,
		Email:       u.Email,
		Name:        name,
		Affiliation: aff,
	}
}
