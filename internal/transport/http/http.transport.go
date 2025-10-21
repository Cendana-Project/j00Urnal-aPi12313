package http

import (
	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	authHttp "github.com/api-monolith-template/internal/transport/http/auth"
	transportmw "github.com/api-monolith-template/internal/transport/middleware"
	"github.com/api-monolith-template/internal/util"
)

type Transport struct {
	router         *gin.Engine
	authController *authHttp.Controller
}

func NewTransport() *Transport { return new(Transport) }

func (t *Transport) WithGinEngine(r *gin.Engine) *Transport {
	t.router = r
	return t
}

func (t *Transport) WithAuthController(c *authHttp.Controller) *Transport {
	t.authController = c
	return t
}

// InitRoute mendaftarkan route publik & middleware global.
// Catatan: healthcheck /_internal/healthz sudah disediakan oleh infrastructure.NewGinEngine(),
// jadi TIDAK perlu didaftarkan ulang di sini.
func (t *Transport) InitRoute() {
	if t.router == nil {
		panic("gin engine is nil: call WithGinEngine before InitRoute")
	}

	// === Middleware global tambahan ===
	// (Recovery kemungkinan sudah dipasang di NewGinEngine; aman jika dipasang lagi, tapi tidak wajib.)
	t.router.Use(transportmw.TraceID()) // generate & propagate trace id

	// === API v1 ===
	v1 := t.router.Group("/v1")

	// --- AUTH (PUBLIC) ---
	auth := v1.Group("/auth")
	{
		auth.POST("/register", func(c *gin.Context) { t.authController.Register(c) })
		auth.POST("/verify-pin", func(c *gin.Context) { t.authController.VerifyPIN(c) })
	}

	// === 404 handler ===
	t.router.NoRoute(func(c *gin.Context) {
		resp := constant.ErrEndpointNotFound.ToResponse()
		// TraceID sudah di-inject oleh middleware TraceID melalui util.HandleResponse di controller/handler error
		util.HandleResponse(c, &resp, nil)
	})
}
