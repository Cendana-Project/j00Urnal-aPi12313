package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/util"
)

// UserRoleChecker checks whether a user has a role (implemented by role.Repository).
type UserRoleChecker interface {
	UserHasRole(userID, roleSlug string) (bool, error)
}

// RequireReviewer ensures the authenticated user has the REVIEWER role.
func RequireReviewer(roles UserRoleChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if roles == nil {
			resp := constant.ErrInternalServerError.ToResponse()
			util.HandleResponse(c, &resp, nil)
			c.Abort()
			return
		}
		userID := util.GetUserID(c)
		if userID == "" {
			util.HandleError(c, constant.ErrUnauthorized)
			c.Abort()
			return
		}
		ok, err := roles.UserHasRole(userID, constant.RoleReviewer)
		if err != nil {
			util.HandleError(c, constant.ErrInternalServerError)
			c.Abort()
			return
		}
		if !ok {
			util.HandleError(c, constant.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
