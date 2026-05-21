package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
)

type fakeApiKeyRepo struct {
	key   string
	err   error
	found bool
}

func (r *fakeApiKeyRepo) FindByKey(_ context.Context, key string) (*auth.ApiKey, error) {
	if r.err != nil {
		return nil, r.err
	}
	if key == r.key && r.found {
		return &auth.ApiKey{ID: 1, Key: key}, nil
	}
	return nil, nil
}

func TestAPIKeyAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validKey := "test-api-key-123"
	svc := auth.NewApiKeyService(&fakeApiKeyRepo{key: validKey, found: true})

	handler := middleware.APIKeyAuth(svc)

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "missing header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "missing authorization header",
		},
		{
			name:       "wrong format",
			authHeader: "Basic abc123",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid authorization format",
		},
		{
			name:       "invalid key",
			authHeader: "Bearer wrong-key",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid api key",
		},
		{
			name:       "valid key",
			authHeader: "Bearer " + validKey,
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(handler)
			r.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Fatalf("body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestAPIKeyAuth_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := auth.NewApiKeyService(&fakeApiKeyRepo{err: errors.New("db down")})
	handler := middleware.APIKeyAuth(svc)

	r := gin.New()
	r.Use(handler)
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer some-key")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
