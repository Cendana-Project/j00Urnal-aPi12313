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
	authCtrl "github.com/api-monolith-template/internal/transport/http/auth"
)

func StartServer() {
	// Infra
	gormDB := infrastructure.InitializeDBConn()
	rdb := infrastructure.NewRedisClient()
	r := infrastructure.NewGinEngine()

	// Repo
	uRepo := userRepo.NewRepository(gormDB)
	rRepo := roleRepo.NewRepository(gormDB)

	// SMTP config + defaults
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
	emCfg := &email.Config{
		Enabled:     true,
		Provider:    "smtp",
		Host:        host,
		Port:        port,
		Username:    username,
		Password:    password,
		FromEmail:   fromEmail,
		FromName:    "MedikaOne",
		Timeout:     15 * time.Second,
		UseSTARTTLS: true,
	}
	sender := email.NewSMTPSender(emCfg)

	// Service
	authService := authSvc.NewService(uRepo, rRepo, rdb, sender)

	// Transport
	authController := authCtrl.NewController(authService)
	httpTransport.
		NewTransport().
		WithGinEngine(r).
		WithAuthController(authController).
		InitRoute()

	_ = r.Run(":" + config.Env.Server.Port)
}
