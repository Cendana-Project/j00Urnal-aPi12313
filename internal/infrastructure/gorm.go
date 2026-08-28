package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// healthCheckTimeout bounds every individual health-check probe (DB ping, Supabase REST call)
// so a hanging dependency can never wedge /_internal/healthz — Render polls that endpoint to
// decide whether this service is alive.
const healthCheckTimeout = 5 * time.Second

// supabaseHTTPClient is used only for the supabase_rest health probe; it must have its own
// bounded timeout since http.DefaultClient has none and would otherwise let a slow/stuck
// Supabase gateway hang the request forever.
var supabaseHTTPClient = &http.Client{Timeout: healthCheckTimeout}

// InitializeDBConn initializes the database connection with production-ready settings.
// A failed initial connection does NOT crash the process, and — critically — does NOT leave DB
// nil either: openDBConn hands back a valid *gorm.DB even when the startup ping failed, because
// database/sql's pool re-dials lazily on every future query. That means DB-backed handlers never
// see a nil connection and never panic; they just get a normal query error until Postgres/Supabase
// becomes reachable, at which point requests (and the periodic checkConnection ticker below) start
// succeeding again on their own — no manual restart needed.
func InitializeDBConn() *gorm.DB {
	StopTickerCh = make(chan bool)
	registerDBHealthChecks()

	conn, err := openDBConn(config.Env.Database.DSN)
	if conn == nil {
		// Only nil when gorm.Open/db.DB() itself failed (bad DSN, driver misconfigured) — not
		// recoverable by retrying the same DSN, but we still must not crash the process.
		fields := logrus.Fields{"error": err}
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		if strings.Contains(errStr, "no such host") || strings.Contains(strings.ToUpper(errStr), "NXDOMAIN") {
			fields["hint"] = "Database hostname could not be resolved — check DATABASE_DSN matches Supabase Dashboard → Project Settings → Database (use current pooler host). Verify network/VPN/firewall."
		}
		logrus.WithFields(fields).Error("failed to initialize database driver; DB-backed endpoints will error until DATABASE_DSN is fixed and the process is restarted")
		go checkConnection(time.NewTicker(config.Env.Database.PingInterval))
		return nil
	}

	DB = conn
	if err != nil {
		logrus.WithError(err).Error("database not reachable at startup; server will start anyway — queries will retry automatically as connectivity recovers")
	} else {
		logrus.Info("database connection established successfully")
	}
	finishDBConnSetup(DB)
	go checkConnection(time.NewTicker(config.Env.Database.PingInterval))
	return DB
}

// finishDBConnSetup runs the one-time setup that only makes sense once a connection is live:
// log level and (dev-only) AutoMigrate. Called on both first connect and successful reconnect.
func finishDBConnSetup(db *gorm.DB) {
	setLogLevel(db)

	// AutoMigrate only in development and NOT using PgBouncer
	// In production or pooled environments, use explicit migrations via goose
	isPooled := strings.Contains(config.Env.Database.DSN, "pgbouncer=true") || strings.Contains(config.Env.Database.DSN, "6543")

	if config.Env.Env != constant.ProductionEnvironment && !isPooled {
		logrus.Info("running GORM AutoMigrate...")
		if err := autoMigrate(db); err != nil {
			logrus.WithError(err).Warn("GORM AutoMigrate failed (this is common with PgBouncer). Skipping auto-migrate...")
		}
	} else {
		if isPooled {
			logrus.Info("skipping GORM AutoMigrate due to pooled connection (PgBouncer)")
		} else {
			logrus.Info("skipping GORM AutoMigrate in production (use explicit migrations)")
		}
	}
}

// registerDBHealthChecks wires the "database" and "supabase_rest" checks used by
// /_internal/healthz. Registered unconditionally, even before a connection exists, so the
// endpoint always reports accurate status instead of missing entries during an outage.
func registerDBHealthChecks() {
	MapHealthCheck["database"] = func(ctx context.Context) error {
		if DB == nil {
			return errors.New("database connection is nil")
		}
		sqlDB, err := DB.WithContext(ctx).DB()
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
		defer cancel()
		return sqlDB.PingContext(ctx)
	}

	// Second, independent check that hits Supabase's PostgREST gateway directly. A raw
	// Postgres ping through the pooler does NOT reset Supabase's free-tier inactivity timer -
	// only requests through its own API surface (REST/Auth/Storage/Realtime) do. This keeps
	// the project from auto-pausing even if the pooler connection above is not counted.
	MapHealthCheck["supabase_rest"] = pingSupabaseRest
}

