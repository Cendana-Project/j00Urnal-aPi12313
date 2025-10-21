package util

import (
	"strings"
	"time"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func HandleError(ctx *gin.Context, err error) {
	switch cErr := err.(type) {
	case response.CustomError:
		resp := cErr.ToResponse()
		resp.TraceID = GetTraceID(ctx)
		resp.Timestamp = time.Now().UTC()
		ctx.JSON(resp.StatusCode, resp)
	case validator.ValidationErrors:
		validationErr := constant.ErrValidationError.ToResponse()
		validationErr.TraceID = GetTraceID(ctx)
		validationErr.Timestamp = time.Now().UTC()
		ctx.JSON(validationErr.StatusCode, validationErr)
	default:
		errStr := err.Error()
		if strings.Contains(errStr, "json") || strings.Contains(errStr, "unmarshal") {
			jsonErr := constant.ErrValidationError.ToResponse()
			detail := constant.GetMessageDetail(constant.MsgValidationError)
			if detail == (response.MessageDetail{}) {
				detail = response.MessageDetail{
					TitleEng: "Validation Error",
					TitleIdn: "format JSON tidak valid",
				}
			}
			detail.TitleEng = coalesce(detail.TitleEng, "Validation Error")
			detail.TitleIdn = "Format JSON tidak valid"

			jsonErr.MessageDetail = detail
			jsonErr.TraceID = GetTraceID(ctx)
			jsonErr.Timestamp = time.Now().UTC()
			ctx.JSON(jsonErr.StatusCode, jsonErr)
			return
		}

		if strings.Contains(errStr, "parsing time") {
			dateErr := constant.ErrInvalidDateFormat.ToResponse()
			dateErr.MessageDetail = response.MessageDetail{
				TitleEng: "Invalid Date Format",
				TitleIdn: "Gunakan format RFC3339 (contoh: 1997-12-22T00:00:00Z)",
			}
			dateErr.TraceID = GetTraceID(ctx)
			dateErr.Timestamp = time.Now().UTC()
			ctx.JSON(dateErr.StatusCode, dateErr)
			return
		}

		internalServerErr := constant.ErrInternalServerError.ToResponse()
		if internalServerErr.MessageDetail == (response.MessageDetail{}) {
			internalServerErr.MessageDetail = constant.GetMessageDetail(constant.MsgInternalServerError)
		}
		internalServerErr.TraceID = GetTraceID(ctx)
		internalServerErr.Timestamp = time.Now().UTC()
		ctx.JSON(internalServerErr.StatusCode, internalServerErr)
	}
}

// coalesce returns b if a is empty
func coalesce(a, b string) string {
	if a == "" {
		return b
	}
	return a
}
