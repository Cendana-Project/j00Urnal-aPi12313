package mapper

import (
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
)

func ToTermResponse(t *entity.PublicationTerm) response.TermResponse {
	return response.TermResponse{
		ID:        t.ID,
		Content:   t.Content,
		Version:   t.Version,
		IsActive:  t.IsActive,
		CreatedAt: t.CreatedAt,
	}
}
