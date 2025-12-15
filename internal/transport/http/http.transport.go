package http

import (
	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	authCtrl "github.com/api-monolith-template/internal/transport/http/auth"
	userCtrl "github.com/api-monolith-template/internal/transport/http/user"
	warmupCtrl "github.com/api-monolith-template/internal/transport/http/warmup"
	transportmw "github.com/api-monolith-template/internal/transport/middleware"
	"github.com/api-monolith-template/internal/util"

	roleRepo "github.com/api-monolith-template/internal/repository/role"

	"github.com/api-monolith-template/internal/infrastructure" // <=== added (to get Redis client here)
)

type Transport struct {
	router           *gin.Engine
	authController   *authCtrl.Controller
	userController   *userCtrl.Controller
	warmupController *warmupCtrl.Controller

	roleRepo *roleRepo.Repository
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
func (t *Transport) WithRoleRepository(repo *roleRepo.Repository) *Transport {
	t.roleRepo = repo
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

		auth.POST("/refresh", t.authController.Refresh)
		auth.POST("/password/forgot", t.authController.PasswordForgot)
		auth.POST("/password/reset", t.authController.PasswordReset)
	}

	// Prepare Redis client for auth middleware (one instance here). // <=== added
	rdb := infrastructure.NewRedisClient() // <=== added

	protected := v1.Group("/")
	protected.Use(transportmw.AuthRequired(rdb))
	{
		protected.GET("/me", t.userController.Me)
		protected.PUT("/auth/password", t.authController.PasswordChange)
		protected.POST("/auth/logout", t.authController.Logout)
	}

	// 404
	t.router.NoRoute(func(c *gin.Context) {
		resp := constant.ErrEndpointNotFound.ToResponse()
		util.HandleResponse(c, &resp, nil)
	})
}
