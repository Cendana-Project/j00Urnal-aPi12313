package http

import (
	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	authCtrl "github.com/api-monolith-template/internal/transport/http/auth"
	hospCtrl "github.com/api-monolith-template/internal/transport/http/hospital"
	userCtrl "github.com/api-monolith-template/internal/transport/http/user"
	transportmw "github.com/api-monolith-template/internal/transport/middleware"
	"github.com/api-monolith-template/internal/util"

	hospRepo "github.com/api-monolith-template/internal/repository/hospital"
	roleRepo "github.com/api-monolith-template/internal/repository/role"
)

type Transport struct {
	router             *gin.Engine
	authController     *authCtrl.Controller
	userController     *userCtrl.Controller
	hospitalController *hospCtrl.Controller

	roleRepo *roleRepo.Repository
	hospRepo *hospRepo.Repository
}

func NewTransport() *Transport                              { return new(Transport) }
func (t *Transport) WithGinEngine(r *gin.Engine) *Transport { t.router = r; return t }
func (t *Transport) WithAuthController(c *authCtrl.Controller) *Transport {
	t.authController = c
	return t
}
func (t *Transport) WithUserController(c *userCtrl.Controller) *Transport {
	t.userController = c
	return t
}
func (t *Transport) WithHospitalController(c *hospCtrl.Controller) *Transport {
	t.hospitalController = c
	return t
}
func (t *Transport) WithRoleRepository(repo *roleRepo.Repository) *Transport {
	t.roleRepo = repo
	return t
}
func (t *Transport) WithHospitalRepository(repo *hospRepo.Repository) *Transport {
	t.hospRepo = repo
	return t
}

func (t *Transport) InitRoute() {
	if t.router == nil {
		panic("gin engine is nil")
	}
	// TIDAK memasang Auth di level router di sini.

	v1 := t.router.Group("/v1")

	// ========== AUTH — PUBLIC ==========
	auth := v1.Group("/auth")
	{
		auth.POST("/register", func(c *gin.Context) { t.authController.Register(c) })
		auth.POST("/resend-pin", func(c *gin.Context) { t.authController.ResendPIN(c) })
		auth.POST("/verify-pin", func(c *gin.Context) { t.authController.VerifyPIN(c) })

		// Split login
		auth.POST("/login", func(c *gin.Context) { t.authController.LoginPublic(c) })
		auth.POST("/login/hospital", func(c *gin.Context) { t.authController.LoginHospital(c) })

		auth.POST("/refresh", func(c *gin.Context) { t.authController.Refresh(c) })
	}

	// === PROTECTED — USER-LEVEL (tanpa tenant) ===
	protected := v1.Group("/")
	protected.Use(transportmw.AuthRequired())
	{
		protected.POST("/auth/choose-role", func(c *gin.Context) { t.authController.ChooseRole(c) })

		protected.PUT("/profile/patient",
			transportmw.RequirePermissions(t.roleRepo, constant.PermissionPatientEdit),
			func(c *gin.Context) { t.userController.UpdatePatientProfile(c) },
		)

		protected.PUT("/profile/doctor",
			transportmw.RequirePermissions(t.roleRepo, constant.PermissionDoctorEdit),
			func(c *gin.Context) { t.userController.UpdateDoctorProfile(c) },
		)

		protected.POST("/hospitals",
			transportmw.RequirePermissions(t.roleRepo, constant.PermissionUserCreate, constant.PermissionRoleAssign),
			func(c *gin.Context) { t.hospitalController.CreateHospital(c) },
		)
	}

	// === PROTECTED — HOSPITAL SCOPED (JWT + Tenant) ===
	tenant := v1.Group("/")
	tenant.Use(transportmw.AuthRequired(), transportmw.TenantContext())
	{
		tenant.POST("/hospitals/:hospital_id/admins",
			transportmw.RequireHospitalPermissions(t.hospRepo, t.roleRepo, constant.PermissionRoleAssign),
			func(c *gin.Context) { t.hospitalController.CreateHospitalAdmin(c) },
		)

		tenant.POST("/hospitals/:hospital_id/staff",
			transportmw.RequireHospitalPermissions(t.hospRepo, t.roleRepo, constant.PermissionUserCreate, constant.PermissionRoleAssign),
			func(c *gin.Context) { t.hospitalController.CreateHospitalStaff(c) },
		)
	}

	// 404
	t.router.NoRoute(func(c *gin.Context) {
		resp := constant.ErrEndpointNotFound.ToResponse()
		util.HandleResponse(c, &resp, nil)
	})
}
