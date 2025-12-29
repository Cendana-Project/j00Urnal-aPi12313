package bootstrap

import (
	"context"
	"fmt"
	"net/http"
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
	// Use configured from email or fallback to default
	fromEmail := config.Env.SMTP.FromEmail
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
