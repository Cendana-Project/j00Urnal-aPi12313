package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/util"
)

// AuthRequired validates Bearer access token (JWT HS256).
func AuthRequired() gin.HandlerFunc {
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
		if sub == "" {
			res := constant.ErrInvalidToken.ToResponse()
			util.HandleResponse(c, &res, nil)
			c.Abort()
			return
		}
		c.Set("user_id", sub)
		c.Next()
	}
}
