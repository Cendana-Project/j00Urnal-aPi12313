package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/service/auth"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct{ svc *auth.Service }

func NewController(svc *auth.Service) *Controller { return &Controller{svc: svc} }

// PUT /v1/profile/patient (protected)
func (ctl *Controller) UpdatePatientProfile(c *gin.Context) {
	var req request.PatientProfileRequest
	if err := util.BindAndValidate(c, &req); err != nil {
		util.HandleError(c, err)
		return
	}
	userID := util.GetUserID(c)
	if err := ctl.svc.CompletePatientProfile(c.Request.Context(), userID, &req); err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{"updated": true}
	util.HandleResponse(c, resp, nil)
}

// PUT /v1/profile/doctor (protected)
func (ctl *Controller) UpdateDoctorProfile(c *gin.Context) {
	var req request.DoctorProfileRequest
	if err := util.BindAndValidate(c, &req); err != nil {
		util.HandleError(c, err)
		return
	}
	userID := util.GetUserID(c)
	if err := ctl.svc.CompleteDoctorProfile(c.Request.Context(), userID, &req); err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = gin.H{"updated": true}
	util.HandleResponse(c, resp, nil)
}
