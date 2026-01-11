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

	// ========== PUBLIC READ ACCESS ==========
	v1.GET("/journals/:id", t.journalController.GetByID)
	v1.GET("/volumes/:id", t.volumeController.GetByID)
	v1.GET("/issues/:id", t.issueController.GetByID)

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
		protected.DELETE("/users/:id", t.userController.Delete) // TODO: Add specific permission check?

		// Journal Management Permission Helper
		requireJournalManage := transportmw.RequirePermissions(t.roleRepo, constant.PermissionJournalManage)

		// Journals
		journals := protected.Group("/journals")
		{
			journals.POST("", requireJournalManage, t.journalController.Create)
			journals.GET("", t.journalController.GetAll)
			journals.PUT("/:id", requireJournalManage, t.journalController.Update)
			journals.PATCH("/:id/status", requireJournalManage, t.journalController.SetStatus)
			journals.POST("/:id/cover", requireJournalManage, t.journalController.UploadCover)
			journals.DELETE("/:id", requireJournalManage, t.journalController.Delete)

			// Volumes nested under journals
			journals.POST("/:id/volumes", requireJournalManage, t.volumeController.Create)
		}

		volumes := protected.Group("/volumes")
		{
			volumes.PUT("/:id", requireJournalManage, t.volumeController.Update)
			volumes.PATCH("/:id/status", requireJournalManage, t.volumeController.SetStatus)
			volumes.DELETE("/:id", requireJournalManage, t.volumeController.Delete)
			// Issues nested under volumes
			volumes.POST("/:id/issues", requireJournalManage, t.issueController.Create)
		}

		// Issues
		issues := protected.Group("/issues")
		{
			issues.PUT("/:id", requireJournalManage, t.issueController.UpdateMetadata)
			issues.PATCH("/:id/status", requireJournalManage, t.issueController.SetStatus)
			issues.POST("/:id/cover", requireJournalManage, t.issueController.UploadCover)
			issues.POST("/:id/full-pdf", requireJournalManage, t.issueController.UploadPDF)
			issues.DELETE("/:id", requireJournalManage, t.issueController.Delete)
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
