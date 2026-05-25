package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
)

type fakeAccessTracker struct {
	userID   int
	apiKeyID int
}

func (t *fakeAccessTracker) RecordUserAccess(_ context.Context, userID int, _ time.Time) error {
	t.userID = userID
	return nil
}

func (t *fakeAccessTracker) RecordApiKeyAccess(_ context.Context, apiKeyID int, _ time.Time) error {
	t.apiKeyID = apiKeyID
	return nil
}

func (t *fakeAccessTracker) Flush(context.Context) (auth.AccessFlushResult, error) {
	return auth.AccessFlushResult{}, nil
}

func TestSystemAPIKeyAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validKey := "test-api-key-123"
	repo := &auth.MockApiKeyRepository{Key: &auth.ApiKey{ID: 1, Key: validKey, Role: auth.ApiKeyRoleSystem, UserID: 0}}
	svc := auth.NewApiKeyService(repo, auth.DefaultApiKeyConfig)

	handler := middleware.SystemAPIKeyAuth()

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantBody   string
		apiKey     *auth.ApiKey
		wantAccess int
	}{
		{
			name:       "missing header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "missing authorization header",
			apiKey:     &auth.ApiKey{ID: 1, Key: validKey, Role: auth.ApiKeyRoleSystem, UserID: 0},
		},
		{
			name:       "wrong format",
			authHeader: "Basic abc123",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid authorization format",
			apiKey:     &auth.ApiKey{ID: 1, Key: validKey, Role: auth.ApiKeyRoleSystem, UserID: 0},
		},
		{
			name:       "invalid key",
			authHeader: "Bearer wrong-key",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid api key",
			apiKey:     &auth.ApiKey{ID: 1, Key: validKey, Role: auth.ApiKeyRoleSystem, UserID: 0},
		},
		{
			name:       "valid key",
			authHeader: "Bearer " + validKey,
			wantStatus: http.StatusOK,
			wantBody:   "ok",
			apiKey:     &auth.ApiKey{ID: 1, Key: validKey, Role: auth.ApiKeyRoleSystem, UserID: 0},
			wantAccess: 1,
		},
		{
			name:       "user key forbidden",
			authHeader: "Bearer user-key",
			wantStatus: http.StatusForbidden,
			wantBody:   "api key role is not allowed",
			apiKey:     &auth.ApiKey{ID: 2, Key: "user-key", Role: auth.ApiKeyRoleUser, UserID: 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := &fakeAccessTracker{}
			repo.Key = tt.apiKey

			r := gin.New()
			r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
			r.Use(middleware.ResolveCurrentUser(application.NewAuthResolutionService(
				auth.NewCurrentUserService(&systemAuthUserRepo{}),
				svc,
				tracker,
			)))
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
			if tracker.apiKeyID != tt.wantAccess {
				t.Fatalf("recorded api key access id = %d, want %d", tracker.apiKeyID, tt.wantAccess)
			}
		})
	}
}

func TestSystemAPIKeyAuth_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := auth.NewApiKeyService(&auth.MockApiKeyRepository{
		OnFindByKey: func(context.Context, string) (*auth.ApiKey, error) {
			return nil, errors.New("db down")
		},
	}, auth.DefaultApiKeyConfig)
	handler := middleware.SystemAPIKeyAuth()

	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
	r.Use(middleware.ResolveCurrentUser(application.NewAuthResolutionService(
		auth.NewCurrentUserService(&systemAuthUserRepo{}),
		svc,
		nil,
	)))
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

func TestSystemAPIKeyAuthReusesResolvedAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validKey := "test-api-key-123"
	repo := &auth.MockApiKeyRepository{Key: &auth.ApiKey{ID: 1, Key: validKey, Role: auth.ApiKeyRoleSystem, UserID: 0}}
	svc := auth.NewApiKeyService(repo, auth.DefaultApiKeyConfig)
	tracker := &fakeAccessTracker{}

	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
	r.Use(middleware.ResolveCurrentUser(application.NewAuthResolutionService(auth.NewCurrentUserService(&systemAuthUserRepo{}), svc, tracker)))
	r.GET("/test", middleware.SystemAPIKeyAuth(), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validKey)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if repo.FindByKeyCalls != 1 {
		t.Fatalf("FindByKey calls = %d, want 1", repo.FindByKeyCalls)
	}
	if tracker.apiKeyID != 1 {
		t.Fatalf("recorded api key access id = %d, want 1", tracker.apiKeyID)
	}
}

type systemAuthUserRepo struct{}

func (systemAuthUserRepo) Update(context.Context, *auth.User) error { return nil }

func (systemAuthUserRepo) FindByID(context.Context, int) (*auth.User, error) { return nil, nil }

func (systemAuthUserRepo) FindByRole(context.Context, string) ([]auth.User, error) {
	return nil, nil
}
