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
	Secret           string `mapstructure:"secret"`
	AccessTTLMinutes int    `mapstructure:"access_ttl_minutes"`
	RefreshTTLDays   int    `mapstructure:"refresh_ttl_days"`
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

	// Explicitly bind environment variables for nested config structures
	// This ensures environment variables properly override config file values
	bindEnvVariables()

	// Make config file optional in environments like Render; allow ENV-only config
	if err := viper.ReadInConfig(); err != nil {
		logrus.Warn("config.yml not found or unreadable; relying on environment variables")
	}

	// Unmarshal combined config (file + overridden by ENV)
	if err := viper.Unmarshal(&Env); err != nil {
		logrus.Fatal("failed to unmarshal config file: ", err)
		return err
	}

	return nil
}

// bindEnvVariables explicitly binds all environment variables
// This is necessary because viper.AutomaticEnv() doesn't work well with nested structs
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
}
