package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	filesession "github.com/gin-contrib/sessions/filesystem"
	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
)

func TestRequireAuthReusesResolvedSessionUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := auth.NewMockUserRepository(map[int]*auth.User{7: {ID: 7, Role: auth.RoleUser}})
	currentUserSvc := auth.NewCurrentUserService(repo)
	apiKeyRepo := &auth.MockApiKeyRepository{}
	apiKeySvc := auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, auth.DefaultApiKeyConfig)
	authResolution := application.NewAuthResolutionService(currentUserSvc, apiKeySvc, nil)
	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", filesession.NewStore(t.TempDir(), []byte("test-secret"))))
	r.Use(ResolveCurrentUser(authResolution))
	r.GET("/login-session", func(c *gin.Context) {
		if err := SetSessionUserID(c, 7); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	r.GET("/protected", RequireAuth(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	cookieValue := captureSessionCookie(t, r)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Cookie", cookieValue)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if repo.FindByIDCalls != 1 {
		t.Fatalf("FindByID calls = %d, want 1", repo.FindByIDCalls)
	}
}

func TestResolveCurrentUserWithUserAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	credential := testAuthMiddlewareCredential(1)
	userRepo := auth.NewMockUserRepository(map[int]*auth.User{7: {ID: 7, Role: auth.RoleUser}})
	apiKeyRepo := &auth.MockApiKeyRepository{Key: &auth.ApiKey{ID: credential.KeyID, SecretHash: credential.SecretHash, Role: auth.ApiKeyRoleUser, UserID: 7}}
	tracker := &authMiddlewareAccessTracker{}
	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
	r.Use(ResolveCurrentUser(application.NewAuthResolutionService(
		auth.NewCurrentUserService(userRepo),
		auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, auth.DefaultApiKeyConfig),
		tracker,
	)))
	r.GET("/protected", RequireAuth(), func(c *gin.Context) {
		u := auth.GetUserFromCtx(c.Request.Context())
		if u == nil || u.ID != 7 {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(HeaderAuthorization, PrefixBearer+credential.Key())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if apiKeyRepo.GetByIDCalls != 1 {
		t.Fatalf("GetByID calls = %d, want 1", apiKeyRepo.GetByIDCalls)
	}
	if userRepo.FindByIDCalls != 1 {
		t.Fatalf("FindByID calls = %d, want 1", userRepo.FindByIDCalls)
	}
	if tracker.apiKeyID != 1 {
		t.Fatalf("recorded api key access id = %d, want 1", tracker.apiKeyID)
	}
	if tracker.userID != 7 {
		t.Fatalf("recorded user access id = %d, want 7", tracker.userID)
	}
}

func TestResolveCurrentUserRejectsSuspendedSessionUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now()
	suspendedAt := now.Add(-time.Hour)
	suspendTill := now.Add(time.Hour)
	repo := auth.NewMockUserRepository(map[int]*auth.User{
		7: {
			ID:          7,
			Role:        auth.RoleUser,
			SuspendedAt: &suspendedAt,
			SuspendTill: &suspendTill,
		},
	})
	currentUserSvc := auth.NewCurrentUserService(repo)
	apiKeyRepo := &auth.MockApiKeyRepository{}
	apiKeySvc := auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, auth.DefaultApiKeyConfig)
	authResolution := application.NewAuthResolutionService(currentUserSvc, apiKeySvc, nil)
	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", filesession.NewStore(t.TempDir(), []byte("test-secret"))))
	r.Use(ResolveCurrentUser(authResolution))
	r.GET("/login-session", func(c *gin.Context) {
		if err := SetSessionUserID(c, 7); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	r.GET("/protected", RequireAuth(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	cookieValue := captureSessionCookie(t, r)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Cookie", cookieValue)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestCSRFMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
	r.Use(CSRF())
	r.GET("/csrf", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	r.POST("/protected", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("post without csrf status = %d, want %d", w.Code, http.StatusForbidden)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/csrf", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("csrf status = %d, want %d", w.Code, http.StatusNoContent)
	}
	token := w.Header().Get(csrfHeader)
	if token == "" {
		t.Fatal("expected csrf header")
	}
	csrfCookie := firstCookie(t, w.Result())

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set("Cookie", csrfCookie)
	req.Header.Set(csrfHeader, token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("post with csrf status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestSetSessionUserIDClearsExistingSessionState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
	r.GET("/csrf", func(c *gin.Context) {
		s := sessions.Default(c)
		s.Set(sessionKeyCSRFToken, "attacker-token")
		if err := s.Save(); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	r.GET("/login-session", func(c *gin.Context) {
		if err := SetSessionUserID(c, 7); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	r.POST("/protected", CSRF(), func(c *gin.Context) { c.Status(http.StatusNoContent) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/csrf", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("csrf setup status = %d, want %d", w.Code, http.StatusNoContent)
	}
	oldCookie := firstCookie(t, w.Result())

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/login-session", nil)
	req.Header.Set("Cookie", oldCookie)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("login status = %d, want %d", w.Code, http.StatusNoContent)
	}
	newCookie := firstCookie(t, w.Result())

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set("Cookie", newCookie)
	req.Header.Set(csrfHeader, "attacker-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("post with old csrf status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestClearSessionExpiresCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
	r.GET("/login-session", func(c *gin.Context) {
		if err := SetSessionUserID(c, 7); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	r.POST("/logout", func(c *gin.Context) {
		if err := ClearSession(c); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})

	cookieValue := captureSessionCookie(t, r)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Cookie", cookieValue)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d", w.Code, http.StatusNoContent)
	}
	res := w.Result()
	defer res.Body.Close()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected expired session cookie")
	}
	if cookies[0].MaxAge >= 0 {
		t.Fatalf("logout cookie MaxAge = %d, want negative", cookies[0].MaxAge)
	}
}

func captureSessionCookie(t *testing.T, r http.Handler) string {
	t.Helper()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login-session", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("session setup status = %d, want %d", w.Code, http.StatusNoContent)
	}
	res := w.Result()
	defer res.Body.Close()
	return firstCookie(t, res)
}

func firstCookie(t *testing.T, res *http.Response) string {
	t.Helper()

	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}
	return cookies[0].String()
}

type authMiddlewareAccessTracker struct {
	userID   int
	apiKeyID int64
}

func (t *authMiddlewareAccessTracker) RecordUserAccess(_ context.Context, userID int, _ time.Time) error {
	t.userID = userID
	return nil
}

func (t *authMiddlewareAccessTracker) RecordApiKeyAccess(_ context.Context, apiKeyID int64, _ time.Time) error {
	t.apiKeyID = apiKeyID
	return nil
}

func (t *authMiddlewareAccessTracker) Flush(context.Context) (auth.AccessFlushResult, error) {
	return auth.AccessFlushResult{}, nil
}

func testAuthMiddlewareCredential(keyID int64) auth.ApiKeyCredential {
	credential := auth.ApiKeyCredential{KeyID: keyID, Secret: []byte("1234567890abcdef")}
	credential.SecretHash = credential.HashSecret()
	return credential
}
