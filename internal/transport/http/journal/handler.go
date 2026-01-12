package journal

import (
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/mapper"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/service/journal"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct {
	svc *journal.Service
}

func NewController(svc *journal.Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) Create(ctx *gin.Context) {
	var req request.CreateJournalRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	userID := util.GetUserID(ctx)
	j, err := c.svc.Create(ctx.Request.Context(), userID, req.Name, req.Description)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.StatusCode = http.StatusCreated
	res.Data = mapper.ToJournalResponse(j)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var req request.UpdateJournalRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	j, err := c.svc.Update(ctx.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToJournalResponse(j)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) SetStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	var req request.UpdateJournalStatusRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	j, err := c.svc.SetStatus(ctx.Request.Context(), id, constant.PublicationStatus(req.Status))
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToJournalResponse(j)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) UploadCover(ctx *gin.Context) {
	id := ctx.Param("id")

	// Try multiple field names
	fieldNames := []string{"cover", "file", "image"}
	var fileHeader *multipart.FileHeader
	var err error

	for _, field := range fieldNames {
		fileHeader, err = ctx.FormFile(field)
		if err == nil {
			break
		}
	}

	if err != nil {
		// Log the actual error for debugging
		util.Infof(ctx.Request.Context(), "UploadCover failed to find file in fields %v: %v", fieldNames, err)
		util.HandleError(ctx, constant.ErrValidationFailed)
		return
	}

	// Validate File: Max 5MB, Images only
	if err := util.ValidateFile(fileHeader, 5*1024*1024, []string{"image/jpeg", "image/png", "image/webp"}); err != nil {
		util.HandleError(ctx, response.CustomError{
			Code:       constant.ErrValidationFailed.Code,
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	userID := util.GetUserID(ctx)
	url, err := c.svc.UploadCover(ctx.Request.Context(), id, fileHeader, userID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = gin.H{"cover_path": url}
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	j, err := c.svc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	if j == nil {
		util.HandleError(ctx, constant.ErrRecordNotFound)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToJournalResponse(j)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.svc.Delete(ctx.Request.Context(), id); err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Message = "Journal deleted successfully"
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) GetAll(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	status := ctx.Query("status")

	var statusPtr *constant.PublicationStatus
	if status != "" {
		s := constant.PublicationStatus(status)
		statusPtr = &s
	}

	journals, total, err := c.svc.GetAll(ctx.Request.Context(), statusPtr, page, limit)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	journalResponses := make([]response.JournalResponse, len(journals))
	for i, j := range journals {
		journalResponses[i] = mapper.ToJournalResponse(&j)
	}

	res := response.NewResponseOK()
	res.Data = journalResponses
	res.Meta = &response.Meta{
		Page:      page,
		PageSize:  limit,
		TotalData: int(total),
	}
	util.HandleResponse(ctx, res, nil)
}
