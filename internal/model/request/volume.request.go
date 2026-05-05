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



manuscript id

47d88530-b996-422d-8faf-335ef63aa8b0


author id
550e8400-e29b-41d4-a716-446655440020


chief editor id
550e8400-e29b-41d4-a716-446655440003



a460f3c0-9e00-4a75-b9ec-0fee563a90d6

550e8400-e29b-41d4-a716-446655440040