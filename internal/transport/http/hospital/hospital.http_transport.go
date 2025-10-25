package hospital

import (
	"github.com/api-monolith-template/internal/model/request"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/response"
	hs "github.com/api-monolith-template/internal/service/hospital"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct {
	svc *hs.Service
}

func NewController(s *hs.Service) *Controller { return &Controller{svc: s} }

func (ctl *Controller) CreateHospital(c *gin.Context) {
	var req request.CreateHospitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}

	h, err := ctl.svc.CreateHospital(c.Request.Context(), &req)
	if err != nil {
		util.HandleError(c, err)
		return
	}

	resp := response.NewResponseOK()
	resp.Data = gin.H{
		"id":          h.ID,
		"code":        h.Code,
		"name":        h.Name,
		"address":     h.Address,
		"city":        h.City,
		"province":    h.Province,
		"country":     h.Country,
		"latitude":    h.Latitude,
		"longitude":   h.Longitude,
		"phone":       h.Phone,
		"description": h.Description,
		"facilities":  h.Facilities,
		"is_active":   h.IsActive,
		"created_at":  h.CreatedAt,
	}
	util.HandleResponse(c, resp, nil)
}
func (ctl *Controller) CreateHospitalAdmin(c *gin.Context) {
	var req request.CreateHospitalAdminRequest
	// hospital_id boleh di path atau body; utamakan path
	if hid := c.Param("hospital_id"); hid != "" {
		req.HospitalID = hid
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}

	uid, err := ctl.svc.CreateHospitalAdmin(c.Request.Context(), req)
	if err != nil {
		util.HandleError(c, err)
		return
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusCreated
	resp.Data = gin.H{"user_id": uid, "role": constant.RoleAdmin}
	util.HandleResponse(c, resp, nil)
}

// POST /v1/hospitals/:hospital_id/staff  (admin tenant)
func (ctl *Controller) CreateHospitalStaff(c *gin.Context) {
	var req request.CreateHospitalStaffRequest
	if hid := c.Param("hospital_id"); hid != "" {
		req.HospitalID = hid
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}

	uid, err := ctl.svc.CreateHospitalStaff(c.Request.Context(), req)
	if err != nil {
		util.HandleError(c, err)
		return
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusCreated
	resp.Data = gin.H{"user_id": uid, "role": req.Role}
	util.HandleResponse(c, resp, nil)
}
