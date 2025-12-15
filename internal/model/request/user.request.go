package request

type UpdateUserReq struct {
	FirstName   string `json:"first_name" form:"first_name" binding:"omitempty,min=1,max=100"`
	LastName    string `json:"last_name" form:"last_name" binding:"omitempty,min=1,max=100"`
	Phone       string `json:"phone" form:"phone" binding:"omitempty,e164"`
	Affiliation string `json:"affiliation" form:"affiliation" binding:"omitempty,max=300"`
}

type GetByIdentifierReq struct {
	Identifier string `json:"identifier" binding:"required"`
}

type GetAllUsersReq struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}
