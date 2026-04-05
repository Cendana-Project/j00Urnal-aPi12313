package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

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
			logrus.WithFields(logrus.Fields{
				"trace_id": util.GetTraceID(c),
				"path":     c.Request.URL.Path,
				"method":   c.Request.Method,
				"reason":   "missing_or_invalid_jwt",
			}).Warn("RequireReviewer: no user_id in context — send Authorization: Bearer <access_token>")
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
			logrus.WithFields(logrus.Fields{
				"trace_id": util.GetTraceID(c),
				"path":     c.Request.URL.Path,
				"method":   c.Request.Method,
				"user_id":  userID,
				"reason":   "user_lacks_REVIEWER_role",
			}).Warn("RequireReviewer: user is authenticated but does not have REVIEWER role")
			util.HandleError(c, constant.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
