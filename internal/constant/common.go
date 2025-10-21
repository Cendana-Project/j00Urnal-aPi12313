package constant

type TokenType string
type ContextKey string

const (
	ProductionEnvironment = "production"

	RequestID = ContextKey("reqId")
	UserID    = ContextKey("user_id")
	TokenID   = ContextKey("jti")

	DB = "db"

	// Token types
	AccessTokenType  TokenType = "access_token"
	RefreshTokenType TokenType = "refresh_token"
)

const (
	// Cache key for all users
	UserAllCacheKey = "user:all"
)

const (
	RateLimitKey = "rate-limit"
)

const (
	// Availability defaults
	DefaultAvailabilityDays = 7
	MaxAvailabilityDays     = 14
	DefaultTimezone         = "Europe/Paris"

	// Pagination defaults (fallback umum)
	DefaultPage     = 1
	DefaultPageSize = 50
	MaxPageSize     = 100
)
