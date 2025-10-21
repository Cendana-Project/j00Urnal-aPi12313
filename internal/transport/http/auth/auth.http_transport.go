package auth

import (
	"github.com/api-monolith-template/internal/constant"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/service/auth"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct{ svc *auth.Service }

func NewController(svc *auth.Service) *Controller { return &Controller{svc: svc} }

// POST /v1/auth/register
func (ctl *Controller) Register(c *gin.Context) {
	var req request.RegisterLiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	u, err := ctl.svc.RegisterLite(c.Request.Context(), &req)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{"user_id": u.ID, "email": u.Email, "status": "pending"}
	util.HandleResponse(c, resp, nil)
}

// POST /v1/auth/resend-pin
func (ctl *Controller) ResendPIN(c *gin.Context) {
	var req request.ResendPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	if err := ctl.svc.ResendPIN(c.Request.Context(), req.Email); err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{"email": req.Email, "status": "pending"}
	util.HandleResponse(c, resp, nil)
}

// POST /v1/auth/verify-pin
func (ctl *Controller) VerifyPIN(c *gin.Context) {
	var req request.VerifyPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	pair, err := ctl.svc.VerifyPIN(c.Request.Context(), req.Email, req.PIN)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken}
	util.HandleResponse(c, resp, nil)
}

// POST /v1/auth/login
func (ctl *Controller) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	pair, err := ctl.svc.Login(c.Request.Context(), req.Identity, req.Password)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken}
	util.HandleResponse(c, resp, nil)
}

// POST /v1/auth/refresh
func (ctl *Controller) Refresh(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	pair, err := ctl.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken}
	util.HandleResponse(c, resp, nil)
}

// POST /v1/auth/choose-role (protected)
func (ctl *Controller) ChooseRole(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		res := constant.ErrUnauthorized.ToResponse()
		util.HandleResponse(c, &res, nil)
		return
	}

	var req request.ChooseRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}

	if err := ctl.svc.ChooseRole(c.Request.Context(), userID, req.Role); err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{"role": req.Role}
	util.HandleResponse(c, resp, nil)
}
