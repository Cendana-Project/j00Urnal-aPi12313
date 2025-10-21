package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/model/contract/service"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct {
	svc service.AuthService
}

func NewController(svc service.AuthService) *Controller { return &Controller{svc: svc} }

// helper ambil trace id yang sudah diset middleware (header <-> writer)
func traceID(c *gin.Context) string {
	if v := c.Writer.Header().Get("X-Request-Id"); v != "" {
		return v
	}
	if v := c.GetHeader("X-Request-Id"); v != "" {
		return v
	}
	return ""
}

// POST /v1/auth/register
func (ctl *Controller) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	out, err := ctl.svc.Register(c.Request.Context(), &req)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = out
	resp.TraceID = traceID(c) // <-- pastikan terisi
	util.HandleResponse(c, resp, nil)
}

// POST /v1/auth/verify-pin
func (ctl *Controller) VerifyPIN(c *gin.Context) {
	var req request.VerifyPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	out, err := ctl.svc.VerifyPIN(c.Request.Context(), req.Email, req.PIN)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = out
	util.HandleResponse(c, resp, nil)
}

type simpleErr string

func (e simpleErr) Error() string { return string(e) }
