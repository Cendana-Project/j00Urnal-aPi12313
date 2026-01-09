package request

import "mime/multipart"

type CreateJournalRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateJournalRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateJournalStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=DRAFT ACTIVE ARCHIVED"`
}

type UploadCoverRequest struct {
	Cover *multipart.FileHeader `form:"cover" binding:"required"`
}
