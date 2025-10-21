package role

import (
	"context"
	"strconv"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/contract/service"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	roleService service.RoleContract
}

func NewController() *Controller {
	return new(Controller)
}

func (c *Controller) WithRoleService(svc service.RoleContract) *Controller {
	c.roleService = svc
	return c
}

func (c *Controller) CreateRole(ctx *gin.Context) {
	var req request.CreateRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	resp, err := c.roleService.CreateNewRole(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) AssignRoleToUser(ctx *gin.Context) {
	var req request.AssignRoleToUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	resp, err := c.roleService.AssignRoleToUser(ctx, &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) RemoveRoleFromUser(ctx *gin.Context) {
	var req request.RemoveRoleFromUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	resp, err := c.roleService.RemoveRoleFromUser(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) GetAllRoles(ctx *gin.Context) {
	var req request.GetAllRoles
	if err := ctx.ShouldBindQuery(&req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	resp, err := c.roleService.GetAllRoles(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) UpdateUserRole(ctx *gin.Context) {
	var req request.UpdateUserRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	// Get user ID from gin context and set it in request context
	userID := ctx.GetString("user_id")
	if userID == "" {
		userID = ctx.GetString(string(constant.UserID))
	}
	if userID == "" {
		util.HandleError(ctx, constant.ErrInvalidToken)
		return
	}

	reqCtx := context.WithValue(ctx.Request.Context(), "user_id", userID)
	reqCtx = context.WithValue(reqCtx, constant.UserID, userID)

	resp, err := c.roleService.UpdateUserRole(reqCtx, &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) UpdateRole(ctx *gin.Context) {
	roleIDStr := ctx.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	var req request.UpdateRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	req.RoleID = roleID

	resp, err := c.roleService.UpdateRole(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) GetRoleByID(ctx *gin.Context) {
	roleIDStr := ctx.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	req := &request.GetRoleByIdReq{
		RoleID: roleID,
	}

	resp, err := c.roleService.GetRoleById(ctx.Request.Context(), req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) DeleteRole(ctx *gin.Context) {
	roleIDStr := ctx.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	req := &request.DeleteRoleReq{
		RoleID: roleID,
	}

	resp, err := c.roleService.DeleteRole(ctx.Request.Context(), req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}

func (c *Controller) GetAllUserByRoles(ctx *gin.Context) {
	roleID := ctx.Param("id")
	if roleID == "" {
		util.HandleError(ctx, constant.ErrValidationError)
		return
	}

	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	req := request.GetAllUserByRolesReq{
		RoleID:   roleUUID,
		Page:     1,
		PageSize: 10,
	}

	if page := ctx.Query("page"); page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum > 0 {
			req.Page = pageNum
		}
	}
	if pageSize := ctx.Query("page_size"); pageSize != "" {
		if pageSizeNum, err := strconv.Atoi(pageSize); err == nil && pageSizeNum > 0 {
			req.PageSize = pageSizeNum
		}
	}

	resp, err := c.roleService.GetAllUserByRoles(ctx.Request.Context(), &req)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, resp, nil)
}
