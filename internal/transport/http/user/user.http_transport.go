package user

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	userrepo "github.com/api-monolith-template/internal/repository/user" // <=== added
	"github.com/api-monolith-template/internal/service/auth"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct {
	svc      *auth.Service
	userRepo *userrepo.Repository // <=== added
}

func NewController(svc *auth.Service, ur *userrepo.Repository) *Controller { // <=== changed
	return &Controller{svc: svc, userRepo: ur}
}

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
	roleSlug = strings.ToUpper(roleSlug) // <=== added
	dto := toMeDTO(u, roleSlug)

	// Enrichment: kalau role cocok ATAU data memang ada
	// PATIENT
	if roleSlug == constant.RolePatient {
		h, w, a, m, err := ctl.userRepo.GetPatientProfileByUserID(userID)
		if err != nil {
			util.HandleError(c, err)
			return
		}
		if h != nil || w != nil || a != nil || m != nil {
			dto.PatientProfile = &response.PatientProfile{HeightCM: h, WeightKG: w, Allergies: a, MedicalHist: m}
		}
	} else {
		// role bukan PATIENT, tapi kalau tabelnya ada, tetap boleh tampilkan (opsional)
		if h, w, a, m, err := ctl.userRepo.GetPatientProfileByUserID(userID); err == nil {
			if h != nil || w != nil || a != nil || m != nil {
				dto.PatientProfile = &response.PatientProfile{HeightCM: h, WeightKG: w, Allergies: a, MedicalHist: m}
			}
		}
	}

	// DOCTOR
	if roleSlug == constant.RoleDoctor {
		sip, spec, err := ctl.userRepo.GetDoctorProfileByUserID(userID)
		if err != nil {
			util.HandleError(c, err)
			return
		}
		if sip != nil || spec != nil {
			dto.DoctorProfile = &response.DoctorProfile{SIPNumber: sip, Specialty: spec}
		}
	} else {
		// role bukan DOCTOR, tapi kalau profil ada, tetap boleh tampilkan (opsional)
		if sip, spec, err := ctl.userRepo.GetDoctorProfileByUserID(userID); err == nil {
			if sip != nil || spec != nil {
				dto.DoctorProfile = &response.DoctorProfile{SIPNumber: sip, Specialty: spec}
			}
		}
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = dto
	util.HandleResponse(c, resp, nil)
}

// GET /v1/tenant/me (tenant-scoped; wajib hint & membership)
func (ctl *Controller) TenantMe(c *gin.Context) {
	userUUID, err := util.GetUserIDFromContext(c)
	if err != nil || userUUID == nil {
		util.HandleError(c, constant.ErrUserNotAuthenticated)
		return
	}
	userID := userUUID.String()

	hintVal, ok := c.Get("hospital_hint")
	if !ok {
		util.HandleError(c, constant.ErrValidationFailed)
		return
	}
	hint := strings.TrimSpace(hintVal.(string))
	if hint == "" {
		util.HandleError(c, constant.ErrValidationFailed)
		return
	}

	// resolve hospital
	hosp, err := ctl.userRepo.ResolveHospitalHint(hint)
	if err != nil {
		util.HandleError(c, constant.ErrHospitalNotFound)
		return
	}

	// cek membership
	isMember, err := ctl.userRepo.IsMemberOfHospital(userID, hosp.ID)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	if !isMember {
		util.HandleError(c, constant.ErrUserNotLinkedToHospital)
		return
	}

	u, err := ctl.userRepo.GetByID(userID)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	if u == nil {
		util.HandleError(c, constant.ErrUserNotFound)
		return
	}

	// role scoped hospital (UPPER)
	roleSlug, err := ctl.userRepo.GetHospitalRoleSlug(userID, hosp.ID)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	roleSlug = strings.ToUpper(roleSlug) // <=== added
	if roleSlug == "" {
		util.HandleError(c, constant.ErrForbidden)
		return
	}

	dto := toMeDTO(u, roleSlug)
	dto.Hospitals = []response.HospitalBrief{{ID: hosp.ID, Code: hosp.Code, Name: hosp.Name}}

	// Enrich sesuai role scoped
	switch roleSlug {
	case constant.RoleDoctor:
		if sip, spec, err := ctl.userRepo.GetDoctorProfileByUserID(userID); err == nil && (sip != nil || spec != nil) {
			dto.DoctorProfile = &response.DoctorProfile{SIPNumber: sip, Specialty: spec}
		}
	case constant.RolePatient:
		if h, w, a, m, err := ctl.userRepo.GetPatientProfileByUserID(userID); err == nil && (h != nil || w != nil || a != nil || m != nil) {
			dto.PatientProfile = &response.PatientProfile{HeightCM: h, WeightKG: w, Allergies: a, MedicalHist: m}
		}
	}

	resp := response.NewResponseOK()
	resp.StatusCode = http.StatusOK
	resp.Data = dto
	util.HandleResponse(c, resp, nil)
}

func toMeDTO(u *entity.User, roleSlug string) response.MeResponse {
	var dobStr *string
	if u.DOB != nil {
		s := u.DOB.Format("2006-01-02")
		dobStr = &s
	}
	var verifiedStr *string
	if u.VerifiedAt != nil {
		s := u.VerifiedAt.UTC().Format(time.RFC3339)
		verifiedStr = &s
	}
	return response.MeResponse{
		ID:         u.ID,
		Email:      u.Email,
		Username:   u.Username,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Phone:      u.Phone,
		Gender:     u.Gender,
		DOB:        dobStr,
		Address:    u.Address,
		Status:     u.Status,
		VerifiedAt: verifiedStr,
		Role:       roleSlug, // sudah dinormalisasi UPPER di pemanggil
	}
}
