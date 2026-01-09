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

	"github.com/redis/go-redis/v9"

	issueCtrl "github.com/api-monolith-template/internal/transport/http/issue"
	journalCtrl "github.com/api-monolith-template/internal/transport/http/journal"
	manuscriptCtrl "github.com/api-monolith-template/internal/transport/http/manuscript"
	volumeCtrl "github.com/api-monolith-template/internal/transport/http/volume"
)

type Transport struct {
	router         *gin.Engine
	authController *authCtrl.Controller
	userController *userCtrl.Controller

	warmupController     *warmupCtrl.Controller
	journalController    *journalCtrl.Controller
	volumeController     *volumeCtrl.Controller
	issueController      *issueCtrl.Controller
	manuscriptController *manuscriptCtrl.Controller

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
func (t *Transport) WithJournalController(c *journalCtrl.Controller) *Transport {
	t.journalController = c
	return t
}
func (t *Transport) WithVolumeController(c *volumeCtrl.Controller) *Transport {
	t.volumeController = c
	return t
}
func (t *Transport) WithIssueController(c *issueCtrl.Controller) *Transport {
	t.issueController = c
	return t
}
func (t *Transport) WithManuscriptController(c *manuscriptCtrl.Controller) *Transport {
	t.manuscriptController = c
	return t
}

func (t *Transport) InitRoute(rdb *redis.Client) {
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

	// Prepare Redis client passed from bootstrap

	protected := v1.Group("/")
	protected.Use(transportmw.AuthRequired(rdb))
	{
		protected.GET("/me", t.userController.Me)
		protected.PUT("/auth/password", t.authController.PasswordChange)
		protected.POST("/auth/logout", t.authController.Logout)

		// Journals
		journals := protected.Group("/journals")
		{
			journals.POST("", t.journalController.Create)
			journals.GET("", t.journalController.GetAll)      // Public? Requirement: "Public can READ ACTIVE". So maybe move GetAll/Get to public group?
			journals.GET("/:id", t.journalController.GetByID) // Public?
			journals.PUT("/:id", t.journalController.Update)
			journals.PATCH("/:id/status", t.journalController.SetStatus)
			journals.POST("/:id/cover", t.journalController.UploadCover)

			// Volumes nested under journals
			journals.POST("/:id/volumes", t.volumeController.Create)
		}

		volumes := protected.Group("/volumes")
		{
			volumes.PUT("/:id", t.volumeController.Update)
			volumes.PATCH("/:id/status", t.volumeController.SetStatus)
			// Issues nested under volumes
			volumes.POST("/:id/issues", t.issueController.Create)
		}

		// Issues
		issues := protected.Group("/issues")
		{
			issues.PUT("/:id", t.issueController.UpdateMetadata)
			issues.PATCH("/:id/status", t.issueController.SetStatus)
			issues.POST("/:id/cover", t.issueController.UploadCover)
			issues.POST("/:id/full-pdf", t.issueController.UploadPDF)
		}

		manuscripts := protected.Group("/manuscripts")
		{
			manuscripts.POST("", t.manuscriptController.Create)
			manuscripts.GET("", t.manuscriptController.GetAll)
			manuscripts.GET("/:id", t.manuscriptController.GetByID)
			manuscripts.PUT("/:id", t.manuscriptController.Update)
			manuscripts.DELETE("/:id", t.manuscriptController.Delete)

			manuscripts.POST("/:id/authors", t.manuscriptController.UpdateAuthors)
			manuscripts.PUT("/:id/authors", t.manuscriptController.UpdateAuthors)

			manuscripts.POST("/:id/files/main", t.manuscriptController.UploadMainFile)
			manuscripts.POST("/:id/files/attachment", t.manuscriptController.UploadAttachment)
		}

		// Manuscripts (Admin Only)
		admin := protected.Group("/admin")
		admin.Use(transportmw.RequirePermissions(t.roleRepo, constant.PermissionManuscriptManage))
		{

		}
	}

	// 404
	t.router.NoRoute(func(c *gin.Context) {
		resp := constant.ErrEndpointNotFound.ToResponse()
		util.HandleResponse(c, &resp, nil)
	})
}
