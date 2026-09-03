package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/util"
)

type accessBlacklistChecker interface {
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
}

type accessTokenStatus uint8

const (
	accessTokenMissing accessTokenStatus = iota
	accessTokenInvalid
	accessTokenValid
	accessTokenBlacklisted
	accessTokenBlacklistUnavailable
)

func accessBlacklistKey(jti string) string { return "access:blacklist:" + jti }

// authenticateAccessToken validates a Bearer access token and checks whether its jti is
// blacklisted. A Redis failure is kept distinct from a bad token so public optional-auth routes
// can safely fall back to anonymous access without returning an outage to public visitors.
func authenticateAccessToken(
	ctx context.Context,
	authorization string,
	secret []byte,
	blacklist accessBlacklistChecker,
) (string, string, accessTokenStatus) {
	parts := strings.Fields(authorization)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", "", accessTokenMissing
	}

	tok, err := jwt.Parse(
		parts[1],
		func(_ *jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !tok.Valid {
		return "", "", accessTokenInvalid
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", accessTokenInvalid
	}
	typ, _ := claims["typ"].(string)
	sub, _ := claims["sub"].(string)
	jti, _ := claims["jti"].(string)
	sub = strings.TrimSpace(sub)
	jti = strings.TrimSpace(jti)
	if typ != "access" || sub == "" || jti == "" {
		return "", "", accessTokenInvalid
	}

	if blacklist != nil {
		blacklisted, err := blacklist.Exists(ctx, accessBlacklistKey(jti)).Result()
		if err != nil {
			return sub, jti, accessTokenBlacklistUnavailable
		}
		if blacklisted > 0 {
			return "", "", accessTokenBlacklisted
		}
	}

	return sub, jti, accessTokenValid
}

func setAuthenticatedContext(c *gin.Context, userID, tokenID string) {
	c.Set(string(constant.UserID), userID)
	c.Set(string(constant.TokenID), tokenID)
}

// AuthOptional enriches a public request when it carries a valid, non-blacklisted access token.
// Missing, malformed, expired, revoked, or unverifiable tokens remain anonymous and never turn a
// public endpoint into an authentication error.
func AuthOptional(rdb *redis.Client) gin.HandlerFunc {
	if rdb == nil {
		return authOptionalWithBlacklist(nil)
	}
	return authOptionalWithBlacklist(rdb)
}

func authOptionalWithBlacklist(blacklist accessBlacklistChecker) gin.HandlerFunc {
	secret := []byte(config.Env.Token.AccessTokenSecret)

	return func(c *gin.Context) {
		userID, tokenID, status := authenticateAccessToken(
			c.Request.Context(),
			c.GetHeader("Authorization"),
			secret,
			blacklist,
		)
		if status == accessTokenValid {
			setAuthenticatedContext(c, userID, tokenID)
		}
		c.Next()
	}
}

// AuthRequired validates a Bearer access token (JWT HS256) and checks its blacklist status.
func AuthRequired(rdb *redis.Client) gin.HandlerFunc {
	if rdb == nil {
		return authRequiredWithBlacklist(nil)
	}
	return authRequiredWithBlacklist(rdb)
}

func authRequiredWithBlacklist(blacklist accessBlacklistChecker) gin.HandlerFunc {
	secret := []byte(config.Env.Token.AccessTokenSecret)

	return func(c *gin.Context) {
		userID, tokenID, status := authenticateAccessToken(
			c.Request.Context(),
			c.GetHeader("Authorization"),
			secret,
			blacklist,
		)

		switch status {
		case accessTokenValid:
			setAuthenticatedContext(c, userID, tokenID)
			c.Next()
		case accessTokenBlacklistUnavailable:
			// Preserve the established protected-route behavior during a Redis outage. The token
			// itself is valid; Redis is an optional cache dependency in this deployment.
			setAuthenticatedContext(c, userID, tokenID)
			c.Next()
		case accessTokenMissing, accessTokenBlacklisted:
			res := constant.ErrUnauthorized.ToResponse()
			util.HandleResponse(c, &res, nil)
			c.Abort()
		default:
			res := constant.ErrInvalidToken.ToResponse()
			util.HandleResponse(c, &res, nil)
			c.Abort()
		}
	}
}
