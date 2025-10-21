package http

import (
	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	authCtrl "github.com/api-monolith-template/internal/transport/http/auth"
	userCtrl "github.com/api-monolith-template/internal/transport/http/user"
	transportmw "github.com/api-monolith-template/internal/transport/middleware"
	"github.com/api-monolith-template/internal/util"
)

type Transport struct {
	router         *gin.Engine
	authController *authCtrl.Controller
	userController *userCtrl.Controller
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

func (t *Transport) InitRoute() {
	if t.router == nil {
		panic("gin engine is nil")
	}

	// middlewares
	t.router.Use(transportmw.TraceID())

	// health
	//t.router.GET("/_internal/healthz", func(c *gin.Context) {
	//	resp := response.NewResponseOK()
	//	resp.Data = gin.H{"status": "ok"}
	//	util.HandleResponse(c, resp, nil)
	//})

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
		protected.PUT("/profile/patient", func(c *gin.Context) { t.userController.UpdatePatientProfile(c) })
		protected.PUT("/profile/doctor", func(c *gin.Context) { t.userController.UpdateDoctorProfile(c) })
	}

	// 404
	t.router.NoRoute(func(c *gin.Context) {
		resp := constant.ErrEndpointNotFound.ToResponse()
		util.HandleResponse(c, &resp, nil)
	})
}
