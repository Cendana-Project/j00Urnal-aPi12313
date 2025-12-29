package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/jpillora/backoff"
	"github.com/pressly/goose/v3"
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
	// NOTE:
	// - In production we rely on explicit SQL migrations instead of AutoMigrate.
	// - Some managed Postgres providers (like Render) can return fatal errors such as
	//   "prepared statement already exists" or "relation already exists" when
	//   AutoMigrate issues DDL on an already‑managed schema.
	// - To avoid bringing the whole service down on startup, we only run AutoMigrate
	//   outside production.
	if config.Env.Env != constant.ProductionEnvironment {
		if err := DB.AutoMigrate(
			&entity.User{},
			&entity.Role{},
			&entity.Permission{},
			&entity.UserRole{},
			&entity.RolePermission{},
		); err != nil {
			logrus.Fatal("failed to auto-migrate: ", err)
		}
	} else {
		logrus.Info("skipping GORM AutoMigrate in production environment")
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

// RunMigrations runs database migrations using goose.
// This function is safe to call multiple times - goose will skip migrations
// that have already been applied. It will not fail if migrations already exist.
func RunMigrations() error {
	migrationDir := "migration/db"

	db, err := sql.Open("postgres", config.Env.Database.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	logrus.Info("running database migrations...")
	// goose.Up with WithAllowMissing() will:
	// - Apply any pending migrations
	// - Skip migrations that have already been applied (no error)
	// - Continue even if some migration files are missing
	if err := goose.Up(db, migrationDir, goose.WithAllowMissing()); err != nil {
		logrus.WithError(err).Error("migration failed")
		return err
	}

	logrus.Info("database migrations completed successfully")
	return nil
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
	// Add prefer_simple_protocol=1 to DSN to disable prepared statements at driver level
	// This prevents "prepared statement already exists" errors on managed Postgres providers
	// like Render/Supabase where connections are pooled and reused
	dsnWithParams := dsn
	if !strings.Contains(dsn, "prefer_simple_protocol") {
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		dsnWithParams = dsn + separator + "prefer_simple_protocol=1"
	}

	psqlDialector := postgres.Open(dsnWithParams)
	db, err := gorm.Open(psqlDialector, &gorm.Config{
		// NOTE:
		// - We deliberately disable GORM's global prepared statement cache here.
		// - On some managed Postgres providers (including Render/Supabase), enabling
		//   PrepareStmt together with connection pooling can trigger
		//   "prepared statement already exists" (SQLSTATE 42P05) errors.
		// - We also add prefer_simple_protocol=1 to DSN to disable prepared statements
		//   at the PostgreSQL driver level, ensuring no prepared statements are created.
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
