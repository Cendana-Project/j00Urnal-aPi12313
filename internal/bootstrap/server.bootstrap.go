package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/email"
	"github.com/api-monolith-template/internal/infrastructure"
	roleRepo "github.com/api-monolith-template/internal/repository/role"
	userRepo "github.com/api-monolith-template/internal/repository/user"
	authSvc "github.com/api-monolith-template/internal/service/auth"
	httpTransport "github.com/api-monolith-template/internal/transport/http"
	authHttp "github.com/api-monolith-template/internal/transport/http/auth"
	userHttp "github.com/api-monolith-template/internal/transport/http/user"
	warmupHttp "github.com/api-monolith-template/internal/transport/http/warmup"
	"github.com/api-monolith-template/internal/util"
	"github.com/sirupsen/logrus"

	issueRepo "github.com/api-monolith-template/internal/repository/issue"
	journalRepo "github.com/api-monolith-template/internal/repository/journal"
	manuscriptRepo "github.com/api-monolith-template/internal/repository/manuscript"
	fileRepo "github.com/api-monolith-template/internal/repository/publicationfile"
	reviewRepo "github.com/api-monolith-template/internal/repository/review"
	termRepo "github.com/api-monolith-template/internal/repository/term"
	volumeRepo "github.com/api-monolith-template/internal/repository/volume"

	issueSvc "github.com/api-monolith-template/internal/service/issue"
	journalSvc "github.com/api-monolith-template/internal/service/journal"
	manuscriptSvc "github.com/api-monolith-template/internal/service/manuscript"
	reviewSvc "github.com/api-monolith-template/internal/service/review"
	storageSvc "github.com/api-monolith-template/internal/service/storage"
	termSvc "github.com/api-monolith-template/internal/service/term"
	volumeSvc "github.com/api-monolith-template/internal/service/volume"

	issueHttp "github.com/api-monolith-template/internal/transport/http/issue"
	journalHttp "github.com/api-monolith-template/internal/transport/http/journal"
	manuscriptHttp "github.com/api-monolith-template/internal/transport/http/manuscript"
	reviewHttp "github.com/api-monolith-template/internal/transport/http/review"
	termHttp "github.com/api-monolith-template/internal/transport/http/term"
	volumeHttp "github.com/api-monolith-template/internal/transport/http/volume"
)

