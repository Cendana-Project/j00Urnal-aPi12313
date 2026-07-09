package util

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/api-monolith-template/internal/constant"
)

// BindAndValidate membaca JSON secara ketat (menolak field tak dikenal),
// lalu menjalankan validator milik Gin yang membaca tag `binding`.
func BindAndValidate(c *gin.Context, dst any) error {
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields() // strict: unknown fields -> error

	if err := dec.Decode(dst); err != nil {
		Infof(c.Request.Context(), "BindAndValidate failed: %v", err)
		return constant.ErrValidationError
	}
	return validateStructWithGinEngine(dst)
}

// BindJSONOrEmpty parses JSON like BindAndValidate, but an empty or whitespace-only body leaves dst unchanged (zero value).
func BindJSONOrEmpty(c *gin.Context, dst any) error {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		Infof(c.Request.Context(), "BindJSONOrEmpty read body: %v", err)
		return constant.ErrValidationError
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return validateStructWithGinEngine(dst)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		Infof(c.Request.Context(), "BindJSONOrEmpty decode: %v", err)
		return constant.ErrValidationError
	}
	return validateStructWithGinEngine(dst)
}

// validateStructWithGinEngine validates a struct using Gin's built-in validator
// which reads `binding` tags (not `validate` tags).
func validateStructWithGinEngine(dst any) error {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return ValidateStruct(dst)
	}
	return v.Struct(dst)
}

// Opsional: helper cepat untuk mengakhiri request dengan error validasi standar
func AbortValidation(c *gin.Context) {
	res := constant.ErrValidationError.ToResponse()
	c.AbortWithStatusJSON(res.StatusCode, res)
}
