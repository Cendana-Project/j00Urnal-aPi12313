package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	userrepo "github.com/api-monolith-template/internal/repository/user"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
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

// =======================================
// AUTH — PUBLIC REGISTRATION & OTP FLOW
// =======================================

func (ctl *Controller) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := util.BindAndValidate(c, &req); err != nil {
		util.HandleError(c, err)
		return
	}
	u, err := ctl.svc.Register(c.Request.Context(), &req)
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

	// Samakan struktur dengan LoginPublic: gunakan response.LoginResponse  // <=== changed
	roleSlug := "" // Jika ingin ada role, tambahkan pengambilan role di service VerifyPIN  // <=== changed

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

	// --- Ambil user_id (sub) dari access token baru, lalu dapatkan role slug --- // <=== added
	userID, _ := extractSubFromJWT(tokens.AccessToken)
	roleSlug := ""
	if userID != "" && ctl.userRepo != nil {
		if slug, err := ctl.userRepo.GetUserRoleSlug(userID); err == nil {
			roleSlug = slug
		}
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = response.LoginResponse{
		AccessToken:           tokens.AccessToken,
		RefreshToken:          tokens.RefreshToken,
		Role:                  roleSlug, // <=== now included
		AccessTokenExpiredAt:  aexp.UTC().Format(time.RFC3339),
		RefreshTokenExpiredAt: rexp.UTC().Format(time.RFC3339),
	}
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

// extractSubFromJWT mengekstrak claim "sub" dari JWT tanpa verifikasi signature.
// Aman untuk use-case ini karena token baru saja kita terbitkan di sisi server.  // <=== added
func extractSubFromJWT(tok string) (string, error) {
	parts := strings.Split(tok, ".")
	if len(parts) < 2 {
		return "", errors.New("invalid jwt")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	var claims struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", err
	}
	return strings.TrimSpace(claims.Sub), nil
}
