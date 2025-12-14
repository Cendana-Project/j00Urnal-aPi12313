package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/email"
	"github.com/api-monolith-template/internal/infrastructure"
	roleRepo "github.com/api-monolith-template/internal/repository/role"
	userRepo "github.com/api-monolith-template/internal/repository/user"
	authSvc "github.com/api-monolith-template/internal/service/auth"
	httpTransport "github.com/api-monolith-template/internal/transport/http"
	authCtrl "github.com/api-monolith-template/internal/transport/http/auth"
	"github.com/api-monolith-template/internal/util"
	"github.com/sirupsen/logrus"
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

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.Env.Server.Port),
		Handler: r.Handler(),
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
