package manuscript

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/mapper"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	rolerepo "github.com/api-monolith-template/internal/repository/role"
	"github.com/api-monolith-template/internal/service/manuscript"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct {
	svc      *manuscript.Service
	roleRepo *rolerepo.Repository
}

func NewController(svc *manuscript.Service, roleRepo *rolerepo.Repository) *Controller {
	return &Controller{svc: svc, roleRepo: roleRepo}
}

// checkAccess verifies if the user has permission to manage manuscript OR is the owner (Main Author).
func (c *Controller) checkAccess(ctx *gin.Context, manuscriptID string) error {
	userID := util.GetUserID(ctx)
	if userID == "" {
		return constant.ErrUnauthorized
	}

	// 1. Check if user is SuperAdmin / Editor (Has ManuscriptManage permission)
	perms, err := c.roleRepo.ListPermissionsByUser(ctx.Request.Context(), userID)
	if err == nil {
		for _, p := range perms {
			if p.Slug == constant.PermissionManuscriptManage {
				return nil // Allowed by Permission
			}
		}
	} else {
		util.Infof(ctx.Request.Context(), "checkAccess: failed to list permissions for user %s: %v", userID, err)
	}

	// 2. Check Ownership
	m, err := c.svc.GetByID(ctx.Request.Context(), manuscriptID)
	if err != nil {
		return err
	}
	if m == nil {
		return constant.ErrRecordNotFound
	}
	if m.MainAuthorID == userID {
		return nil // Allowed by Ownership
	}

	return constant.ErrForbidden
}

func (c *Controller) Create(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	var req request.CreateManuscriptRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	m, err := c.svc.Create(ctx.Request.Context(), userID, req.IssueID, req.Title, req.Abstract)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.StatusCode = http.StatusCreated
	res.Data = mapper.ToManuscriptResponse(m)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) GetAll(ctx *gin.Context) {
	issueID := ctx.Query("issue_id")
	if issueID == "" {
		util.HandleError(ctx, constant.ErrValidationError)
		return
	}

	ms, err := c.svc.ListByIssue(ctx.Request.Context(), issueID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToManuscriptListResponse(ms)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	m, err := c.svc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	if m == nil {
		util.HandleError(ctx, constant.ErrRecordNotFound)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToManuscriptResponse(m)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var req request.UpdateManuscriptRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	if err := c.checkAccess(ctx, id); err != nil {
		util.HandleError(ctx, err)
		return
	}

	m, err := c.svc.Update(ctx.Request.Context(), id, req.Title, req.Abstract)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToManuscriptResponse(m)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.checkAccess(ctx, id); err != nil {
		util.HandleError(ctx, err)
		return
	}

	if err := c.svc.Delete(ctx.Request.Context(), id); err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Message = "Manuscript deleted successfully"
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) UpdateAuthors(ctx *gin.Context) {
	id := ctx.Param("id")
	var req request.UpdateManuscriptAuthorsRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	if err := c.checkAccess(ctx, id); err != nil {
		util.HandleError(ctx, err)
		return
	}

	authors := make([]entity.ManuscriptAuthor, len(req.Authors))
	for i, a := range req.Authors {
		authors[i] = entity.ManuscriptAuthor{
			UserID:          a.UserID,
			AuthorName:      a.AuthorName,
			AuthorEmail:     a.AuthorEmail,
			Affiliation:     a.Affiliation,
			IsCorresponding: a.IsCorresponding,
			OrderPosition:   a.OrderPosition,
		}
	}

	if err := c.svc.UpdateAuthors(ctx.Request.Context(), id, authors); err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Message = "Authors updated successfully"
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) UploadMainFile(ctx *gin.Context) {
	id := ctx.Param("id")
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		util.HandleError(ctx, constant.ErrValidationError)
		return
	}

	// Validate Main File: Max 20MB, PDF/DOC/DOCX
	allowedMimes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
	if err := util.ValidateFile(fileHeader, 20*1024*1024, allowedMimes); err != nil {
		util.HandleError(ctx, response.CustomError{
			Code:       constant.ErrValidationFailed.Code,
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	if err := c.checkAccess(ctx, id); err != nil {
		util.HandleError(ctx, err)
		return
	}

	f, err := c.svc.UploadFile(ctx.Request.Context(), id, constant.FileTypeMain, fileHeader)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = f
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) UploadAttachment(ctx *gin.Context) {
	id := ctx.Param("id")
	fileType := ctx.PostForm("type")
	if fileType == "" {
		fileType = string(constant.FileTypeSupplement)
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		util.HandleError(ctx, constant.ErrValidationError)
		return
	}

	// Validate: Max 10MB, Images/PDF/Excel/CSV
	allowedMimes := []string{
		"image/jpeg", "image/png", "image/webp",
		"application/pdf",
		"text/csv", "text/plain",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}
	if err := util.ValidateFile(fileHeader, 10*1024*1024, allowedMimes); err != nil {
		util.HandleError(ctx, response.CustomError{
			Code:       constant.ErrValidationFailed.Code,
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	if err := c.checkAccess(ctx, id); err != nil {
		util.HandleError(ctx, err)
		return
	}

	f, err := c.svc.UploadFile(ctx.Request.Context(), id, constant.FileType(fileType), fileHeader)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = f
	util.HandleResponse(ctx, res, nil)
}
