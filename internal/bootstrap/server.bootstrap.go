package bootstrap

import (
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
)

func StartServer() {
	gormDB := infrastructure.InitializeDBConn()
	rdb := infrastructure.NewRedisClient()
	r := infrastructure.NewGinEngine()

	// repos
	uRepo := userRepo.NewRepository(gormDB)
	rRepo := roleRepo.NewRepository(gormDB)

	// SMTP sender
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

	// service
	authService := authSvc.NewService(uRepo, rRepo, rdb, sender)

	// controllers
	authController := authHttp.NewController(authService)
	userController := userHttp.NewController(authService)

	// transport
	httpTransport.NewTransport().
		WithGinEngine(r).
		WithAuthController(authController).
		WithUserController(userController).
		InitRoute()

	_ = r.Run(":" + config.Env.Server.Port)
}
