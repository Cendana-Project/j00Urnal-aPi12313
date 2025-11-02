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

	"github.com/api-monolith-template/internal/infrastructure" // <=== added (to get Redis client here)
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

	// Prepare Redis client for auth middleware (one instance here). // <=== added
	rdb := infrastructure.NewRedisClient() // <=== added

	// === PROTECTED — USER-LEVEL ===
	protected := v1.Group("/")
	protected.Use(transportmw.AuthRequired(rdb)) // <=== changed: pass rdb
	{
		protected.GET("/me", t.userController.Me) // <=== added

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

		// === NEW: Logout endpoints ===
		protected.POST("/auth/logout", t.authController.Logout)
	}

	// === PROTECTED — HOSPITAL SCOPED (JWT + Tenant) ===
	tenant := v1.Group("/")
	tenant.Use(transportmw.AuthRequired(rdb), transportmw.TenantContext()) // <=== changed: pass rdb
	{
		tenant.POST("/hospitals/:hospital_id/admins",
			transportmw.RequireHospitalPermissions(t.hospRepo, t.roleRepo, constant.PermissionRoleAssign),
			t.hospitalController.CreateHospitalAdmin,
		)

		tenant.POST("/hospitals/:hospital_id/staff",
			transportmw.RequireHospitalAdminOrSuper(t.hospRepo, t.roleRepo), // <=== changed: enforce hospital admin or super admin
			t.hospitalController.CreateHospitalStaff,
		)

		tenant.GET("/tenant/me", t.userController.TenantMe) // <=== added
	}

	// 404
	t.router.NoRoute(func(c *gin.Context) {
		resp := constant.ErrEndpointNotFound.ToResponse()
		util.HandleResponse(c, &resp, nil)
	})
}
