package mapper

import (
	"fmt"
	"sort"

	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
)

func ToManuscriptResponse(m *entity.Manuscript) response.ManuscriptResponse {
	// Sort co-authors by OrderPosition
	sort.Slice(m.Authors, func(i, j int) bool {
		return m.Authors[i].OrderPosition < m.Authors[j].OrderPosition
	})

	var allAuthors []response.ManuscriptAuthorResponse
	var allAuthorsSorted []string

	// 1. Add Main Author (Always First)
	if m.MainAuthor != nil {
		firstName := ""
		if m.MainAuthor.FirstName != nil {
			firstName = *m.MainAuthor.FirstName
		}
		lastName := ""
		if m.MainAuthor.LastName != nil {
			lastName = *m.MainAuthor.LastName
		}
		mainAuthorName := fmt.Sprintf("%s %s", firstName, lastName)
		allAuthors = append(allAuthors, response.ManuscriptAuthorResponse{
			ID:              m.MainAuthor.ID, // Use User ID for main author ID in this context? Or empty?
			UserID:          &m.MainAuthor.ID,
			AuthorName:      mainAuthorName,
			AuthorEmail:     m.MainAuthor.Email,
			Affiliation:     "",   // Main author affiliation typically not in User table, might need adjustment if required
			IsCorresponding: true, // Main author is typically corresponding
			OrderPosition:   0,
		})
		allAuthorsSorted = append(allAuthorsSorted, mainAuthorName)
	}

	// 2. Add Co-Authors
	for _, a := range m.Authors {
		allAuthors = append(allAuthors, response.ManuscriptAuthorResponse{
			ID:              a.ID,
			UserID:          a.UserID,
			AuthorName:      a.AuthorName,
			AuthorEmail:     a.AuthorEmail,
			Affiliation:     a.Affiliation,
			IsCorresponding: a.IsCorresponding,
			OrderPosition:   a.OrderPosition,
		})
		allAuthorsSorted = append(allAuthorsSorted, a.AuthorName)
	}

	files := make([]response.ManuscriptFileResponse, len(m.Files))
	for i, f := range m.Files {
		files[i] = response.ManuscriptFileResponse{
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

	// Assigned editor
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
		AssignedEditorID:   assignedEditorID,
		AssignedEditorName: assignedEditorName,
		PublishedAt:        m.PublishedAt,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
		Authors:            allAuthors,
		AuthorsSorted:      allAuthorsSorted,
		Files:              files,
	}
}

func ToManuscriptListResponse(ms []entity.Manuscript) []response.ManuscriptResponse {
	res := make([]response.ManuscriptResponse, len(ms))
	for i, m := range ms {
		res[i] = ToManuscriptResponse(&m)
	}
	return res
}
