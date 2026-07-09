package mapper

import (
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
)

func ToIssueResponse(i *entity.Issue) response.IssueResponse {
	resp := response.IssueResponse{
		ID:              i.ID,
		VolumeID:        i.VolumeID,
		Number:          i.Number,
		PublicationDate: i.PublicationDate.Format("2006-01-02"),
		Status:          string(i.Status),
		CoverPath:       i.CoverPath,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
	}

	if i.Volume != nil {
		v := i.Volume
		volResp := response.VolumeResponse{
			ID:        v.ID,
			JournalID: v.JournalID,
			Year:      v.Year,
			Number:    v.Number,
			Status:    string(v.Status),
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
		if v.Journal != nil {
			jResp := ToJournalResponse(v.Journal)
			volResp.Journal = &jResp
		}
		resp.Volume = &volResp
	}

	if len(i.Manuscripts) > 0 {
		resp.Manuscripts = make([]response.ManuscriptResponse, len(i.Manuscripts))
		for k, m := range i.Manuscripts {
			resp.Manuscripts[k] = ToManuscriptResponse(&m)
		}
	}

	return resp
}

func ToIssueResponses(issues []*entity.Issue) []*response.IssueResponse {
	var responses []*response.IssueResponse
	for _, i := range issues {
		resp := ToIssueResponse(i)
		responses = append(responses, &resp)
	}
	return responses
}
