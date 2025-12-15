package bootstrap

import (
	"os"
	"strconv"
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
)

func StartServer() {
	// Infra
	gormDB := infrastructure.InitializeDBConn()
	rdb := infrastructure.NewRedisClient()
	r := infrastructure.NewGinEngine()

	// Repositories
	uRepo := userRepo.NewRepository(gormDB)
	rRepo := roleRepo.NewRepository(gormDB)

	// SMTP sender config (fallback default)
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
	fromEmail := config.Env.SMTP.From
	if fromEmail == "" {
		fromEmail = "no-reply@medikaone.id"
	}

	// Configure email timeout (default 30s, override via EMAIL_TIMEOUT_SECONDS)
	timeoutSeconds := 30
	if v := os.Getenv("EMAIL_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeoutSeconds = n
		}
	}

	sender := email.NewSMTPSender(&email.Config{
		Enabled:     true,
		Provider:    "smtp",
		Host:        host,
		Port:        port,
		Username:    username,
		Password:    password,
		FromEmail:   fromEmail,
		FromName:    "", // Biarkan sender.go mem-parse nama dari FromEmail
		UseSTARTTLS: true,
		Timeout:     time.Duration(timeoutSeconds) * time.Second,
	})

	// Services
	authService := authSvc.NewService(uRepo, rRepo, rdb, sender)

	// Controllers
	authController := authHttp.NewController(authService, uRepo)
	userController := userHttp.NewController(authService, uRepo)
	warmupController := warmupHttp.NewController()

	// HTTP Transport + routes
	httpTransport.NewTransport().
		WithGinEngine(r).
		WithAuthController(authController).
		WithUserController(userController).
		WithWarmupController(warmupController).
		WithRoleRepository(rRepo).
		InitRoute()

	// Start server
	_ = r.Run(":" + config.Env.Server.Port)
}
