package volume

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/mapper"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/repository/role"
	"github.com/api-monolith-template/internal/service/volume"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct {
	svc      *volume.Service
	roleRepo *role.Repository
}

func NewController(svc *volume.Service, roleRepo *role.Repository) *Controller {
	return &Controller{svc: svc, roleRepo: roleRepo}
}

func (c *Controller) Create(ctx *gin.Context) {
	journalID := ctx.Param("id") // /journals/:id/volumes
	var req request.CreateVolumeRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	v, err := c.svc.Create(ctx.Request.Context(), journalID, req.Year, req.Number)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.StatusCode = http.StatusCreated
	res.Data = mapper.ToVolumeResponse(v)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) GetAll(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	isPrivileged := false
	if userID != "" {
		roles, err := c.roleRepo.ListRolesByUser(ctx.Request.Context(), userID)
		if err == nil {
			for _, r := range roles {
				if r.Slug == constant.RoleSuperAdmin || r.Slug == constant.RoleChiefEditor {
					isPrivileged = true
					break
				}
			}
		}
	}

	volumes, err := c.svc.GetAll(ctx.Request.Context())
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	// Filter to PUBLISHED only for public users
	if !isPrivileged {
		filtered := make([]*entity.Volume, 0, len(volumes))
		for _, v := range volumes {
			if v != nil && v.Status == constant.PublicationStatusPublished {
				filtered = append(filtered, v)
			}
		}
		volumes = filtered
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToVolumeResponses(volumes)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var req request.UpdateVolumeRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	v, err := c.svc.Update(ctx.Request.Context(), id, req.Year, req.Number)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToVolumeResponse(v)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) SetStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	var req request.UpdateVolumeStatusRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	v, err := c.svc.SetStatus(ctx.Request.Context(), id, constant.PublicationStatus(req.Status))
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToVolumeResponse(v)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	v, err := c.svc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	if v == nil {
		util.HandleError(ctx, constant.ErrRecordNotFound)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToVolumeResponse(v)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.svc.Delete(ctx.Request.Context(), id); err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Message = "Volume deleted successfully"
	util.HandleResponse(ctx, res, nil)
}
