package mapper

import (
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
)

func ToVolumeResponse(v *entity.Volume) response.VolumeResponse {
	return response.VolumeResponse{
		ID:        v.ID,
		JournalID: v.JournalID,
		Year:      v.Year,
		Number:    v.Number,
		Status:    string(v.Status),
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}
