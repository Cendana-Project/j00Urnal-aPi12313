package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	ServiceName    = ""
	ServiceVersion = ""
)

var (
	Env *EnvConfig
)

type EnvConfig struct {
	Env                     string        `mapstructure:"env"`
	LogLevel                string        `mapstructure:"log_level"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
	Token                   Token         `mapstructure:"token"`
	Server                  Server        `mapstructure:"server"`
	Database                Database      `mapstructure:"database"`
	Redis                   Redis         `mapstructure:"redis"`
	SMTP                    SMTP          `mapstructure:"smtp"`
	Supabase                Supabase      `mapstructure:"supabase"`
}

type Supabase struct {
	URL            string `mapstructure:"url"`
	ServiceRoleKey string `mapstructure:"service_role_key"`
	AnonRoleKey    string `mapstructure:"anon_role_key"`
	Bucket         string `mapstructure:"bucket"`
}

type Redis struct {
	IsCacheDisable       bool          `mapstructure:"is_cache_disable"`
	CacheDSN             string        `mapstructure:"cache_dsn"`
	DefaultCacheDuration time.Duration `mapstructure:"default_cache_duration"`
	MaxRetry             int           `mapstructure:"max_retry"`
	MaxIdleConns         int           `mapstructure:"max_idle_conns"`
	MaxActiveConns       int           `mapstructure:"max_active_conns"`
	MaxConnLifetime      time.Duration `mapstructure:"max_conn_lifetime"`
}

type Token struct {
	PasswordSalt             string        `mapstructure:"password_salt"`
	AccessTokenSecret        string        `mapstructure:"access_token_secret"`
	AccessTokenDuration      time.Duration `mapstructure:"access_token_duration"`
	RefreshTokenDuration     time.Duration `mapstructure:"refresh_token_duration"`
	RefreshTokenSecret       string        `mapstructure:"refresh_token_secret"`
	ForgotPasswordDuration   time.Duration `mapstructure:"forgot_password_duration" env:"FORGOT_PASSWORD_DURATION" envDefault:"1h"`
	ForgotPasswordRateLimit  int           `mapstructure:"forgot_password_rate_limit" env:"FORGOT_PASSWORD_RATE_LIMIT" envDefault:"3"`
	ForgotPasswordRateWindow time.Duration `mapstructure:"forgot_password_rate_window" env:"FORGOT_PASSWORD_RATE_WINDOW" envDefault:"1h"`
}

type Server struct {
	Port        string `mapstructure:"port"`
	BaseURL     string `mapstructure:"base_url"`
	FrontendURL string `mapstructure:"frontend_url"`
}

type Database struct {
	DSN             string        `mapstructure:"dsn"`
	DirectDSN       string        `mapstructure:"direct_dsn"` // Direct connection for migrations (bypasses PgBouncer)
	PingInterval    time.Duration `mapstructure:"ping_interval"`
	ReconnectFactor float64       `mapstructure:"reconnect_factor"`
	MinJitter       time.Duration `mapstructure:"min_jitter"`
	MaxJitter       time.Duration `mapstructure:"max_jitter"`
	MaxRetry        int           `mapstructure:"max_retry"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
}

type SMTP struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	FromEmail string `mapstructure:"from_email"`
}

