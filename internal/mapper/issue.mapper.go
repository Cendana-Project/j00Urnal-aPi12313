package mapper

import (
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
)

func ToIssueResponse(i *entity.Issue) response.IssueResponse {
	return response.IssueResponse{
		ID:              i.ID,
		VolumeID:        i.VolumeID,
		Number:          i.Number,
		PublicationDate: i.PublicationDate.Format("2006-01-02"),
		Status:          string(i.Status),
		CoverPath:       i.CoverPath,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
	}
}
