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

type AWSConfig struct {
	Region     string         `mapstructure:"region"`
	S3Bucket   string         `mapstructure:"s3_bucket"`
	ChatBucket S3BucketConfig `mapstructure:"chat_bucket"`
}

type S3BucketConfig struct {
	Region   string `mapstructure:"region"`
	S3Bucket string `mapstructure:"s3_bucket"`
}

type EnvConfig struct {
	Env                     string        `mapstructure:"env"`
	LogLevel                string        `mapstructure:"log_level"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
	Token                   Token         `mapstructure:"token"`
	Server                  Server        `mapstructure:"server"`
	Database                Database      `mapstructure:"database"`
	Redis                   Redis         `mapstructure:"redis"`
	SMTP                    SMTP          `mapstructure:"smtp"`
	AWS                     AWSConfig     `mapstructure:"aws"`
	Stripe                  Stripe        `mapstructure:"stripe"`
	Chat                    Chat          `mapstructure:"chat"`
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

type Stripe struct {
	SecretKey      string `mapstructure:"secret_key"`
	PublishableKey string `mapstructure:"publishable_key"`
	WebhookSecret  string `mapstructure:"webhook_secret"`

	DefaultCurrency     string   `mapstructure:"default_currency"`
	SupportedCurrencies []string `mapstructure:"supported_currencies"`

	MinAmounts map[string]float64 `mapstructure:"min_amounts"`
	MaxAmounts map[string]float64 `mapstructure:"max_amounts"`

	MaxRetries int           `mapstructure:"max_retries"`
	RetryDelay time.Duration `mapstructure:"retry_delay"`
}

type Chat struct {
	LogLevel  string          `mapstructure:"log_level"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
	Room      RoomConfig      `mapstructure:"room"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

type WebSocketConfig struct {
	ReadBufferSize        int           `mapstructure:"read_buffer_size"`
	WriteBufferSize       int           `mapstructure:"write_buffer_size"`
	MaxMessageSize        int64         `mapstructure:"max_message_size"`
	WriteWait             time.Duration `mapstructure:"write_wait"`
	PongWait              time.Duration `mapstructure:"pong_wait"`
	PingPeriod            time.Duration `mapstructure:"ping_period"`
	MaxConnectionsPerUser int           `mapstructure:"max_connections_per_user"`
	ConnectionTimeout     time.Duration `mapstructure:"connection_timeout"`

	MaxContentLength         int      `mapstructure:"max_content_length"`
	MaxAttachmentSize        int64    `mapstructure:"max_attachment_size"`
	SupportedAttachmentTypes []string `mapstructure:"supported_attachment_types"`

	EnableTypingIndicators bool `mapstructure:"enable_typing_indicators"`
	EnableReadReceipts     bool `mapstructure:"enable_read_receipts"`
	EnableMessageReactions bool `mapstructure:"enable_message_reactions"`
	EnableFileUploads      bool `mapstructure:"enable_file_uploads"`

	TypingIndicatorDuration   time.Duration `mapstructure:"typing_indicator_duration"`
	InactiveConnectionCleanup time.Duration `mapstructure:"inactive_connection_cleanup"`
	MessageCleanupBatchSize   int           `mapstructure:"message_cleanup_batch_size"`
}

type RoomConfig struct {
	MaxGroupSize             int    `mapstructure:"max_group_size"`
	MaxRoomNameLength        int    `mapstructure:"max_room_name_length"`
	MaxRoomDescriptionLength int    `mapstructure:"max_room_description_length"`
	DefaultRoomType          string `mapstructure:"default_room_type"`
	AutoArchiveAfterDays     int    `mapstructure:"auto_archive_after_days"`
	CleanupArchivedAfterDays int    `mapstructure:"cleanup_archived_after_days"`
}

type RateLimitConfig struct {
	MessagesPerMinute      int `mapstructure:"messages_per_minute"`
	ReactionsPerMinute     int `mapstructure:"reactions_per_minute"`
	TypingUpdatesPerMinute int `mapstructure:"typing_updates_per_minute"`
	RoomCreationsPerHour   int `mapstructure:"room_creations_per_hour"`
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
