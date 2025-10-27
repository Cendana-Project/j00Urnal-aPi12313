package auth

import (
	"encoding/json"
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
	if err := util.BindAndValidate(c, &req); err != nil {
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
	if err := util.BindAndValidate(c, &req); err != nil {
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
	if err := util.BindAndValidate(c, &req); err != nil {
		util.HandleError(c, err)
		return
	}
	email := strings.TrimSpace(req.Email)
	pin := strings.TrimSpace(req.PIN)
	tokens, aexp, rexp, err := ctl.svc.VerifyPIN(c.Request.Context(), email, pin)
	if err != nil {
		util.HandleError(c, err)
		return
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{
		"access_token":             tokens.AccessToken,
		"refresh_token":            tokens.RefreshToken,
		"access_token_expired_at":  aexp.UTC().Format(time.RFC3339),
		"refresh_token_expired_at": rexp.UTC().Format(time.RFC3339),
	}
	util.HandleResponse(c, resp, nil)
}

// =============================
// AUTH — LOGIN
// =============================

func (ctl *Controller) LoginPublic(c *gin.Context) {
	var req request.LoginRequest
	if err := util.BindAndValidate(c, &req); err != nil {
		util.HandleError(c, err)
		return
	}
	identity := strings.TrimSpace(req.Identity)
	password := strings.TrimSpace(req.Password)

	tokens, roles, aexp, rexp, err := ctl.svc.Login(c.Request.Context(), identity, password)
	if err != nil {
		util.HandleError(c, err)
		return
	}

	roleSlug := ""
	if len(roles) > 0 {
		roleSlug = roles[0].Slug
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = response.LoginResponse{
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
	if err := util.BindAndValidate(c, &req); err != nil {
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

	roleSlug := ""
	if len(res.Roles) > 0 {
		roleSlug = res.Roles[0].Slug
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = response.LoginHospitalResponse{
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
	if err := util.BindAndValidate(c, &req); err != nil {
		util.HandleError(c, err)
		return
	}
	tokens, aexp, rexp, err := ctl.svc.Refresh(c.Request.Context(), strings.TrimSpace(req.RefreshToken))
	if err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{
		"access_token":          tokens.AccessToken,
		"refresh_token":         tokens.RefreshToken,
		"accessTokenExpiredAt":  aexp.UTC().Format(time.RFC3339),
		"refreshTokenExpiredAt": rexp.UTC().Format(time.RFC3339),
	}
	util.HandleResponse(c, resp, nil)
}

func (ctl *Controller) ChooseRole(c *gin.Context) {
	userID := util.GetUserID(c)
	if userID == "" {
		res := constant.ErrUnauthorized.ToResponse()
		util.HandleResponse(c, &res, nil)
		return
	}
	var req request.ChooseRoleRequest
	if err := util.BindAndValidate(c, &req); err != nil {
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
	if err := util.BindAndValidate(c, &req); err != nil {
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

func (ctl *Controller) PasswordReset(c *gin.Context) {
	var req request.PasswordResetRequest
	if err := util.BindAndValidate(c, &req); err != nil {
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

func (ctl *Controller) PasswordChange(c *gin.Context) {
	var req request.PasswordChangeRequest
	if err := util.BindAndValidate(c, &req); err != nil {
		util.HandleError(c, err)
		return
	}
	userID := util.GetUserID(c)
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

// === SET PROFILE (gabungan choose-role + set profile) ===

func (ctl *Controller) SetProfile(c *gin.Context) {
	var req request.SetProfileRequest
	if err := util.BindAndValidate(c, &req); err != nil {
		util.HandleError(c, err)
		return
	}
	userID := util.GetUserID(c)
	if userID == "" {
		util.HandleError(c, constant.ErrUnauthorized)
		return
	}

	// Safety normalize role ke UPPERCASE di service juga.
	req.Role = strings.ToUpper(strings.TrimSpace(req.Role))

	// Pastikan profile non-null dan JSON well-formed
	if req.Profile == nil || len(*req.Profile) == 0 || string(*req.Profile) == "null" {
		util.HandleError(c, constant.ErrValidationError)
		return
	}
	var tmp json.RawMessage
	if err := json.Unmarshal(*req.Profile, &tmp); err != nil {
		util.HandleError(c, constant.ErrValidationError)
		return
	}

	res, err := ctl.svc.SetProfile(c.Request.Context(), userID, req.Role, req.Profile)
	if err != nil {
		util.HandleError(c, err)
		return
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = res
	util.HandleResponse(c, resp, nil)
}

// =============================
// AUTH — LOGOUT (new)
// =============================

func (ctl *Controller) Logout(c *gin.Context) { // <=== added
	// body optional: { "refresh_token": "..." }
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&body)

	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
		util.HandleError(c, constant.ErrUnauthorized)
		return
	}
	access := strings.TrimSpace(h[7:])

	if err := ctl.svc.Logout(c.Request.Context(), access, strings.TrimSpace(body.RefreshToken)); err != nil {
		util.HandleError(c, err)
		return
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{
		"status":          "logout_success",
		"revoked_refresh": body.RefreshToken != "",
	}
	util.HandleResponse(c, resp, nil)
}
