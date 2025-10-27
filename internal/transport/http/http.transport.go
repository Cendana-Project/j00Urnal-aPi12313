package http

import (
	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	authCtrl "github.com/api-monolith-template/internal/transport/http/auth"
	hospCtrl "github.com/api-monolith-template/internal/transport/http/hospital"
	userCtrl "github.com/api-monolith-template/internal/transport/http/user"
	warmupCtrl "github.com/api-monolith-template/internal/transport/http/warmup"
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
	warmupController   *warmupCtrl.Controller

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
func (t *Transport) WithWarmupController(c *warmupCtrl.Controller) *Transport {
	t.warmupController = c
	return t
}

func (t *Transport) InitRoute() {
	if t.router == nil {
		panic("gin engine is nil")
	}
	// ========== WARMUP — PUBLIC ==========
	t.router.GET("/ping", func(c *gin.Context) { t.warmupController.Ping(c) })

	v1 := t.router.Group("/v1")

	// ========== AUTH — PUBLIC ==========
	auth := v1.Group("/auth")
	{
		auth.POST("/register", t.authController.Register)
		auth.POST("/resend-pin", t.authController.ResendPIN)
		auth.POST("/verify-pin", t.authController.VerifyPIN)

		auth.POST("/login", t.authController.LoginPublic)
		auth.POST("/login/hospital", t.authController.LoginHospital)

		auth.POST("/refresh", t.authController.Refresh)
		auth.POST("/password/forgot", t.authController.PasswordForgot)
		auth.POST("/password/reset", t.authController.PasswordReset)
	}

	// === PROTECTED — USER-LEVEL ===
	protected := v1.Group("/")
	protected.Use(transportmw.AuthRequired())
	{
		protected.POST("/auth/choose-role", t.authController.ChooseRole)

		protected.PUT("/profile/patient",
			transportmw.RequirePermissions(t.roleRepo, constant.PermissionPatientEdit),
			t.userController.UpdatePatientProfile,
		)

		protected.PUT("/auth/password", t.authController.PasswordChange)

		protected.PUT("/profile/doctor",
			transportmw.RequirePermissions(t.roleRepo, constant.PermissionDoctorEdit),
			t.userController.UpdateDoctorProfile,
		)

		protected.POST("/hospitals",
			transportmw.RequirePermissions(t.roleRepo, constant.PermissionUserCreate, constant.PermissionRoleAssign),
			t.hospitalController.CreateHospital,
		)

		protected.POST("/auth/set-profile", t.authController.SetProfile) // endpoint gabungan
	}

	// === PROTECTED — HOSPITAL SCOPED (JWT + Tenant) ===
	tenant := v1.Group("/")
	tenant.Use(transportmw.AuthRequired(), transportmw.TenantContext())
	{
		tenant.POST("/hospitals/:hospital_id/admins",
			transportmw.RequireHospitalPermissions(t.hospRepo, t.roleRepo, constant.PermissionRoleAssign),
			t.hospitalController.CreateHospitalAdmin,
		)

		tenant.POST("/hospitals/:hospital_id/staff",
			transportmw.RequireHospitalPermissions(t.hospRepo, t.roleRepo, constant.PermissionUserCreate, constant.PermissionRoleAssign),
			t.hospitalController.CreateHospitalStaff,
		)
	}

	// 404
	t.router.NoRoute(func(c *gin.Context) {
		resp := constant.ErrEndpointNotFound.ToResponse()
		util.HandleResponse(c, &resp, nil)
	})
}
