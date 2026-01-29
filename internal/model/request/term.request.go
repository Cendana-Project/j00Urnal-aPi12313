package request

type CreateTermRequest struct {
	Content string `json:"content" binding:"required"`
}
