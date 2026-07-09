package issue

import (
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/mapper"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/repository/role"
	"github.com/api-monolith-template/internal/service/issue"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct {
	svc      *issue.Service
	roleRepo *role.Repository
}

func NewController(svc *issue.Service, roleRepo *role.Repository) *Controller {
	return &Controller{svc: svc, roleRepo: roleRepo}
}

func (c *Controller) Create(ctx *gin.Context) {
	volumeID := ctx.Param("id") // /volumes/:id/issues
	var req request.CreateIssueRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	pubDate, err := time.Parse("2006-01-02", req.PublicationDate)
	if err != nil {
		util.HandleError(ctx, constant.ErrInvalidDateFormat)
		return
	}

	i, err := c.svc.Create(ctx.Request.Context(), volumeID, req.Number, pubDate)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.StatusCode = http.StatusCreated
	res.Data = mapper.ToIssueResponse(i)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) GetAll(ctx *gin.Context) {
	userID := util.GetUserID(ctx)
	// Public role check: only SuperAdmin / ChiefEditor see non-PUBLISHED issues
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

	issues, err := c.svc.GetAll(ctx.Request.Context())
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	// Filter to PUBLISHED only for public users
	if !isPrivileged {
		filtered := make([]*entity.Issue, 0, len(issues))
		for _, i := range issues {
			if i != nil && i.Status == constant.PublicationStatusPublished {
				filtered = append(filtered, i)
			}
		}
		issues = filtered
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToIssueResponses(issues)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) UpdateMetadata(ctx *gin.Context) {
	id := ctx.Param("id")
	var req request.UpdateIssueRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	pubDate, err := time.Parse("2006-01-02", req.PublicationDate)
	if err != nil {
		util.HandleError(ctx, constant.ErrInvalidDateFormat)
		return
	}

	i, err := c.svc.Update(ctx.Request.Context(), id, req.Number, pubDate, constant.PublicationStatus(req.Status))
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToIssueResponse(i)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) SetStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	var req request.UpdateIssueStatusRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	// We need to fetch existing issue to get number and date if we only update status
	// Since service Update() requires all fields (my design flaw? or I should add SetStatus to service separately?)
	// I added SetStatus logic in Update() but Update() takes all fields.
	// I should probably add specific SetStatus method in service or fetch-then-update here.

	// Let's rely on fetch-then-update here for simplicity given current service design.
	// Actually, issue service Update() signature: Update(ctx, id, number, pubDate, status)
	// I don't have existing number/date here.
	// I should add `SetStatus` to IssueService or `GetByID` then update.
	// For better performance/design, `SetStatus` in service is better.
	// I will fetch and update for now as I can't easily change service without another tool call cycle overhead (mental check: I can change service, but file is already written).
	// Let's modify service? No, let's fetch here.

	current, err := c.svc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	if current == nil {
		util.HandleError(ctx, constant.ErrRecordNotFound)
		return
	}

	updated, err := c.svc.Update(ctx.Request.Context(), id, current.Number, current.PublicationDate, constant.PublicationStatus(req.Status))
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToIssueResponse(updated)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) UploadFile(ctx *gin.Context) {
	// id := ctx.Param("id") // Unused
	util.HandleError(ctx, constant.ErrEndpointNotFound)

	// If endpoint is specific (e.g. /issues/:id/cover), we know type.
	// Requirements: POST /issues/{id}/cover, POST /issues/{id}/full-pdf
	// So I should have separate handler methods.

	// Let's check which route called this? Or just split them.
	// I'll implementing UploadCover and UploadPDF separately.
	util.HandleError(ctx, constant.ErrEndpointNotFound)
}

func (c *Controller) UploadCover(ctx *gin.Context) {
	c.handleUpload(ctx, constant.FileTypeCover, "cover")
}

func (c *Controller) UploadPDF(ctx *gin.Context) {
	c.handleUpload(ctx, constant.FileTypeFullIssuePDF, "pdf") // Form field name 'pdf'? or 'file'?
	// User said: "ONE full issue PDF" matches "full-issue.pdf" logic. Form field name could be 'file'.
}

func (c *Controller) handleUpload(ctx *gin.Context, fType constant.FileType, fieldName string) {
	id := ctx.Param("id")
	// Try multiple field names
	fieldNames := []string{fieldName, "file", "image", "pdf"}
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
		util.Infof(ctx.Request.Context(), "Issue.handleUpload failed to find file in fields %v: %v", fieldNames, err)
		util.HandleError(ctx, constant.ErrValidationError)
		return
	}

	// Validate based on type
	var maxBytes int64
	var allowedMimes []string

	if fType == constant.FileTypeCover {
		maxBytes = 5 * 1024 * 1024 // 5MB
		allowedMimes = []string{"image/jpeg", "image/png", "image/webp"}
	} else if fType == constant.FileTypeFullIssuePDF {
		maxBytes = 50 * 1024 * 1024 // 50MB
		allowedMimes = []string{"application/pdf"}
	}

	if maxBytes > 0 {
		if err := util.ValidateFile(fileHeader, maxBytes, allowedMimes); err != nil {
			util.HandleError(ctx, response.CustomError{
				Code:       constant.ErrValidationFailed.Code,
				StatusCode: http.StatusBadRequest,
				Message:    err.Error(),
			})
			return
		}
	}

	userID := util.GetUserID(ctx)
	url, err := c.svc.UploadFile(ctx.Request.Context(), id, fileHeader, fType, userID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Data = gin.H{"file_path": url}
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := c.svc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	if i == nil {
		util.HandleError(ctx, constant.ErrRecordNotFound)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToIssueResponse(i)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.svc.Delete(ctx.Request.Context(), id); err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.Message = "Issue deleted successfully"
	util.HandleResponse(ctx, res, nil)
}