func StartServer() {
	ctx := context.Background()

	// Infra
	gormDB := infrastructure.InitializeDBConn()
	rdb := infrastructure.NewRedisClient()
	if !config.Env.Redis.IsCacheDisable && rdb != nil {
		_, err := rdb.Ping(ctx).Result()
		util.ContinueOrFatal(err)
	}

	// Get database connection for shutdown
	db, err := gormDB.DB()
	util.ContinueOrFatal(err)
	err = db.Ping()
	util.ContinueOrFatal(err)

	r := infrastructure.NewGinEngine()

	// Repositories
	uRepo := userRepo.NewRepository(gormDB)
	rRepo := roleRepo.NewRepository(gormDB)

	jRepo := journalRepo.NewRepository(gormDB)
	vRepo := volumeRepo.NewRepository(gormDB)
	iRepo := issueRepo.NewRepository(gormDB)
	mRepo := manuscriptRepo.NewRepository(gormDB)
	fRepo := fileRepo.NewRepository(gormDB)
	tRepo := termRepo.NewRepository(gormDB)
	rvRepo := reviewRepo.NewRepository(gormDB)

	// SMTP sender config (fallback default)
	host := config.Env.SMTP.Host
	if host == "" {
		host = "smtp.gmail.com"
	}
	port := config.Env.SMTP.Port
	if port == 0 {
		port = 587
	}

	// Email sender configuration
	// Allow disabling email via EMAIL_ENABLED env var
	emailEnabled := true
	if v := os.Getenv("EMAIL_ENABLED"); v == "false" || v == "0" {
		emailEnabled = false
		logrus.Warn("Email sending is DISABLED via EMAIL_ENABLED env var")
	}

	// Configure email timeout (default 10s, override via EMAIL_TIMEOUT_SECONDS)
	timeoutSeconds := 10
	if v := os.Getenv("EMAIL_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeoutSeconds = n
		}
	}

	// Email provider selection: "brevo-api", "smtp" (default)
	// Use EMAIL_PROVIDER env var to choose
	provider := os.Getenv("EMAIL_PROVIDER")
	logrus.Infof("EMAIL_PROVIDER env var value: '%s'", provider) // Debug log
	if provider == "" {
		provider = "smtp" // default to SMTP for backward compatibility
		logrus.Info("EMAIL_PROVIDER not set, defaulting to 'smtp'")
	}

	// Auto-detect Brevo if using smtp-relay.brevo.com
	smtpHost := config.Env.SMTP.Host
	if strings.Contains(strings.ToLower(smtpHost), "brevo") || strings.Contains(strings.ToLower(smtpHost), "sendinblue") {
		logrus.Info("Brevo/Sendinblue SMTP server detected")
	}

	var sender email.Sender

	if !emailEnabled {
		// Email disabled - use nil sender (registration will auto-activate users)
		sender = nil
		logrus.Info("Email service is DISABLED - users will be auto-activated")
	} else if provider == "brevo-api" {
		// Brevo HTTP API (recommended for cloud environments - no SMTP port blocking)
		apiKey := config.Env.SMTP.Password // Reuse password field for API key
		fromEmail := config.Env.SMTP.FromEmail
		if fromEmail == "" {
			fromEmail = "no-reply@medikaone.id"
		}
		sender = email.NewBrevoAPISender(apiKey, fromEmail, time.Duration(timeoutSeconds)*time.Second)
		logrus.Info("Using Brevo HTTP API for email (port 443 - no SMTP blocking)")
	} else {
		// Traditional SMTP sender (may be blocked in cloud environments)
		host := config.Env.SMTP.Host
		if host == "" {
			host = "smtp.gmail.com"
		}
		port := config.Env.SMTP.Port
		if port == 0 {
			port = 587
		}
		username := config.Env.SMTP.Username
		password := config.Env.SMTP.Password
		fromEmail := config.Env.SMTP.FromEmail
		if fromEmail == "" {
			fromEmail = "no-reply@medikaone.id"
		}

		// Auto-detect STARTTLS vs SSL based on port
		useSTARTTLS := port == 587

		sender = email.NewSMTPSender(&email.Config{
			Enabled:     true,
			Provider:    "smtp",
			Host:        host,
			Port:        port,
			Username:    username,
			Password:    password,
			FromEmail:   fromEmail,
			FromName:    "",
			UseSTARTTLS: useSTARTTLS,
			Timeout:     time.Duration(timeoutSeconds) * time.Second,
		})
		logrus.Infof("Using SMTP email sender: %s:%d", host, port)
	}

	// Services
	storageService := storageSvc.NewService()
	manuscriptService := manuscriptSvc.NewService(mRepo, iRepo, jRepo, tRepo, storageService)
	authService := authSvc.NewService(uRepo, rRepo, rdb, sender, manuscriptService)
	termService := termSvc.NewService(tRepo)
	reviewService := reviewSvc.NewService(rvRepo, manuscriptService, storageService, sender, uRepo, rRepo)

	issueService := issueSvc.NewService(iRepo, vRepo, fRepo, storageService, manuscriptService)
	volumeService := volumeSvc.NewService(vRepo, jRepo, issueService)
	journalService := journalSvc.NewService(jRepo, fRepo, storageService, volumeService)

	// Controllers
	authController := authHttp.NewController(authService, uRepo)
	userController := userHttp.NewController(authService, uRepo)
	warmupController := warmupHttp.NewController(storageService)

	journalController := journalHttp.NewController(journalService)
	volumeController := volumeHttp.NewController(volumeService, rRepo)
	issueController := issueHttp.NewController(issueService, rRepo)
	manuscriptController := manuscriptHttp.NewController(manuscriptService, rRepo)
	reviewController := reviewHttp.NewController(reviewService)
	termController := termHttp.NewController(termService)

	// HTTP Transport + routes
	httpTransport.NewTransport().
		WithGinEngine(r).
		WithAuthController(authController).
		WithUserController(userController).
		WithWarmupController(warmupController).
		WithJournalController(journalController).
		WithVolumeController(volumeController).
		WithIssueController(issueController).
		WithManuscriptController(manuscriptController).
		WithReviewController(reviewController).
		WithTermController(termController).
		WithRoleRepository(rRepo).
		InitRoute(rdb)

	// Configure HTTP server with production-ready timeouts
	readTimeout := 15 * time.Second
	writeTimeout := 15 * time.Second
	idleTimeout := 60 * time.Second
	maxHeaderBytes := 1 << 20 // 1 MB

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", config.Env.Server.Port),
		Handler:        r.Handler(),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    idleTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
	// start http server
	go func() {
		logrus.Info(fmt.Sprintf("running at http://0.0.0.0:%s", config.Env.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			util.ContinueOrFatal(err)
		}
	}()

	wait := gracefulShutdown(ctx, config.Env.GracefulShutdownTimeout, map[string]operation{
		"database connection": func(ctx context.Context) error {
			infrastructure.StopTickerCh <- true
			return db.Close()
		},
		"redis connection": func(ctx context.Context) error {
			if rdb != nil {
				return rdb.Close()
			}
			return nil
		},
		"gin server": func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	<-wait
}
