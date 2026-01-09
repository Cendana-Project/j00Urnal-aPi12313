package mapper

import (
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
)

func ToJournalResponse(j *entity.Journal) response.JournalResponse {
	return response.JournalResponse{
		ID:          j.ID,
		Name:        j.Name,
		Description: j.Description,
		Status:      string(j.Status),
		CoverPath:   j.CoverPath,
		CreatedBy:   j.CreatedBy,
		CreatedAt:   j.CreatedAt,
		UpdatedAt:   j.UpdatedAt,
	}
}