// LoadConfig loads configuration from .env file and environment variables
func LoadConfig() error {
	// Load .env file into environment variables (if it exists)
	// This allows environment variables to override .env file values
	if err := godotenv.Load(); err != nil {
		// File .env is optional, so we only log if it's not a "file not found" error
		if !strings.Contains(err.Error(), "no such file") {
			logrus.Warnf("failed to load .env file: %v", err)
		}
	}

	// Set default values
	setDefaults()

	// Configure viper to read from environment variables
	viper.SetEnvPrefix("") // No prefix, read variables as-is
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	// Explicitly bind all environment variables
	bindEnvVariables()

	// Unmarshal into config struct
	if err := viper.Unmarshal(&Env); err != nil {
		logrus.Fatal("failed to unmarshal config: ", err)
		return err
	}

	// Use PORT env var (set by Render/Heroku) if SERVER_PORT is not explicitly set
	// Check if SERVER_PORT was actually set in environment, if not, use PORT
	if os.Getenv("SERVER_PORT") == "" {
		if port := os.Getenv("PORT"); port != "" {
			Env.Server.Port = port
		}
	}

	// Validate configuration
	if err := Env.Validate(); err != nil {
		logrus.Fatal("config validation failed: ", err)
		return err
	}

	logrus.Info("configuration loaded successfully from environment variables")
	return nil
}

// setDefaults sets default values for optional configuration
func setDefaults() {
	// Application defaults
	viper.SetDefault("env", "development")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("graceful_shutdown_timeout", "30s")

	// Server defaults
	viper.SetDefault("server.port", "8080")

	// Database defaults
	viper.SetDefault("database.ping_interval", "30s")
	viper.SetDefault("database.reconnect_factor", 2)
	viper.SetDefault("database.min_jitter", "200ms")
	viper.SetDefault("database.max_jitter", "500ms")
	viper.SetDefault("database.max_retry", 5)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 30)
	viper.SetDefault("database.max_conn_lifetime", "1h")

	// Redis defaults
	viper.SetDefault("redis.is_cache_disable", false)
	viper.SetDefault("redis.default_cache_duration", "15m")
	viper.SetDefault("redis.max_retry", 5)
	viper.SetDefault("redis.max_idle_conns", 5)
	viper.SetDefault("redis.max_active_conns", 20)
	viper.SetDefault("redis.max_conn_lifetime", "1h")

	// Token defaults
	viper.SetDefault("token.access_token_duration", "1h")
	viper.SetDefault("token.refresh_token_duration", "720h") // 30 days

	// Supabase defaults
	viper.SetDefault("supabase.bucket", "publication")
}

// bindEnvVariables explicitly binds all environment variables
func bindEnvVariables() {
	// Top level
	viper.BindEnv("env", "ENV")
	viper.BindEnv("log_level", "LOG_LEVEL")
	viper.BindEnv("graceful_shutdown_timeout", "GRACEFUL_SHUTDOWN_TIMEOUT")

	// Server
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.base_url", "SERVER_BASE_URL")
	viper.BindEnv("server.frontend_url", "SERVER_FRONTEND_URL", "FRONTEND_URL")

	// Database
	viper.BindEnv("database.dsn", "DATABASE_DSN")
	viper.BindEnv("database.direct_dsn", "DATABASE_DIRECT_DSN") // For migrations
	viper.BindEnv("database.ping_interval", "DATABASE_PING_INTERVAL")
	viper.BindEnv("database.reconnect_factor", "DATABASE_RECONNECT_FACTOR")
	viper.BindEnv("database.min_jitter", "DATABASE_MIN_JITTER")
	viper.BindEnv("database.max_jitter", "DATABASE_MAX_JITTER")
	viper.BindEnv("database.max_retry", "DATABASE_MAX_RETRY")
	viper.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	viper.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_conn_lifetime", "DATABASE_MAX_CONN_LIFETIME")

	// Redis
	viper.BindEnv("redis.is_cache_disable", "REDIS_IS_CACHE_DISABLE")
	viper.BindEnv("redis.cache_dsn", "REDIS_CACHE_DSN")
	viper.BindEnv("redis.default_cache_duration", "REDIS_DEFAULT_CACHE_DURATION")
	viper.BindEnv("redis.max_retry", "REDIS_MAX_RETRY")
	viper.BindEnv("redis.max_idle_conns", "REDIS_MAX_IDLE_CONNS")
	viper.BindEnv("redis.max_active_conns", "REDIS_MAX_ACTIVE_CONNS")
	viper.BindEnv("redis.max_conn_lifetime", "REDIS_MAX_CONN_LIFETIME")

	// Token
	viper.BindEnv("token.password_salt", "TOKEN_PASSWORD_SALT")
	viper.BindEnv("token.access_token_secret", "TOKEN_ACCESS_TOKEN_SECRET")
	viper.BindEnv("token.access_token_duration", "TOKEN_ACCESS_TOKEN_DURATION")
	viper.BindEnv("token.refresh_token_secret", "TOKEN_REFRESH_TOKEN_SECRET")
	viper.BindEnv("token.refresh_token_duration", "TOKEN_REFRESH_TOKEN_DURATION")
	viper.BindEnv("token.forgot_password_duration", "TOKEN_FORGOT_PASSWORD_DURATION")
	viper.BindEnv("token.forgot_password_rate_limit", "TOKEN_FORGOT_PASSWORD_RATE_LIMIT")
	viper.BindEnv("token.forgot_password_rate_window", "TOKEN_FORGOT_PASSWORD_RATE_WINDOW")

	// SMTP
	viper.BindEnv("smtp.host", "SMTP_HOST")
	viper.BindEnv("smtp.port", "SMTP_PORT")
	viper.BindEnv("smtp.username", "SMTP_USERNAME")
	viper.BindEnv("smtp.password", "SMTP_PASSWORD")
	viper.BindEnv("smtp.from_email", "SMTP_FROM_EMAIL")

	// Supabase
	viper.BindEnv("supabase.url", "SUPABASE_URL")
	viper.BindEnv("supabase.service_role_key", "SUPABASE_SERVICE_ROLE_KEY")
	viper.BindEnv("supabase.anon_role_key", "SUPABASE_ANON_ROLE_KEY")
	viper.BindEnv("supabase.bucket", "SUPABASE_BUCKET")
}

