package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/response"
	userrepo "github.com/api-monolith-template/internal/repository/user"
	"github.com/api-monolith-template/internal/service/auth"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct {
	svc      *auth.Service
	userRepo *userrepo.Repository
}

func NewController(svc *auth.Service, ur *userrepo.Repository) *Controller {
	return &Controller{svc: svc, userRepo: ur}
}

// GET /v1/me (global, non-tenant; role global opsional)
func (ctl *Controller) Me(c *gin.Context) {
	userUUID, err := util.GetUserIDFromContext(c)
	if err != nil || userUUID == nil {
		util.HandleError(c, constant.ErrUserNotAuthenticated)
		return
	}
	userID := userUUID.String()

	u, err := ctl.userRepo.GetByID(userID)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	if u == nil {
		util.HandleError(c, constant.ErrUserNotFound)
		return
	}

	// role global (boleh kosong), normalisasi ke UPPER
	roleSlug, err := ctl.userRepo.GetUserRoleSlug(userID)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	roleSlug = strings.ToUpper(roleSlug)
	dto := toMeDTO(u, roleSlug)

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = dto
	util.HandleResponse(c, resp, nil)
}

func toMeDTO(u *entity.User, roleSlug string) response.MeResponse {
	return response.MeResponse{
		ID:          u.ID,
		Email:       u.Email,
		Username:    u.Username,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Affiliation: u.Affiliation,
		Phone:       u.Phone,
		Status:      u.Status,
		VerifiedAt:  u.VerifiedAt,
		LastLogin:   u.LastLogin,
		Role:        roleSlug,
	}
}
