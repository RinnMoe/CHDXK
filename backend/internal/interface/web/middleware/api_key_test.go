package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
)

type fakeApiKeyRepo struct {
	key     string
	apiKey  *auth.ApiKey
	err     error
	touched bool
}

func (r *fakeApiKeyRepo) FindByKey(_ context.Context, key string) (*auth.ApiKey, error) {
	if r.err != nil {
		return nil, r.err
	}
	if key == r.key && r.apiKey != nil {
		return r.apiKey, nil
	}
	return nil, nil
}

func (r *fakeApiKeyRepo) ListByUser(_ context.Context, _ int) ([]auth.ApiKey, error) {
	return nil, nil
}

func (r *fakeApiKeyRepo) CountByUser(_ context.Context, _ int) (int, error) {
	return 0, nil
}

func (r *fakeApiKeyRepo) Create(_ context.Context, _ *auth.ApiKey) error {
	return nil
}

func (r *fakeApiKeyRepo) DeleteByUser(_ context.Context, _ int, _ int) (bool, error) {
	return false, nil
}

func (r *fakeApiKeyRepo) TouchLastUsed(_ context.Context, _ int, _ time.Time) error {
	r.touched = true
	return nil
}

func TestSystemAPIKeyAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validKey := "test-api-key-123"
	repo := &fakeApiKeyRepo{key: validKey, apiKey: &auth.ApiKey{ID: 1, Key: validKey, Role: auth.ApiKeyRoleSystem, UserID: 0}}
	svc := auth.NewApiKeyService(repo, auth.DefaultApiKeyConfig())

	handler := middleware.SystemAPIKeyAuth(svc)

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantBody   string
		apiKey     *auth.ApiKey
		wantTouch  bool
		keyValue   string
	}{
		{
			name:       "missing header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "missing authorization header",
			apiKey:     &auth.ApiKey{ID: 1, Key: validKey, Role: auth.ApiKeyRoleSystem, UserID: 0},
			keyValue:   validKey,
		},
		{
			name:       "wrong format",
			authHeader: "Basic abc123",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid authorization format",
			apiKey:     &auth.ApiKey{ID: 1, Key: validKey, Role: auth.ApiKeyRoleSystem, UserID: 0},
			keyValue:   validKey,
		},
		{
			name:       "invalid key",
			authHeader: "Bearer wrong-key",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid api key",
			apiKey:     &auth.ApiKey{ID: 1, Key: validKey, Role: auth.ApiKeyRoleSystem, UserID: 0},
			keyValue:   validKey,
		},
		{
			name:       "valid key",
			authHeader: "Bearer " + validKey,
			wantStatus: http.StatusOK,
			wantBody:   "ok",
			apiKey:     &auth.ApiKey{ID: 1, Key: validKey, Role: auth.ApiKeyRoleSystem, UserID: 0},
			wantTouch:  true,
			keyValue:   validKey,
		},
		{
			name:       "user key forbidden",
			authHeader: "Bearer user-key",
			wantStatus: http.StatusForbidden,
			wantBody:   "api key role is not allowed",
			apiKey:     &auth.ApiKey{ID: 2, Key: "user-key", Role: auth.ApiKeyRoleUser, UserID: 7},
			keyValue:   "user-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.touched = false
			repo.key = tt.keyValue
			repo.apiKey = tt.apiKey

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
			if repo.touched != tt.wantTouch {
				t.Fatalf("touched = %v, want %v", repo.touched, tt.wantTouch)
			}
		})
	}
}

func TestSystemAPIKeyAuth_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := auth.NewApiKeyService(&fakeApiKeyRepo{err: errors.New("db down")}, auth.DefaultApiKeyConfig())
	handler := middleware.SystemAPIKeyAuth(svc)

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
