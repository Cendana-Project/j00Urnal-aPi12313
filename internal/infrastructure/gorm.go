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

	// Use lib/pq instead of pgx for PgBouncer compatibility
	// lib/pq doesn't have aggressive statement caching like pgx
	_ "github.com/lib/pq" // PostgreSQL driver
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var (
	DB           *gorm.DB
	StopTickerCh chan bool
)

// InitializeDBConn initializes database connection with production-ready settings
func InitializeDBConn() *gorm.DB {
	conn, err := openDBConn(config.Env.Database.DSN)
	if err != nil {
		fields := logrus.Fields{"error": err}
		errStr := err.Error()
		if strings.Contains(errStr, "no such host") || strings.Contains(strings.ToUpper(errStr), "NXDOMAIN") {
			fields["hint"] = "Database hostname could not be resolved — check DATABASE_DSN matches Supabase Dashboard → Project Settings → Database (use current pooler host). Verify network/VPN/firewall."
		}
		logrus.WithFields(fields).Fatal("failed to connect to database")
	}

	DB = conn
	StopTickerCh = make(chan bool)

	// Start background health check
	go checkConnection(time.NewTicker(config.Env.Database.PingInterval))

	// Set log level
	setLogLevel(DB)

	// AutoMigrate only in development and NOT using PgBouncer
	// In production or pooled environments, use explicit migrations via goose
	isPooled := strings.Contains(config.Env.Database.DSN, "pgbouncer=true") || strings.Contains(config.Env.Database.DSN, "6543")

	if config.Env.Env != constant.ProductionEnvironment && !isPooled {
		logrus.Info("running GORM AutoMigrate...")
		if err := autoMigrate(DB); err != nil {
			logrus.WithError(err).Warn("GORM AutoMigrate failed (this is common with PgBouncer). Skipping auto-migrate...")
		}
	} else {
		if isPooled {
			logrus.Info("skipping GORM AutoMigrate due to pooled connection (PgBouncer)")
		} else {
			logrus.Info("skipping GORM AutoMigrate in production (use explicit migrations)")
		}
	}

	// Register health check
	MapHealthCheck["database"] = func(ctx context.Context) error {
		if DB == nil {
			return errors.New("database connection is nil")
		}
		sqlDB, err := DB.WithContext(ctx).DB()
		if err != nil {
			return err
		}
		return sqlDB.Ping()
	}

	logrus.Info("database connection established successfully")
	return DB
}

// autoMigrate runs GORM AutoMigrate for development environments
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.User{},
		&entity.Role{},
		&entity.Permission{},
		&entity.UserRole{},
		&entity.RolePermission{},
	)
}

// setLogLevel configures GORM logger based on application log level
func setLogLevel(db *gorm.DB) {
	switch config.Env.LogLevel {
	case "error":
		db.Logger = db.Logger.LogMode(gormLogger.Error)
	case "warn":
		db.Logger = db.Logger.LogMode(gormLogger.Warn)
	default:
		db.Logger = db.Logger.LogMode(gormLogger.Info)
	}
}

// RunMigrations runs database migrations using goose
// Safe to call multiple times - goose will skip already applied migrations
func RunMigrations() error {
	migrationDir := "migration/db"

	// Use direct connection for migrations (bypasses PgBouncer if configured)
	dsn := config.Env.Database.DirectDSN
	if dsn == "" {
		// Fallback to regular DSN if direct DSN not configured
		dsn = config.Env.Database.DSN
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open migration connection: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	logrus.Info("running database migrations...")
	if err := goose.Up(db, migrationDir, goose.WithAllowMissing()); err != nil {
		logrus.WithError(err).Error("migration failed")
		return err
	}

	logrus.Info("database migrations completed successfully")
	return nil
}

// checkConnection periodically checks database connection health
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

			// Ping to verify connection
			if err := sqlDB.Ping(); err != nil {
				logrus.WithError(err).Error("database ping failed, attempting reconnect...")
				reconnectDBConn()
				continue
			}

			// Log connection pool stats periodically
			if checkCount%10 == 0 {
				logPoolStats(sqlDB)
			}
		}
	}
}

