package middleware

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/util"
)

type PermissionLister interface {
	ListPermissionsByUser(ctx context.Context, userID string) ([]entity.Permission, error)
}

// RequirePermissions memastikan user (dari JWT) memiliki minimal
// salah satu dari permission 'required'.
func RequirePermissions(roleRepo PermissionLister, required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if roleRepo == nil {
			// repo belum diinject: balas 500 yang rapi, jangan panic
			resp := constant.ErrInternalServerError.ToResponse()
			util.HandleResponse(c, &resp, nil)
			c.Abort()
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			resp := constant.ErrUnauthorized.ToResponse()
			util.HandleResponse(c, &resp, nil)
			c.Abort()
			return
		}

		perms, err := roleRepo.ListPermissionsByUser(c.Request.Context(), userID)
		if err != nil {
			resp := constant.ErrInternalServerError.ToResponse()
			util.HandleResponse(c, &resp, nil)
			c.Abort()
			return
		}

		have := make(map[string]struct{}, len(perms))
		for _, p := range perms {
			have[p.Slug] = struct{}{}
		}

		allowed := false
		for _, need := range required {
			if _, ok := have[need]; ok {
				allowed = true
				break
			}
		}
		if !allowed {
			resp := constant.ErrForbidden.ToResponse()
			util.HandleResponse(c, &resp, nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
