package mapper

import (
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
)

func ToManuscriptResponse(m *entity.Manuscript) response.ManuscriptResponse {
	authors := make([]response.ManuscriptAuthorResponse, len(m.Authors))
	for i, a := range m.Authors {
		authors[i] = response.ManuscriptAuthorResponse{
			ID:              a.ID,
			UserID:          a.UserID,
			AuthorName:      a.AuthorName,
			AuthorEmail:     a.AuthorEmail,
			Affiliation:     a.Affiliation,
			IsCorresponding: a.IsCorresponding,
			OrderPosition:   a.OrderPosition,
		}
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

	return response.ManuscriptResponse{
		ID:           m.ID,
		IssueID:      m.IssueID,
		Title:        m.Title,
		Abstract:     m.Abstract,
		Status:       string(m.Status),
		MainAuthorID: m.MainAuthorID,
		PublishedAt:  m.PublishedAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		Authors:      authors,
		Files:        files,
	}
}

func ToManuscriptListResponse(ms []entity.Manuscript) []response.ManuscriptResponse {
	res := make([]response.ManuscriptResponse, len(ms))
	for i, m := range ms {
		res[i] = ToManuscriptResponse(&m)
	}
	return res
}
