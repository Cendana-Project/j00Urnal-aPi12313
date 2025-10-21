package http

import (
	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	authCtrl "github.com/api-monolith-template/internal/transport/http/auth"
	userCtrl "github.com/api-monolith-template/internal/transport/http/user"
	transportmw "github.com/api-monolith-template/internal/transport/middleware"
	"github.com/api-monolith-template/internal/util"

	rolerepo "github.com/api-monolith-template/internal/repository/role"
)

type Transport struct {
	router         *gin.Engine
	authController *authCtrl.Controller
	userController *userCtrl.Controller

	// >>> tambahkan ini
	roleRepo *rolerepo.Repository
}

func NewTransport() *Transport { return new(Transport) }

func (t *Transport) WithGinEngine(r *gin.Engine) *Transport { t.router = r; return t }
func (t *Transport) WithAuthController(c *authCtrl.Controller) *Transport {
	t.authController = c
	return t
}
func (t *Transport) WithUserController(c *userCtrl.Controller) *Transport {
	t.userController = c
	return t
}

// >>> builder baru untuk inject role repo
func (t *Transport) WithRoleRepository(repo *rolerepo.Repository) *Transport {
	t.roleRepo = repo
	return t
}

func (t *Transport) InitRoute() {
	if t.router == nil {
		panic("gin engine is nil")
	}

	// middlewares
	t.router.Use(transportmw.TraceID())

	v1 := t.router.Group("/v1")

	// AUTH — public
	auth := v1.Group("/auth")
	{
		auth.POST("/register", func(c *gin.Context) { t.authController.Register(c) })
		auth.POST("/resend-pin", func(c *gin.Context) { t.authController.ResendPIN(c) })
		auth.POST("/verify-pin", func(c *gin.Context) { t.authController.VerifyPIN(c) })
		auth.POST("/login", func(c *gin.Context) { t.authController.Login(c) })
		auth.POST("/refresh", func(c *gin.Context) { t.authController.Refresh(c) })
	}

	// Protected
	protected := v1.Group("/")
	protected.Use(transportmw.AuthRequired())
	{
		protected.POST("/auth/choose-role", func(c *gin.Context) { t.authController.ChooseRole(c) })

		// >>> TANAM PERMISSION DI SINI
		protected.PUT("/profile/patient",
			transportmw.RequirePermissions(t.roleRepo, constant.PermissionPatientEdit),
			func(c *gin.Context) { t.userController.UpdatePatientProfile(c) },
		)
		protected.PUT("/profile/doctor",
			transportmw.RequirePermissions(t.roleRepo, constant.PermissionDoctorEdit),
			func(c *gin.Context) { t.userController.UpdateDoctorProfile(c) },
		)
	}

	// 404
	t.router.NoRoute(func(c *gin.Context) {
		resp := constant.ErrEndpointNotFound.ToResponse()
		util.HandleResponse(c, &resp, nil)
	})
}