// pingSupabaseRest issues a lightweight, time-bounded request through Supabase's PostgREST
// API gateway (not a direct Postgres connection), since Supabase's inactivity tracker only
// counts requests going through its own API surface.
func pingSupabaseRest(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	url := fmt.Sprintf("%s/rest/v1/", config.Env.Supabase.URL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", config.Env.Supabase.ServiceRoleKey)
	req.Header.Set("Authorization", "Bearer "+config.Env.Supabase.ServiceRoleKey)

	resp, err := supabaseHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase REST ping failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	return nil
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

			// Ping (time-bounded) to verify connection
			pingCtx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
			pingErr := sqlDB.PingContext(pingCtx)
			cancel()
			if pingErr != nil {
				logrus.WithError(pingErr).Error("database ping failed, attempting reconnect...")
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

// reconnectDBConn attempts to reconnect to the database with exponential backoff. It never
// crashes the process: if every attempt in this pass fails, it logs and returns — the periodic
// checkConnection ticker calls it again on the next tick, so retries continue indefinitely until
// Postgres/Supabase comes back, and the server self-heals without a manual restart.
//
// Deliberately does NOT null out DB while retrying: the previous connection (even if currently
// unreachable) is left in place until a confirmed-good replacement is ready to swap in, so
// GetDB() never hands callers a nil *gorm.DB mid-reconnect — that used to cause every DB-backed
// handler to panic instead of returning a normal query error.
func reconnectDBConn() {
	logrus.Warn("attempting to reconnect to database...")
	old := DB

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
			// openDBConn may still hand back a usable-but-unreachable *gorm.DB alongside the
			// error; close it before discarding so retries don't leak pool goroutines.
			closeGormDB(conn)
			time.Sleep(b.Duration())
			continue
		}

		// Successfully reconnected — swap in the new connection, then close the old one.
		DB = conn
		finishDBConnSetup(DB)
		logrus.Info("successfully reconnected to database")
		closeGormDB(old)
		b.Reset()
		return
	}

	// Give up for this pass only — not fatal. The periodic checkConnection ticker will retry
	// again on its next tick. DB still points at the previous connection, so callers keep
	// getting normal query errors instead of a nil-DB panic in the meantime.
	logrus.WithError(lastErr).Error("reconnect attempts exhausted for this pass; will retry again on the next health check tick")
}

// closeGormDB closes the underlying *sql.DB of db if non-nil, ignoring errors (best-effort
// cleanup when discarding a stale or failed connection object).
func closeGormDB(db *gorm.DB) {
	if db == nil {
		return
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}
}

// openDBConn opens a new GORM database connection with optimized settings.
//
// Builds the *sql.DB via plain database/sql.Open first, then wraps it into GORM via
// postgres.Config{Conn: sqlDB} instead of handing GORM the DSN directly. database/sql.Open is
// documented to never dial the network — it only validates arguments — so this guarantees the
// object-construction step itself can't fail for connectivity reasons (e.g. DNS not resolving
// because Supabase is paused). Handing GORM the DSN directly used to skip that guarantee: its
// postgres dialector's own Initialize() dials eagerly, so a DNS failure made gorm.Open itself
// return nil, which is exactly the case that used to leave the global DB nil and made every
// DB-backed handler panic instead of returning a normal query error.
func openDBConn(dsn string) (*gorm.DB, error) {
	// Enhance DSN with necessary parameters
	dsnWithParams := buildDSN(dsn)

	// lib/pq is registered under driver name "postgres" (blank-imported in this package).
	sqlDB, err := sql.Open("postgres", dsnWithParams)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	configureConnectionPool(sqlDB)

	// CRITICAL: Use postgres.New with PreferSimpleProtocol to force lib/pq behavior
	// This completely avoids pgx's aggressive statement caching
	dialector := postgres.New(postgres.Config{
		Conn:                 sqlDB, // pre-opened pool — see comment above for why
		PreferSimpleProtocol: true,  // Force simple protocol - NO prepared statements
	})

	// Configure GORM. DisableAutomaticPing: gorm.Open() otherwise pings internally with no
	// timeout/context of its own and, on failure, still returns a fully-valid *gorm.DB — but
	// relying on that undocumented detail is fragile. We run our own bounded PingContext right
	// after instead, so connectivity failures here can only come from dialector.Initialize()
	// itself, which (with a pre-opened Conn) never touches the network.
	db, err := gorm.Open(dialector, &gorm.Config{
		PrepareStmt:          false, // Double safety: disable at GORM level too
		TranslateError:       true,
		NowFunc:              func() time.Time { return time.Now().UTC() },
		DisableAutomaticPing: true,
	})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to initialize gorm: %w", err)
	}

	// Test connection (time-bounded — a network blackhole must not hang boot/reconnect forever).
	// On failure we deliberately do NOT close/discard db: database/sql's pool re-dials lazily on
	// every future query, so returning it anyway means callers always get a valid, usable
	// *gorm.DB. This matters because a nil DB previously caused every DB-backed handler to panic
	// (nil pointer dereference) instead of failing with a normal, cleanly-propagated query error.
	pingCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return db, fmt.Errorf("failed to ping database: %w", err)
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
		return ensureConnectTimeout(dsn)
	}

	params := []string{}
	hasParams := strings.Contains(dsn, "?")

	// For managed Postgres (Render/others), enforce SSL in production
	if !strings.Contains(dsn, "sslmode") && config.Env.Env == constant.ProductionEnvironment {
		params = append(params, "sslmode=require")
	}

	// No additional parameters needed - PrepareStmt: false handles everything

	if len(params) == 0 {
		return ensureConnectTimeout(dsn)
	}

	separator := "?"
	if hasParams {
		separator = "&"
	}

	return ensureConnectTimeout(dsn + separator + strings.Join(params, "&"))
}

// ensureConnectTimeout guarantees a libpq connect_timeout is present so a network blackhole
// (packets silently dropped, e.g. a paused/misrouted host) can't hang the TCP+auth phase of a
// connection attempt indefinitely — capped here, PingContext bounds the phase after that.
func ensureConnectTimeout(dsn string) string {
	if strings.Contains(dsn, "connect_timeout") {
		return dsn
	}
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + "connect_timeout=10"
}
