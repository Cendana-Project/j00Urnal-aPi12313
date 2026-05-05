package request

type CreateVolumeRequest struct {
	Year   int `json:"year" binding:"required"`
	Number int `json:"number" binding:"required"`
}

type UpdateVolumeRequest struct {
	Year   int `json:"year" binding:"required"`
	Number int `json:"number" binding:"required"`
}

type UpdateVolumeStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=DRAFT ACTIVE ARCHIVED"`
}