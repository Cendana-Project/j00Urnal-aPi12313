package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	hospRepo "github.com/api-monolith-template/internal/repository/hospital"
	roleRepo "github.com/api-monolith-template/internal/repository/role"
	"github.com/api-monolith-template/internal/util"
)

func RequireHospitalPermissions(hRepo *hospRepo.Repository, rRepo *roleRepo.Repository, required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			resp := constant.ErrUnauthorized.ToResponse()
			util.HandleResponse(c, &resp, nil)
			c.Abort()
			return
		}
		if hRepo == nil || rRepo == nil {
			resp := constant.ErrInternalServerError.ToResponse()
			util.HandleResponse(c, &resp, nil)
			c.Abort()
			return
		}

		hintVal, ok := c.Get("hospital_hint")
		if !ok {
			resp := constant.ErrValidationFailed.ToResponse()
			resp.Message = "hospital context required (X-Hospital-ID or X-Hospital-Code or :hospital_id)"
			util.HandleResponse(c, &resp, nil)
			c.Abort()
			return
		}
		hint := hintVal.(string)

		hospitalID, err := hRepo.ResolveHospitalID(c.Request.Context(), hint)
		if err != nil || hospitalID == "" {
			resp := constant.ErrRecordNotFound.ToResponse()
			resp.StatusCode = http.StatusNotFound
			resp.Message = "hospital not found or inactive"
			util.HandleResponse(c, &resp, nil)
			c.Abort()
			return
		}
		c.Set("hospital_id", hospitalID)

		if isSuper, _ := rRepo.IsUserSuperAdmin(c.Request.Context(), userID); isSuper {
			c.Next()
			return
		}

		perms, err := rRepo.ListHospitalPermissionsByUser(c.Request.Context(), hospitalID, userID)
		if err != nil {
			resp := constant.ErrInternalServerError.ToResponse()
			util.HandleResponse(c, &resp, nil)
			c.Abort()
			return
		}
		have := map[string]struct{}{}
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
