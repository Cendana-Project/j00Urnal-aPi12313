package config

import (
	"strings"
	"time"

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
	JWT                     JWTConfig     `mapstructure:"jwt"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`

	// BARU (lebih fleksibel): gunakan duration string, contoh: "15m", "720h"
	AccessTTL  string `mapstructure:"access_ttl"`
	RefreshTTL string `mapstructure:"refresh_ttl"`

	// Fallback lama (tetap didukung agar tidak breaking)
	AccessTTLMinutes int `mapstructure:"access_ttl_minutes"`
	RefreshTTLDays   int `mapstructure:"refresh_ttl_days"`
}

// ParseDurations mengembalikan durasi access & refresh dengan prioritas:
// AccessTTL/RefreshTTL (string) > AccessTTLMinutes/RefreshTTLDays (int) > default (15m, 30d)
func (j JWTConfig) ParseDurations() (access time.Duration, refresh time.Duration) {
	// Access
	if d, err := time.ParseDuration(j.AccessTTL); err == nil && d > 0 {
		access = d
	} else if j.AccessTTLMinutes > 0 {
		access = time.Duration(j.AccessTTLMinutes) * time.Minute
	} else {
		access = 15 * time.Minute
	}
	// Refresh
	if d, err := time.ParseDuration(j.RefreshTTL); err == nil && d > 0 {
		refresh = d
	} else if j.RefreshTTLDays > 0 {
		refresh = time.Duration(j.RefreshTTLDays) * 24 * time.Hour
	} else {
		refresh = 30 * 24 * time.Hour
	}
	return
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
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	// Bind env explicitly untuk nested struct
	bindEnvVariables()

	// File config opsional; ENV saja juga boleh
	if err := viper.ReadInConfig(); err != nil {
		logrus.Warn("config.yml not found or unreadable; relying on environment variables")
	}

	if err := viper.Unmarshal(&Env); err != nil {
		logrus.Fatal("failed to unmarshal config file: ", err)
		return err
	}
	return nil
}

// bindEnvVariables: pastikan field terikat ENV vars
func bindEnvVariables() {
	// Top level
	viper.BindEnv("env", "ENV")
	viper.BindEnv("log_level", "LOG_LEVEL")
	viper.BindEnv("graceful_shutdown_timeout", "GRACEFUL_SHUTDOWN_TIMEOUT")

	// Server
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.base_url", "SERVER_BASE_URL")

	// Database
	viper.BindEnv("database.dsn", "DATABASE_DSN")
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
	viper.BindEnv("token.forgot_password_rate_limit", "TOKEN_FORGOT_PASSWORD_RATE_LIMIT")
	viper.BindEnv("token.forgot_password_rate_window", "TOKEN_FORGOT_PASSWORD_RATE_WINDOW")

	// JWT
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.access_ttl", "JWT_ACCESS_TTL")   // e.g. "15m"
	viper.BindEnv("jwt.refresh_ttl", "JWT_REFRESH_TTL") // e.g. "720h"
	viper.BindEnv("jwt.access_ttl_minutes", "JWT_ACCESS_TTL_MINUTES")
	viper.BindEnv("jwt.refresh_ttl_days", "JWT_REFRESH_TTL_DAYS")

	// SMTP
	viper.BindEnv("smtp.host", "SMTP_HOST")
	viper.BindEnv("smtp.port", "SMTP_PORT")
	viper.BindEnv("smtp.username", "SMTP_USERNAME")
	viper.BindEnv("smtp.password", "SMTP_PASSWORD")
	viper.BindEnv("smtp.from", "SMTP_FROM")
}
