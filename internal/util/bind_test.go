package util

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBindJSONOrEmpty_emptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(nil))

	type payload struct {
		X string `json:"x"`
	}
	var dst payload
	if err := BindJSONOrEmpty(c, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.X != "" {
		t.Fatalf("expected zero, got %+v", dst)
	}
}

func TestBindJSONOrEmpty_withJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"x":"a"}`)))

	type payload struct {
		X string `json:"x"`
	}
	var dst payload
	if err := BindJSONOrEmpty(c, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.X != "a" {
		t.Fatalf("got %+v", dst)
	}
}
