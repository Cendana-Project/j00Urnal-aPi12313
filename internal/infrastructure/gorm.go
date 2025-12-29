package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/jpillora/backoff"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var (
	DB           *gorm.DB
	StopTickerCh chan bool
)

func InitializeDBConn() *gorm.DB {
	conn, err := openDBConn(config.Env.Database.DSN)
	if err != nil {
		logrus.WithField("databaseDSN", config.Env.Database.DSN).Fatal("failed to connect  database: ", err)
	}

	DB = conn
	StopTickerCh = make(chan bool)

	go checkConnection(time.NewTicker(config.Env.Database.PingInterval))

	switch config.Env.LogLevel {
	case "error":
		DB.Logger = DB.Logger.LogMode(gormLogger.Error)
	case "warn":
		DB.Logger = DB.Logger.LogMode(gormLogger.Warn)
	default:
		DB.Logger = DB.Logger.LogMode(gormLogger.Info)
	}

	// AutoMigrate to sync schema
	if err := DB.AutoMigrate(
		&entity.User{},
		&entity.Role{},
		&entity.Permission{},
		&entity.UserRole{},
		&entity.RolePermission{},
	); err != nil {
		logrus.Fatal("failed to auto-migrate: ", err)
	}

	MapHealthCheck["database"] = func(ctx context.Context) error {
		if DB == nil {
			return errors.New("disconnect")
		}

		sqlDB, err := DB.WithContext(ctx).DB()
		if err != nil {
			return err
		}

		return sqlDB.Ping()
	}

	logrus.Info("connection to database Server success...")
	return DB
}

func checkConnection(ticker *time.Ticker) {
	for {
		select {
		case <-StopTickerCh:
			ticker.Stop()
			return
		case <-ticker.C:
			if _, err := DB.DB(); err != nil {
				reconnectDBConn()
			}
		}
	}
}

func reconnectDBConn() {
	b := backoff.Backoff{
		Factor: config.Env.Database.ReconnectFactor,
		Jitter: true,
		Min:    config.Env.Database.MinJitter,
		Max:    config.Env.Database.MaxJitter,
	}

	dbRetryAttempts := config.Env.Database.MaxRetry

	for b.Attempt() < float64(dbRetryAttempts) {
		conn, err := openDBConn(config.Env.Database.DSN)
		if err != nil {
			logrus.WithField("databaseDSN", config.Env.Database.DSN).Error("failed to connect database: ", err)
		}

		if conn != nil {
			DB = conn
			break
		}
		time.Sleep(b.Duration())
	}

	if b.Attempt() >= float64(dbRetryAttempts) {
		logrus.Fatal("maximum retry to connect database")
	}
	b.Reset()
}

func openDBConn(dsn string) (*gorm.DB, error) {
	psqlDialector := postgres.Open(dsn)
	db, err := gorm.Open(psqlDialector, &gorm.Config{
		// NOTE:
		// - We deliberately disable GORM's global prepared statement cache here.
		// - On some managed Postgres providers (including Render), enabling
		//   PrepareStmt together with AutoMigrate can trigger
		//   "prepared statement already exists" (SQLSTATE 42P05) errors at startup.
		// - Disabling it avoids those startup failures with negligible impact
		//   for this API workload.
		PrepareStmt:    false,
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}

	conn, err := db.DB()
	if err != nil {
		logrus.Fatal(err)
	}
	conn.SetMaxIdleConns(config.Env.Database.MaxIdleConns)
	conn.SetMaxOpenConns(config.Env.Database.MaxOpenConns)
	conn.SetConnMaxLifetime(config.Env.Database.MaxConnLifetime)

	return db, nil
}
