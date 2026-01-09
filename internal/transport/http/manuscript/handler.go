package manuscript

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/mapper"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/service/manuscript"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct {
	svc *manuscript.Service
}

func NewController(svc *manuscript.Service) *Controller {
	return &Controller{svc: svc}
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

	f, err := c.svc.UploadFile(ctx.Request.Context(), id, constant.FileType(fileType), fileHeader)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = f
	util.HandleResponse(ctx, res, nil)
}
