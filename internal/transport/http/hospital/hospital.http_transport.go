package hospital

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/request"
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

// POST /v1/hospitals/:hospital_id/admins
func (ctl *Controller) CreateHospitalAdmin(c *gin.Context) {
	// Bind URI ke struct kecil agar validasi tidak mengenai field JSON. // <=== changed
	var uri struct {
		HospitalID string `uri:"hospital_id" binding:"required"` // tambahkan ,uuid jika perlu
	}
	if err := c.ShouldBindUri(&uri); err != nil { // <=== changed
		util.HandleError(c, err)
		return
	}

	var req request.CreateHospitalAdminRequest
	req.HospitalID = uri.HospitalID // ambil dari path // <=== changed
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}

	// Normalisasi gender (opsional pengaman) // <=== changed
	if req.Gender != nil {
		g := strings.ToUpper(strings.TrimSpace(*req.Gender))
		req.Gender = &g
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

// POST /v1/hospitals/:hospital_id/staff
func (ctl *Controller) CreateHospitalStaff(c *gin.Context) {
	// 1) Bind hanya URI ke struct kecil → hindari validasi field JSON saat tahap ini. // <=== changed
	var uri struct {
		HospitalID string `uri:"hospital_id" binding:"required"` // tambahkan ,uuid jika perlu
	}
	if err := c.ShouldBindUri(&uri); err != nil { // <=== changed
		util.HandleError(c, err)
		return
	}

	// 2) Bind JSON ke DTO utama, isi HospitalID dari URI. // <=== changed
	var req request.CreateHospitalStaffRequest
	req.HospitalID = uri.HospitalID                // <=== changed
	if err := c.ShouldBindJSON(&req); err != nil { // <=== changed
		util.HandleError(c, err)
		return
	}

	// 3) Normalisasi role & gender supaya konsisten dengan validator/DB. // <=== changed
	req.Role = strings.ToUpper(strings.TrimSpace(req.Role))
	if req.Gender != nil {
		g := strings.ToUpper(strings.TrimSpace(*req.Gender))
		req.Gender = &g
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
