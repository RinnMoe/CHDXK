package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"jcourse/pkg/requestid"
)

func TestRequestIDUsesExistingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/", func(c *gin.Context) {
		id, ok := requestid.FromContext(c.Request.Context())
		if !ok || id != "existing-request-id" {
			t.Fatalf("request id from context = %q, %v", id, ok)
		}
		if got := c.GetString(requestid.GinKey); got != "existing-request-id" {
			t.Fatalf("gin request id = %q", got)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestid.HeaderRequestID, "existing-request-id")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(requestid.HeaderRequestID); got != "existing-request-id" {
		t.Fatalf("response request id = %q", got)
	}
}

func TestRequestIDGeneratesMissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/", func(c *gin.Context) {
		id, ok := requestid.FromContext(c.Request.Context())
		if !ok || id == "" {
			t.Fatalf("request id from context = %q, %v", id, ok)
		}
		if got := c.GetHeader(requestid.HeaderRequestID); got != id {
			t.Fatalf("request header request id = %q, want %q", got, id)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(requestid.HeaderRequestID); len(got) != 34 {
		t.Fatalf("generated request id = %q, len = %d", got, len(got))
	}
}
