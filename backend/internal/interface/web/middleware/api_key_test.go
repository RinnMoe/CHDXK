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
	apiKeyID int64
}

func (t *fakeAccessTracker) RecordUserAccess(_ context.Context, userID int, _ time.Time) error {
	t.userID = userID
	return nil
}

func (t *fakeAccessTracker) RecordApiKeyAccess(_ context.Context, apiKeyID int64, _ time.Time) error {
	t.apiKeyID = apiKeyID
	return nil
}

func (t *fakeAccessTracker) Flush(context.Context) (auth.AccessFlushResult, error) {
	return auth.AccessFlushResult{}, nil
}

func TestSystemAPIKeyAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	systemCredential := testMiddlewareCredential(1)
	userCredential := testMiddlewareCredential(2)
	repo := &auth.MockApiKeyRepository{Key: &auth.ApiKey{ID: systemCredential.KeyID, SecretHash: systemCredential.SecretHash, Role: auth.ApiKeyRoleSystem, UserID: 0}}
	svc := auth.NewApiKeyService(repo, repo, auth.DefaultApiKeyConfig)

	handler := middleware.SystemAPIKeyAuth()

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantBody   string
		apiKey     *auth.ApiKey
		wantAccess int64
	}{
		{
			name:       "missing header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "missing authorization header",
			apiKey:     &auth.ApiKey{ID: systemCredential.KeyID, SecretHash: systemCredential.SecretHash, Role: auth.ApiKeyRoleSystem, UserID: 0},
		},
		{
			name:       "wrong format",
			authHeader: "Basic abc123",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid authorization format",
			apiKey:     &auth.ApiKey{ID: systemCredential.KeyID, SecretHash: systemCredential.SecretHash, Role: auth.ApiKeyRoleSystem, UserID: 0},
		},
		{
			name:       "invalid key",
			authHeader: "Bearer wrong-key",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid api key",
			apiKey:     &auth.ApiKey{ID: systemCredential.KeyID, SecretHash: systemCredential.SecretHash, Role: auth.ApiKeyRoleSystem, UserID: 0},
		},
		{
			name:       "valid key",
			authHeader: "Bearer " + systemCredential.Key(),
			wantStatus: http.StatusOK,
			wantBody:   "ok",
			apiKey:     &auth.ApiKey{ID: systemCredential.KeyID, SecretHash: systemCredential.SecretHash, Role: auth.ApiKeyRoleSystem, UserID: 0},
			wantAccess: 1,
		},
		{
			name:       "user key forbidden",
			authHeader: "Bearer " + userCredential.Key(),
			wantStatus: http.StatusForbidden,
			wantBody:   "api key role is not allowed",
			apiKey:     &auth.ApiKey{ID: userCredential.KeyID, SecretHash: userCredential.SecretHash, Role: auth.ApiKeyRoleUser, UserID: 7},
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

	credential := testMiddlewareCredential(1)
	repo := &auth.MockApiKeyRepository{
		OnGetByID: func(context.Context, int64) (*auth.ApiKey, error) {
			return nil, errors.New("db down")
		},
	}
	svc := auth.NewApiKeyService(repo, repo, auth.DefaultApiKeyConfig)
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
	req.Header.Set("Authorization", "Bearer "+credential.Key())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestSystemAPIKeyAuthReusesResolvedAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	credential := testMiddlewareCredential(1)
	repo := &auth.MockApiKeyRepository{Key: &auth.ApiKey{ID: credential.KeyID, SecretHash: credential.SecretHash, Role: auth.ApiKeyRoleSystem, UserID: 0}}
	svc := auth.NewApiKeyService(repo, repo, auth.DefaultApiKeyConfig)
	tracker := &fakeAccessTracker{}

	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
	r.Use(middleware.ResolveCurrentUser(application.NewAuthResolutionService(auth.NewCurrentUserService(&systemAuthUserRepo{}), svc, tracker)))
	r.GET("/test", middleware.SystemAPIKeyAuth(), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+credential.Key())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if repo.GetByIDCalls != 1 {
		t.Fatalf("GetByID calls = %d, want 1", repo.GetByIDCalls)
	}
	if tracker.apiKeyID != 1 {
		t.Fatalf("recorded api key access id = %d, want 1", tracker.apiKeyID)
	}
}

type systemAuthUserRepo struct{}

func (systemAuthUserRepo) Update(context.Context, *auth.User) error { return nil }

func (systemAuthUserRepo) FindByID(context.Context, int) (*auth.User, error) { return nil, nil }

func (systemAuthUserRepo) FindAdmin(context.Context) ([]auth.User, error) {
	return nil, nil
}

func testMiddlewareCredential(keyID int64) auth.ApiKeyCredential {
	credential := auth.ApiKeyCredential{KeyID: keyID, Secret: []byte("1234567890abcdef")}
	credential.SecretHash = credential.HashSecret()
	return credential
}