// logPoolStats logs database connection pool statistics
func logPoolStats(db *sql.DB) {
	stats := db.Stats()
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

// GetDB returns a new isolated session from the current database connection.
// Each call creates a fresh GORM session with NewDB: true, ensuring
// concurrent goroutines never share internal GORM statement state
// (preloads, clauses, etc.), which prevents race conditions under load.
// Use this instead of storing the DB pointer in long-lived structs
// to ensure you always have the latest connection after a reconnect.
func GetDB() *gorm.DB {
	if DB == nil {
		return nil
	}
	return DB.Session(&gorm.Session{NewDB: true})
}

// reconnectDBConn attempts to reconnect to database with exponential backoff
func reconnectDBConn() {
	logrus.Warn("attempting to reconnect to database...")

	// Close existing connection to free up resources
	if DB != nil {
		if sqlDB, err := DB.DB(); err == nil {
			logrus.Info("closing stale database connection before reconnecting...")
			sqlDB.Close()
		}
	}

	b := backoff.Backoff{
		Factor: config.Env.Database.ReconnectFactor,
		Jitter: true,
		Min:    config.Env.Database.MinJitter,
		Max:    config.Env.Database.MaxJitter,
	}

	maxRetries := config.Env.Database.MaxRetry
	var lastErr error

	for b.Attempt() < float64(maxRetries) {
		conn, err := openDBConn(config.Env.Database.DSN)
		if err != nil {
			lastErr = err
			logrus.WithError(err).WithField("attempt", int(b.Attempt())+1).Error("reconnection failed")
			time.Sleep(b.Duration())
			continue
		}

		// Verify connection works
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
	logrus.WithError(lastErr).Fatal("maximum retry attempts reached, failed to reconnect")
}

// openDBConn opens a new GORM database connection with optimized settings
func openDBConn(dsn string) (*gorm.DB, error) {
	// Enhance DSN with necessary parameters
	dsnWithParams := buildDSN(dsn)

	// CRITICAL: Use postgres.New with PreferSimpleProtocol to force lib/pq behavior
	// This completely avoids pgx's aggressive statement caching
	dialector := postgres.New(postgres.Config{
		DriverName:           "postgres", // Use lib/pq driver (imported with _)
		DSN:                  dsnWithParams,
		PreferSimpleProtocol: true, // Force simple protocol - NO prepared statements
	})

	// Configure GORM
	db, err := gorm.Open(dialector, &gorm.Config{
		PrepareStmt:    false, // Double safety: disable at GORM level too
		TranslateError: true,
		NowFunc:        func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Get underlying *sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Configure connection pool
	configureConnectionPool(sqlDB)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.Debug("database connection pool configured successfully")
	return db, nil
}

// configureConnectionPool sets up connection pool with production-ready settings
func configureConnectionPool(db *sql.DB) {
	cfg := config.Env.Database

	// Default sensible values for production
	maxIdleConns := 10
	maxOpenConns := 25
	connMaxLifetime := 30 * time.Minute
	connMaxIdleTime := 10 * time.Minute

	// Use config values if provided
	if cfg.MaxIdleConns > 0 {
		maxIdleConns = cfg.MaxIdleConns
	}
	if cfg.MaxOpenConns > 0 {
		maxOpenConns = cfg.MaxOpenConns
	}
	if cfg.MaxConnLifetime > 0 {
		connMaxLifetime = cfg.MaxConnLifetime
	}

	// For PgBouncer/Supabase, adjust settings
	if config.Env.Env == constant.ProductionEnvironment {
		// PgBouncer recommends lower connection counts
		// Free tier typically allows 20-30 connections
		if maxOpenConns > 20 {
			maxOpenConns = 20
		}
		if maxIdleConns > 5 {
			maxIdleConns = 5
		}

		logrus.WithFields(logrus.Fields{
			"max_idle_conns":     maxIdleConns,
			"max_open_conns":     maxOpenConns,
			"conn_max_lifetime":  connMaxLifetime,
			"conn_max_idle_time": connMaxIdleTime,
		}).Info("production connection pool configured for PgBouncer")
	}

	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)
}

// buildDSN adds necessary parameters to DSN for managed Postgres compatibility
func buildDSN(dsn string) string {
	// Don't modify if it's already properly configured
	if strings.Contains(dsn, "pgbouncer=true") {
		// Supabase connection pooler - already optimized
		return dsn
	}

	params := []string{}
	hasParams := strings.Contains(dsn, "?")

	// For managed Postgres (Render/others), enforce SSL in production
	if !strings.Contains(dsn, "sslmode") && config.Env.Env == constant.ProductionEnvironment {
		params = append(params, "sslmode=require")
	}

	// No additional parameters needed - PrepareStmt: false handles everything

	if len(params) == 0 {
		return dsn
	}

	separator := "?"
	if hasParams {
		separator = "&"
	}

	return dsn + separator + strings.Join(params, "&")
}
