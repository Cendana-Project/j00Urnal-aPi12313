package warmup

import (
	"net/http"
	"time"

	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/service/storage"
	"github.com/api-monolith-template/internal/util"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	storageSvc *storage.Service
}

func NewController(storageSvc *storage.Service) *Controller {
	return &Controller{
		storageSvc: storageSvc,
	}
}

func (c *Controller) Ping(ctx *gin.Context) {
	resp := response.BaseResponse{
		StatusCode: http.StatusOK,
		Message:    "pong",
		MessageDetail: response.MessageDetail{
			TitleEng: "PONG",
			DescEng:  "Service is alive and ready",
			TitleIdn: "PONG",
			DescIdn:  "Layanan aktif dan siap",
		},
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"status":    "ok",
		},
	}
	util.HandleResponse(ctx, &resp, nil)
}

func (c *Controller) Health(ctx *gin.Context) {
	storageStatus := "up"
	storageErr := ""

	if c.storageSvc != nil {
		if err := c.storageSvc.HealthCheck(ctx); err != nil {
			storageStatus = "down"
			storageErr = err.Error()
		}
	} else {
		storageStatus = "not_configured"
	}

	statusCode := http.StatusOK
	message := "success"
	if storageStatus == "down" {
		statusCode = http.StatusServiceUnavailable
		message = "service unavailable"
	}

	resp := response.BaseResponse{
		StatusCode: statusCode,
		Message:    message,
		MessageDetail: response.MessageDetail{
			TitleEng: "Health Check",
			DescEng:  "Service and bucket health status",
			TitleIdn: "Health Check",
			DescIdn:  "Status kesehatan layanan dan bucket",
		},
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"status":    "up",
			"bucket": map[string]interface{}{
				"status": storageStatus,
				"error":  storageErr,
			},
		},
	}
	util.HandleResponse(ctx, &resp, nil)
}
