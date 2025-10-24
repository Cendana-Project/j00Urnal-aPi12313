package permission

import (
	"github.com/api-monolith-template/internal/model/contract/service"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	permissionService service.PermissionContract
}

func NewController(permissionService service.PermissionContract) *Controller {
	return &Controller{
		permissionService: permissionService,
	}
}

func (c *Controller) GetAllPermissions(ctx *gin.Context) {
	var req request.GetAllPermissionsReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	resp, err := c.permissionService.GetAllPermissions(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) CreatePermission(ctx *gin.Context) {
	var req request.CreatePermissionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	resp, err := c.permissionService.CreatePermission(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) GetPermissionByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	req := request.GetPermissionByIDReq{ID: id}
	resp, err := c.permissionService.GetPermissionByID(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) UpdatePermission(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	var req request.UpdatePermissionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.HandleError(ctx, err)
		return
	}
	req.ID = id

	resp, err := c.permissionService.UpdatePermission(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) DeletePermission(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	req := request.DeletePermissionReq{ID: id}
	resp, err := c.permissionService.DeletePermission(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) AssignPermissionToRole(ctx *gin.Context) {
	var req request.AssignPermissionToRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	resp, err := c.permissionService.AssignPermissionToRole(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) RemovePermissionFromRole(ctx *gin.Context) {
	var req request.RemovePermissionFromRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	resp, err := c.permissionService.RemovePermissionFromRole(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) GetRolePermissions(ctx *gin.Context) {
	roleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	var req request.GetRolePermissionsReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		util.HandleError(ctx, err)
		return
	}
	req.RoleID = roleID

	resp, err := c.permissionService.GetRolePermissions(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}
