package upload

import (
	"net/http"
	"strings"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/contract/repository"
	"github.com/api-monolith-template/internal/service/upload"
	"github.com/api-monolith-template/internal/util"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	s3svc          *upload.S3UploaderService
	userRepository repository.UserContract
}

func NewUploadController(svc *upload.S3UploaderService) *Controller {
	return &Controller{s3svc: svc}
}

func (c *Controller) WithUserRepository(repo repository.UserContract) *Controller {
	c.userRepository = repo
	return c
}

func (c *Controller) BuildingPhoto(ctx *gin.Context) {
	prefix := "building_photos"
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	url, err := c.s3svc.UploadFile(ctx.Request.Context(), prefix, file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"url": url})
}

func (c *Controller) BuildingPhotoDelete(ctx *gin.Context) {
	var body struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.s3svc.DeleteFile(ctx.Request.Context(), body.URL); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) FieldPhoto(ctx *gin.Context) {
	prefix := "field_photos"
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	url, err := c.s3svc.UploadFile(ctx.Request.Context(), prefix, file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"url": url})
}

func (c *Controller) FieldPhotoDelete(ctx *gin.Context) {
	var body struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.s3svc.DeleteFile(ctx.Request.Context(), body.URL); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) UserPhoto(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	userRaw, exists := ctx.Get(string(constant.UserID))
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	userIDStr, ok := userRaw.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	_, err = uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id format"})
		return
	}

	user, err := c.userRepository.GetByIdentifier(ctx.Request.Context(), userIDStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	logger := util.NewDefaultLogger(ctx.Request.Context())
	oldPhotoURL := user.PhotoURL

	// Check if user already has a photo and validate file type/size
	if file.Size == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file cannot be empty"})
		return
	}

	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "only image files are allowed"})
		return
	}

	// Delete old photo if exists
	if oldPhotoURL != "" {
		logger.WithFields(logrus.Fields{
			"userID":      user.ID,
			"oldPhotoURL": oldPhotoURL,
		}).Info("Deleting old user photo")

		err = c.s3svc.DeleteFile(ctx.Request.Context(), oldPhotoURL)
		if err != nil {
			logger.WithField("error", err.Error()).Warn("Failed to delete old photo, but continuing with upload")
		} else {
			logger.WithFields(logrus.Fields{
				"userID":      user.ID,
				"oldPhotoURL": oldPhotoURL,
			}).Info("Successfully deleted old user photo")
		}
	}

	// Upload new photo
	prefix := "user_photos/" + userIDStr
	url, err := c.s3svc.UploadFile(ctx.Request.Context(), prefix, file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update user's PhotoURL in database
	user.PhotoURL = url
	logger.WithFields(logrus.Fields{
		"userID":   user.ID,
		"photoURL": url,
	}).Info("Updating user PhotoURL in database")

	err = c.userRepository.Update(ctx.Request.Context(), user)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to update user PhotoURL in database")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user photo"})
		return
	}

	logger.WithFields(logrus.Fields{
		"userID":   user.ID,
		"photoURL": url,
	}).Info("Successfully updated user PhotoURL in database")

	ctx.JSON(http.StatusCreated, gin.H{"url": url})
}

func (c *Controller) UserPhotoDelete(ctx *gin.Context) {
	userRaw, exists := ctx.Get(string(constant.UserID))
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	userIDStr, ok := userRaw.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	// Get current user data
	user, err := c.userRepository.GetByIdentifier(ctx.Request.Context(), userIDStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	logger := util.NewDefaultLogger(ctx.Request.Context())

	if user.PhotoURL == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user has no photo to delete"})
		return
	}

	// Delete photo from S3
	logger.WithFields(logrus.Fields{
		"userID":   user.ID,
		"photoURL": user.PhotoURL,
	}).Info("Deleting user photo from S3")

	err = c.s3svc.DeleteFile(ctx.Request.Context(), user.PhotoURL)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to delete photo from S3")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete photo"})
		return
	}

	user.PhotoURL = ""
	logger.WithFields(logrus.Fields{
		"userID": user.ID,
	}).Info("Clearing user PhotoURL in database")

	err = c.userRepository.Update(ctx.Request.Context(), user)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to clear user PhotoURL in database")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	logger.WithFields(logrus.Fields{
		"userID": user.ID,
	}).Info("Successfully deleted user photo")

	ctx.JSON(http.StatusOK, gin.H{"message": "Photo deleted successfully"})
}
