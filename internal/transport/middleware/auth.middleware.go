package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9" // <=== added

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/util"
)

// key builder (local, simple)
func accessBlacklistKey(jti string) string { return "access:blacklist:" + jti } // <=== added

// AuthRequired validates Bearer access token (JWT HS256) + blacklist check. // <=== changed
func AuthRequired(rdb *redis.Client) gin.HandlerFunc { // <=== changed (accept redis client)
	secret := []byte(config.Env.JWT.Secret)

	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			res := constant.ErrUnauthorized.ToResponse()
			util.HandleResponse(c, &res, nil)
			c.Abort()
			return
		}
		raw := strings.TrimSpace(h[7:])

		tok, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
			// only HS256
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return secret, nil
		})
		if err != nil || !tok.Valid {
			res := constant.ErrInvalidToken.ToResponse()
			util.HandleResponse(c, &res, nil)
			c.Abort()
			return
		}

		claims, ok := tok.Claims.(jwt.MapClaims)
		if !ok || claims["typ"] != "access" {
			res := constant.ErrInvalidToken.ToResponse()
			util.HandleResponse(c, &res, nil)
			c.Abort()
			return
		}

		sub, _ := claims["sub"].(string)
		jti, _ := claims["jti"].(string) // <=== added
		if sub == "" || jti == "" {      // <=== changed
			res := constant.ErrInvalidToken.ToResponse()
			util.HandleResponse(c, &res, nil)
			c.Abort()
			return
		}

		// Blacklist check // <=== added
		if rdb != nil {
			if n, err := rdb.Exists(c, accessBlacklistKey(jti)).Result(); err == nil && n > 0 {
				res := constant.ErrUnauthorized.ToResponse()
				util.HandleResponse(c, &res, nil)
				c.Abort()
				return
			}
		}

		c.Set("user_id", sub)
		c.Set("token_id", jti) // <=== added
		c.Next()
	}
}
