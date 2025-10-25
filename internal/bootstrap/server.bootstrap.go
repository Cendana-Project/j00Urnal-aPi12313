package bootstrap

import (
	"time"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/email"
	"github.com/api-monolith-template/internal/infrastructure"
	hospRepo "github.com/api-monolith-template/internal/repository/hospital"
	roleRepo "github.com/api-monolith-template/internal/repository/role"
	userRepo "github.com/api-monolith-template/internal/repository/user"
	authSvc "github.com/api-monolith-template/internal/service/auth"
	hospSvc "github.com/api-monolith-template/internal/service/hospital"
	httpTransport "github.com/api-monolith-template/internal/transport/http"
	authHttp "github.com/api-monolith-template/internal/transport/http/auth"
	hospHttp "github.com/api-monolith-template/internal/transport/http/hospital"
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
	hRepo := hospRepo.NewRepository(gormDB)

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
	fromEmail := username
	if fromEmail == "" {
		fromEmail = "no-reply@medikaone.id"
	}

	sender := email.NewSMTPSender(&email.Config{
		Enabled:     true,
		Provider:    "smtp",
		Host:        host,
		Port:        port,
		Username:    username,
		Password:    password,
		FromEmail:   fromEmail,
		FromName:    "MedikaOne",
		UseSTARTTLS: true,
		Timeout:     15 * time.Second,
	})

	// Services
	authService := authSvc.NewService(uRepo, rRepo, rdb, sender, hRepo)
	hospitalService := hospSvc.NewService(uRepo, rRepo, hRepo, rdb)

	// Controllers
	authController := authHttp.NewController(authService)
	userController := userHttp.NewController(authService)
	hospitalController := hospHttp.NewController(hospitalService)
	warmupController := warmupHttp.NewController()

	// HTTP Transport + routes
	httpTransport.NewTransport().
		WithGinEngine(r).
		WithAuthController(authController).
		WithUserController(userController).
		WithHospitalController(hospitalController).
		WithWarmupController(warmupController).
		WithRoleRepository(rRepo).
		WithHospitalRepository(hRepo).
		InitRoute()

	// Start server
	_ = r.Run(":" + config.Env.Server.Port)
}
