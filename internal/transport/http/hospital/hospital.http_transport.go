package hospital

import (
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

// POST /v1/hospitals  (super_admin only)
type createHospitalReq struct {
	Code string `json:"code"`
	Name string `json:"name" binding:"required"`
}

func (ctl *Controller) CreateHospital(c *gin.Context) {
	var req createHospitalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}
	id, err := ctl.svc.CreateHospital(c.Request.Context(), hs.CreateHospitalReq{
		Code: req.Code, Name: req.Name,
	})
	if err != nil {
		util.HandleError(c, err)
		return
	}
	resp := response.NewResponseOK()
	resp.Data = gin.H{"hospital_id": id}
	util.HandleResponse(c, resp, nil)
}

// POST /v1/hospitals/:hospital_id/admins  (super_admin)
type createAdminReq struct {
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"`
}

func (ctl *Controller) CreateHospitalAdmin(c *gin.Context) {
	var req createAdminReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}

	id, err := ctl.svc.CreateHospitalAdmin(c.Request.Context(), hs.CreateHospitalAdminReq{
		HospitalID: c.Param("hospital_id"),
		Email:      req.Email, Phone: req.Phone, Password: req.Password,
		FirstName: req.FirstName, LastName: req.LastName,
	})
	if err != nil {
		util.HandleError(c, err)
		return
	}

	resp := response.NewResponseOK()
	resp.Data = gin.H{"user_id": id, "role": constant.RoleAdmin}
	util.HandleResponse(c, resp, nil)
}

// POST /v1/hospitals/:hospital_id/staff  (admin tenant)
type createStaffReq struct {
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"`
	Role      string `json:"role" binding:"required"` // doctor|nurse|receptionist|bod|admin
}

func (ctl *Controller) CreateHospitalStaff(c *gin.Context) {
	var req createStaffReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.HandleError(c, err)
		return
	}

	uid, err := ctl.svc.CreateHospitalStaff(c.Request.Context(), hs.CreateHospitalStaffReq{
		HospitalHint: c.Param("hospital_id"),
		Email:        req.Email, Phone: req.Phone, Password: req.Password,
		FirstName: req.FirstName, LastName: req.LastName,
		RoleSlug: req.Role,
	})
	if err != nil {
		util.HandleError(c, err)
		return
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusCreated
	resp.Data = gin.H{"user_id": uid, "role": req.Role}
	util.HandleResponse(c, resp, nil)
}
