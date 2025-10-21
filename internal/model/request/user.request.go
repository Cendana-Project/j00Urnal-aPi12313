package request

import (
	"time"
)

type CreateUserReq struct {
	Username  string    `json:"username" binding:"required,min=3,max=30,unique_db=users:username"`
	FirstName string    `json:"first_name" binding:"required"`
	LastName  string    `json:"last_name" binding:"required"`
	Email     string    `json:"email" binding:"required,email,unique_db=users:email"`
	Password  string    `json:"password" binding:"required,min=8,max=30"`
	Phone     string    `json:"phone" binding:"omitempty,e164"`
	Location  string    `json:"location" binding:"omitempty,max=100"`
	PhotoURL  string    `json:"photo_url" binding:"omitempty,url"`
	BirthDate time.Time `json:"birth_date" binding:"omitempty,ltefield=now"`
}

type UpdateUserReq struct {
	FirstName string    `json:"first_name" form:"first_name" binding:"omitempty,min=3,max=30"`
	LastName  string    `json:"last_name" form:"last_name" binding:"omitempty,min=3,max=30"`
	Password  string    `json:"password" form:"password" binding:"omitempty,min=8,max=30"`
	Phone     string    `json:"phone" form:"phone" binding:"omitempty,e164"`
	Location  string    `json:"location" form:"location" binding:"omitempty,max=100"`
	PhotoURL  string    `json:"photo_url" binding:"omitempty,url"`
	BirthDate time.Time `json:"birth_date" form:"birth_date" binding:"omitempty,ltefield=now"`
}

type GetByIdentifierReq struct {
	Identifier string `json:"identifier" binding:"required"`
}

type GetAllUsersReq struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}
