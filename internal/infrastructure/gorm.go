package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	checkCount := 0
	for {
		select {
		case <-StopTickerCh:
			ticker.Stop()
			return
		case <-ticker.C:
			checkCount++
			if DB == nil {
				logrus.Error("database connection is nil, attempting reconnect...")
				reconnectDBConn()
				continue
			}

			sqlDB, err := DB.DB()
			if err != nil {
				logrus.WithError(err).Error("failed to get database connection, attempting reconnect...")
				reconnectDBConn()
				continue
			}

			// Ping to verify connection is still alive
			if err := sqlDB.Ping(); err != nil {
				logrus.WithError(err).Error("database ping failed, attempting reconnect...")
				reconnectDBConn()
				continue
			}

			// Log connection pool stats periodically (every 10th check)
			if checkCount%10 == 0 {
				stats := sqlDB.Stats()
				logrus.WithFields(logrus.Fields{
					"open_connections":     stats.OpenConnections,
					"in_use":               stats.InUse,
					"idle":                 stats.Idle,
					"wait_count":           stats.WaitCount,
					"wait_duration":        stats.WaitDuration,
					"max_idle_closed":      stats.MaxIdleClosed,
					"max_idle_time_closed": stats.MaxIdleTimeClosed,
					"max_lifetime_closed":  stats.MaxLifetimeClosed,
				}).Debug("database connection pool stats")
			}
		}
	}
}

func reconnectDBConn() {
	logrus.Warn("attempting to reconnect to database...")

	// Close existing connection if any
	if DB != nil {
		if sqlDB, err := DB.DB(); err == nil {
			sqlDB.Close()
		}
	}

	b := backoff.Backoff{
		Factor: config.Env.Database.ReconnectFactor,
		Jitter: true,
		Min:    config.Env.Database.MinJitter,
		Max:    config.Env.Database.MaxJitter,
	}

	dbRetryAttempts := config.Env.Database.MaxRetry

	var lastErr error
	for b.Attempt() < float64(dbRetryAttempts) {
		conn, err := openDBConn(config.Env.Database.DSN)
		if err != nil {
			lastErr = err
			logrus.WithError(err).WithField("attempt", int(b.Attempt())+1).Error("failed to reconnect to database")
			time.Sleep(b.Duration())
			continue
		}

		// Verify connection is working
		if sqlDB, err := conn.DB(); err == nil {
			if err := sqlDB.Ping(); err != nil {
				lastErr = err
				logrus.WithError(err).Error("reconnected but ping failed")
				sqlDB.Close()
				time.Sleep(b.Duration())
				continue
			}
		}

		// Successfully reconnected
		DB = conn
		logrus.Info("successfully reconnected to database")
		b.Reset()
		return
	}

	// All retries exhausted
	logrus.WithError(lastErr).Fatal("maximum retry attempts reached, failed to reconnect to database")
}

func openDBConn(dsn string) (*gorm.DB, error) {
	// Build DSN with parameters optimized for managed Postgres providers
	// This prevents "prepared statement already exists" errors on Render/Supabase
	dsnWithParams := buildDSNWithParams(dsn)

	// Configure GORM with best practices for managed Postgres
	psqlDialector := postgres.Open(dsnWithParams)
	db, err := gorm.Open(psqlDialector, &gorm.Config{
		// Disable prepared statements to avoid conflicts in managed Postgres
		// where connections are pooled and reused across instances
		PrepareStmt:    false,
		TranslateError: true,
		// Disable NamingStrategy to avoid unnecessary schema introspection
		// which can cause prepared statement conflicts
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Get underlying *sql.DB for connection pool configuration
	conn, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Configure connection pool with best practices for managed Postgres
	// Conservative settings to avoid connection exhaustion and prepared statement conflicts
	maxIdleConns := config.Env.Database.MaxIdleConns
	maxOpenConns := config.Env.Database.MaxOpenConns
	connMaxLifetime := config.Env.Database.MaxConnLifetime

	// For managed Postgres (Render/Supabase), use more conservative settings
	if config.Env.Env == constant.ProductionEnvironment {
		// Limit idle connections to prevent prepared statement conflicts
		if maxIdleConns > 5 {
			maxIdleConns = 5
		}
		// Limit open connections to match managed Postgres connection limits
		if maxOpenConns > 20 {
			maxOpenConns = 20
		}
		// Shorter connection lifetime to force reconnection and clear prepared statements
		if connMaxLifetime > 30*time.Minute {
			connMaxLifetime = 30 * time.Minute
		}
	}

	conn.SetMaxIdleConns(maxIdleConns)
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetConnMaxLifetime(connMaxLifetime)
	// Set connection timeout to prevent hanging connections
	conn.SetConnMaxIdleTime(5 * time.Minute)

	// Test the connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"max_idle_conns":    maxIdleConns,
		"max_open_conns":    maxOpenConns,
		"conn_max_lifetime": connMaxLifetime,
	}).Debug("database connection pool configured")

	return db, nil
}

// buildDSNWithParams adds necessary parameters to DSN for managed Postgres compatibility
// This function ensures the DSN is optimized for managed Postgres providers like Render/Supabase
// to prevent "prepared statement already exists" errors
func buildDSNWithParams(dsn string) string {
	// Check if DSN already has query parameters
	hasParams := strings.Contains(dsn, "?")
	params := []string{}

	// Add prefer_simple_protocol=1 to disable prepared statements at driver level
	// Note: This parameter works with lib/pq driver. For pgx driver (used by GORM),
	// we rely on PrepareStmt: false in GORM config, but adding this doesn't hurt.
	if !strings.Contains(dsn, "prefer_simple_protocol") {
		params = append(params, "prefer_simple_protocol=1")
	}

	// For managed Postgres in production, ensure we use SSL
	if !strings.Contains(dsn, "sslmode") && config.Env.Env == constant.ProductionEnvironment {
		params = append(params, "sslmode=require")
	}

	// Add statement_cache_mode=describe to avoid prepared statement caching issues
	// This is a pgx-specific parameter that helps with managed Postgres
	if !strings.Contains(dsn, "statement_cache_mode") {
		params = append(params, "statement_cache_mode=describe")
	}

	if len(params) == 0 {
		return dsn
	}

	// Append parameters to DSN
	separator := "?"
	if hasParams {
		separator = "&"
	}

	return dsn + separator + strings.Join(params, "&")
}
