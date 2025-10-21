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

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatal("failed to read config file: ", err)
	}

	err = viper.Unmarshal(&Env)
	if err != nil {
		logrus.Fatal("failed to unmarshal config file: ", err)
		return err
	}

	return nil
}