// Validate checks if all required configuration values are set correctly
func (c *EnvConfig) Validate() error {
	var errs []string

	// Validate required fields
	if c.Database.DSN == "" {
		errs = append(errs, "DATABASE_DSN is required")
	}

	if c.Token.PasswordSalt == "" {
		errs = append(errs, "TOKEN_PASSWORD_SALT is required")
	} else if len(c.Token.PasswordSalt) < 16 {
		errs = append(errs, "TOKEN_PASSWORD_SALT must be at least 16 characters")
	}

	if c.Token.AccessTokenSecret == "" {
		errs = append(errs, "TOKEN_ACCESS_TOKEN_SECRET is required")
	} else if len(c.Token.AccessTokenSecret) < 16 {
		errs = append(errs, "TOKEN_ACCESS_TOKEN_SECRET must be at least 16 characters")
	}

	if c.Token.RefreshTokenSecret == "" {
		errs = append(errs, "TOKEN_REFRESH_TOKEN_SECRET is required")
	} else if len(c.Token.RefreshTokenSecret) < 16 {
		errs = append(errs, "TOKEN_REFRESH_TOKEN_SECRET must be at least 16 characters")
	}

	if c.Server.Port == "" {
		errs = append(errs, "SERVER_PORT is required")
	}

	// Validate Redis if cache is enabled
	if !c.Redis.IsCacheDisable && c.Redis.CacheDSN == "" {
		errs = append(errs, "REDIS_CACHE_DSN is required when cache is enabled")
	}

	// Validate environment
	validEnvs := map[string]bool{"development": true, "staging": true, "production": true}
	if !validEnvs[c.Env] {
		errs = append(errs, fmt.Sprintf("ENV must be one of: development, staging, production (got: %s)", c.Env))
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	if c.Supabase.URL == "" {
		errs = append(errs, "SUPABASE_URL is required")
	}
	if c.Supabase.ServiceRoleKey == "" {
		errs = append(errs, "SUPABASE_SERVICE_ROLE_KEY is required")
	}

	return nil
}
