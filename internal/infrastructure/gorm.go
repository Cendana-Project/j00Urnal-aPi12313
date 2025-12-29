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

	// Log DSN parameters (without credentials) for debugging
	logrus.WithFields(logrus.Fields{
		"has_statement_cache_capacity": strings.Contains(dsnWithParams, "statement_cache_capacity=0"),
		"has_statement_cache_mode":     strings.Contains(dsnWithParams, "statement_cache_mode=describe"),
	}).Debug("database DSN parameters configured")

	// Configure GORM with best practices for managed Postgres
	psqlDialector := postgres.Open(dsnWithParams)

	// GORM config optimized for managed Postgres to prevent prepared statement conflicts
	gormConfig := &gorm.Config{
		// CRITICAL: Disable prepared statements completely
		// This is the most important setting to prevent "prepared statement already exists" errors
		PrepareStmt:    false,
		TranslateError: true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// Disable query fields caching to reduce prepared statement usage
		DisableForeignKeyConstraintWhenMigrating: false,
	}

	db, err := gorm.Open(psqlDialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Get underlying *sql.DB for connection pool configuration
	conn, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Configure connection pool with best practices for production
	// Since we've disabled statement cache with statement_cache_capacity=0,
	// we can use more reasonable connection pool settings
	maxIdleConns := config.Env.Database.MaxIdleConns
	maxOpenConns := config.Env.Database.MaxOpenConns
	connMaxLifetime := config.Env.Database.MaxConnLifetime

	// For managed Postgres (Render/Supabase), optimize for production workload
	if config.Env.Env == constant.ProductionEnvironment {
		// Production-optimized connection pool settings
		// Max idle connections: 20-30% of max open connections
		// This balances connection reuse with resource efficiency
		if maxIdleConns > 10 {
			maxIdleConns = 10
		}
		if maxIdleConns < 5 {
			maxIdleConns = 5
		}

		// Max open connections: match managed Postgres connection limits
		// Render free tier typically allows 20-30 connections
		// For higher tiers, can be increased to 50-100
		if maxOpenConns > 25 {
			maxOpenConns = 25
		}

		// Connection lifetime: 15-30 minutes is optimal
		// Long enough for performance, short enough for leak detection
		// and to pick up connection parameter changes
		if connMaxLifetime > 30*time.Minute {
			connMaxLifetime = 30 * time.Minute
		}
		if connMaxLifetime < 15*time.Minute {
			connMaxLifetime = 15 * time.Minute
		}

		// Idle timeout: 5-10 minutes balances reuse and cleanup
		conn.SetConnMaxIdleTime(10 * time.Minute)
	} else {
		// Development: use more relaxed settings
		conn.SetConnMaxIdleTime(15 * time.Minute)
	}

	conn.SetMaxIdleConns(maxIdleConns)
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetConnMaxLifetime(connMaxLifetime)
	// Note: ConnMaxIdleTime is set conditionally above based on environment

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

	// CRITICAL: Completely disable pgx statement cache
	// The only way to truly disable statement cache in pgx is to set capacity to 0
	// statement_cache_mode=describe still uses cache, just in a different mode
	// Setting capacity to 0 prevents any caching at all
	if !strings.Contains(dsn, "statement_cache_capacity") {
		params = append(params, "statement_cache_capacity=0")
	}

	// Also set statement_cache_mode=describe as additional safeguard
	if !strings.Contains(dsn, "statement_cache_mode") {
		params = append(params, "statement_cache_mode=describe")
	}

	// Add prefer_simple_protocol=1 as fallback (works with lib/pq, ignored by pgx but harmless)
	if !strings.Contains(dsn, "prefer_simple_protocol") {
		params = append(params, "prefer_simple_protocol=1")
	}

	// For managed Postgres in production, ensure we use SSL
	if !strings.Contains(dsn, "sslmode") && config.Env.Env == constant.ProductionEnvironment {
		params = append(params, "sslmode=require")
	}

	// Disable connection pooling at driver level to force fresh connections
	// This helps prevent prepared statement conflicts when connections are reused
	if !strings.Contains(dsn, "pool_max_conns") && config.Env.Env == constant.ProductionEnvironment {
		// Let application-level connection pool handle this instead
		// Don't add pool_max_conns here as it conflicts with our connection pool settings
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
