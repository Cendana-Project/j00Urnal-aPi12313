package request

type CreateIssueRequest struct {
	Number          int    `json:"number" binding:"required"`
	PublicationDate string `json:"publication_date" binding:"required" time_format:"2006-01-02"` // Use string and parse manually or use binding
}

type UpdateIssueRequest struct {
	Number          int    `json:"number" binding:"required"`
	PublicationDate string `json:"publication_date" binding:"required"`
	Status          string `json:"status" binding:"required,oneof=DRAFT ACTIVE ARCHIVED"`
}

type UpdateIssueStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=DRAFT ACTIVE ARCHIVED"`
}
