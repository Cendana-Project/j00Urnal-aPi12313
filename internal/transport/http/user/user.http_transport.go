package user

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/contract/service"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	authService service.AuthService
	userService service.UserContract
	s3Service   service.S3UploaderService
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) WithAuthService(svc service.AuthService) *Controller {
	c.authService = svc
	return c
}

func (c *Controller) WithUserService(svc service.UserContract) *Controller {
	c.userService = svc
	return c
}

func (c *Controller) WithS3Service(svc service.S3UploaderService) *Controller {
	c.s3Service = svc
	return c
}

func (c *Controller) GetAllUsers(ctx *gin.Context) {
	resp, err := c.userService.GetAllUsers(ctx)
	util.HandleResponse(ctx, resp, err)
}

func (c *Controller) GetAllUsersSimple(ctx *gin.Context) {
	resp, err := c.userService.GetAllUsersSimple(ctx)
	util.HandleResponse(ctx, resp, err)
}

func (c *Controller) GetByIdentifier(ctx *gin.Context) {
	var identifier string

	identifier = ctx.Param("id")

	if identifier == "" {
		identifier = ctx.Query("identifier")
	}

	if identifier == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "identifier is required as path parameter or query parameter",
		})
		return
	}

	resp, err := c.userService.GetByIdentifier(ctx, identifier)
	util.HandleResponse(ctx, resp, err)
}

func (c *Controller) GetByIdentifierSimple(ctx *gin.Context) {
	identifier := ctx.Param("id")
	if identifier == "" {
		identifier = ctx.Query("identifier")
	}
	if identifier == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "identifier is required as path parameter or query parameter",
		})
		return
	}

	resp, err := c.userService.GetByIdentifierSimple(ctx, identifier)
	util.HandleResponse(ctx, resp, err)
}

func (c *Controller) DeleteUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	if _, err := uuid.Parse(userID); err != nil {
		util.HandleError(ctx, constant.ErrInvalidUUIDFormat)
		return
	}

	resp, err := c.userService.Delete(ctx, userID)
	util.HandleResponse(ctx, resp, err)
}

func (c *Controller) UpsertUser(ctx *gin.Context) {
	authenticatedUserID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		util.HandleError(ctx, constant.ErrInvalidToken)
		return
	}

	idParam := ctx.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		util.HandleError(ctx, constant.ErrInvalidUUIDFormat)
		return
	}

	if userID != *authenticatedUserID {
		util.HandleError(ctx, constant.ErrUnauthorizedUpdate)
		return
	}

	baseResponse, err := c.userService.GetByIdentifier(ctx, userID.String())
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	user, ok := baseResponse.Data.(*entity.User)
	if !ok || user == nil {
		util.HandleError(ctx, constant.ErrUserNotFound)
		return
	}

	contentType := ctx.GetHeader("Content-Type")
	var updateReq request.UpdateUserReq

	util.NewDefaultLogger(ctx).WithField("content_type", contentType).Info("Processing user update request")

	if strings.Contains(contentType, "multipart/form-data") {
		util.NewDefaultLogger(ctx).Info("Detected multipart form data, starting processing")

		form, err := ctx.MultipartForm()
		if err != nil {
			util.NewDefaultLogger(ctx).WithField("multipart_error", err.Error()).Error("Failed to parse multipart form")
			util.HandleError(ctx, err)
			return
		}
		util.NewDefaultLogger(ctx).Info("Successfully parsed multipart form")
		defer form.RemoveAll()

		if err := ctx.ShouldBind(&updateReq); err != nil {
			util.NewDefaultLogger(ctx).WithField("bind_error", err.Error()).Error("Failed to bind multipart form data")
			util.HandleError(ctx, err)
			return
		}
		util.NewDefaultLogger(ctx).WithField("bound_data", updateReq).Info("Successfully bound multipart form data")

		if photoFile, ok := form.File["photo_url"]; ok && len(photoFile) > 0 {
			util.NewDefaultLogger(ctx).Info("Photo file detected, processing upload")

			if c.s3Service == nil {
				util.NewDefaultLogger(ctx).Error("S3Service is nil")
				util.HandleError(ctx, constant.ErrInternalServerError)
				return
			}
			file := photoFile[0]

			const maxFileSize = 5 * 1024 * 1024 // 5MB
			if file.Size > maxFileSize {
				util.HandleError(ctx, constant.ErrValidationFailed)
				return
			}

			if file.Size == 0 {
				util.HandleError(ctx, constant.ErrValidationFailed)
				return
			}

			fileContentType := file.Header.Get("Content-Type")
			allowedTypes := []string{"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp"}
			isValidType := false
			for _, allowedType := range allowedTypes {
				if fileContentType == allowedType {
					isValidType = true
					break
				}
			}
			if !isValidType {
				util.HandleError(ctx, constant.ErrValidationFailed)
				return
			}

			if user.PhotoURL != "" {
				if err := c.s3Service.DeleteFile(ctx.Request.Context(), user.PhotoURL); err != nil {
					util.NewDefaultLogger(ctx).WithField("error", err.Error()).Warn("Failed to delete old photo")
				}
			}

			prefix := "user_photos/" + userID.String()
			photoURL, err := c.s3Service.UploadFile(ctx.Request.Context(), prefix, file)
			if err != nil {
				util.HandleError(ctx, err)
				return
			}

			user.PhotoURL = photoURL
		} else {
			util.NewDefaultLogger(ctx).Info("No photo file found in form.File[\"photo_url\"]")
		}
	} else {
		var rawRequest map[string]interface{}
		if err := ctx.ShouldBindJSON(&rawRequest); err != nil {
			util.HandleError(ctx, err)
			return
		}

		restrictedFields := []string{
			"username", "Username",
			"email", "Email",
			"match_count", "MatchCount",
			"is_email_verified", "IsEmailVerified",
			"verification_token", "VerificationToken",
			"verification_sent_at", "VerificationSentAt",
			"last_login", "LastLogin",
			"created_at", "CreatedAt",
			"updated_at", "UpdatedAt",
			"deleted_at", "DeletedAt",
			"id", "ID",
			"password_hash", "PasswordHash",
		}

		for _, field := range restrictedFields {
			if _, ok := rawRequest[field]; ok {
				util.HandleError(ctx, constant.ErrValidationFailed)
				return
			}
		}

		jsonBytes, err := json.Marshal(rawRequest)
		if err != nil {
			util.HandleError(ctx, err)
			return
		}
		if err := json.Unmarshal(jsonBytes, &updateReq); err != nil {
			util.HandleError(ctx, err)
			return
		}
	}

	if updateReq.FirstName != "" {
		user.FirstName = updateReq.FirstName
	}
	if updateReq.LastName != "" {
		user.LastName = updateReq.LastName
	}
	if updateReq.Phone != "" {
		user.Phone = updateReq.Phone
	}
	if updateReq.Location != "" {
		user.Location = updateReq.Location
	}
	if !updateReq.BirthDate.IsZero() {
		user.BirthDate = updateReq.BirthDate
	}

	resp, err := c.userService.Upsert(ctx, user)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	util.HandleResponse(ctx, &response.BaseResponse{
		Message: "User updated successfully",
		Data:    resp.Data,
	}, nil)
}
