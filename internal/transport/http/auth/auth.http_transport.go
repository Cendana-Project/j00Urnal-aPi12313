package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/service/auth"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct{ svc *auth.Service }

func NewController(svc *auth.Service) *Controller { return &Controller{svc: svc} }

// =======================================
// AUTH — PUBLIC REGISTRATION & OTP FLOW
// =======================================

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

func (ctl *Controller) ResendPIN(c *gin.Context) {
	var req request.ResendPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	email := strings.TrimSpace(req.Email)
	if err := ctl.svc.ResendPIN(c.Request.Context(), email); err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{"email": email, "status": "pending"}
	util.HandleResponse(c, resp, nil)
}

func (ctl *Controller) VerifyPIN(c *gin.Context) {
	var req request.VerifyPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	email := strings.TrimSpace(req.Email)
	pin := strings.TrimSpace(req.PIN)
	tokens, aexp, rexp, err := ctl.svc.VerifyPIN(c.Request.Context(), email, pin) // <=== changed
	if err != nil {
		util.HandleError(c, err)
		return
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{
		"access_token":          tokens.AccessToken,
		"refresh_token":         tokens.RefreshToken,
		"accessTokenExpiredAt":  aexp.UTC().Format(time.RFC3339), // <=== added
		"refreshTokenExpiredAt": rexp.UTC().Format(time.RFC3339), // <=== added
	}
	util.HandleResponse(c, resp, nil)
}

// =============================
// AUTH — LOGIN (OPSI A: split)
// =============================

// Login publik (tanpa hospital hint)
func (ctl *Controller) LoginPublic(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	identity := strings.TrimSpace(req.Identity)
	password := strings.TrimSpace(req.Password)

	tokens, roles, aexp, rexp, err := ctl.svc.Login(c.Request.Context(), identity, password) // <=== changed
	if err != nil {
		util.HandleError(c, err)
		return
	}

	// Pilih slug pertama jika ada banyak role
	roleSlug := ""
	if len(roles) > 0 {
		roleSlug = roles[0].Slug // <=== changed: kirim hanya slug
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = response.LoginResponse{ // <=== changed
		AccessToken:           tokens.AccessToken,
		RefreshToken:          tokens.RefreshToken,
		Role:                  roleSlug,
		AccessTokenExpiredAt:  aexp.UTC().Format(time.RFC3339),
		RefreshTokenExpiredAt: rexp.UTC().Format(time.RFC3339),
	}
	util.HandleResponse(c, resp, nil)
}

func (ctl *Controller) LoginHospital(c *gin.Context) {
	var req request.LoginHospitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	hint := ""
	if req.HospitalID != nil && *req.HospitalID != "" {
		hint = *req.HospitalID
	} else if req.HospitalCode != nil && *req.HospitalCode != "" {
		hint = *req.HospitalCode
	}

	res, err := ctl.svc.LoginHospital(c.Request.Context(), req.Identifier, req.Password, hint)
	if err != nil {
		util.HandleError(c, err)
		return
	}

	// Pilih slug pertama
	roleSlug := ""
	if len(res.Roles) > 0 {
		roleSlug = res.Roles[0].Slug // <=== changed
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = response.LoginHospitalResponse{ // <=== changed
		AccessToken:           res.AccessToken,
		RefreshToken:          res.RefreshToken,
		ExpiresIn:             res.ExpiresIn,
		TokenType:             res.TokenType,
		HospitalID:            res.HospitalID,
		Role:                  roleSlug,
		AccessTokenExpiredAt:  res.AccessExp.UTC().Format(time.RFC3339),
		RefreshTokenExpiredAt: res.RefreshExp.UTC().Format(time.RFC3339),
	}
	util.HandleResponse(c, resp, nil)
}

func (ctl *Controller) Refresh(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	tokens, aexp, rexp, err := ctl.svc.Refresh(c.Request.Context(), strings.TrimSpace(req.RefreshToken)) // <=== changed
	if err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{
		"access_token":          tokens.AccessToken,
		"refresh_token":         tokens.RefreshToken,
		"accessTokenExpiredAt":  aexp.UTC().Format(time.RFC3339), // <=== added
		"refreshTokenExpiredAt": rexp.UTC().Format(time.RFC3339), // <=== added
	}
	util.HandleResponse(c, resp, nil)
}

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
	role := strings.TrimSpace(req.Role)
	if err := ctl.svc.ChooseRole(c.Request.Context(), userID, role); err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{"role": role}
	util.HandleResponse(c, resp, nil)
}

func (ctl *Controller) PasswordForgot(c *gin.Context) {
	var req request.PasswordForgotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	if err := ctl.svc.PasswordForgot(c.Request.Context(), &req); err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.Data = gin.H{"status": "pin_sent"}
	util.HandleResponse(c, resp, nil)
}

// POST /v1/auth/password/reset
func (ctl *Controller) PasswordReset(c *gin.Context) {
	var req request.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	if err := ctl.svc.PasswordReset(c.Request.Context(), &req); err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.Data = gin.H{"status": "password_updated"}
	util.HandleResponse(c, resp, nil)
}

// PUT /v1/auth/password  (AuthRequired)
func (ctl *Controller) PasswordChange(c *gin.Context) {
	var req request.PasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	userID := util.GetUserID(c) // ambil dari JWT (kita sudah pakai di tempat lain)
	if userID == "" {
		util.HandleError(c, constant.ErrUnauthorized)
		return
	}
	if err := ctl.svc.PasswordChange(c.Request.Context(), userID, &req); err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.Data = gin.H{"status": "password_changed"}
	util.HandleResponse(c, resp, nil)
}
