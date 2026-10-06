package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
)

func TestResolveCurrentUserWithOIDCSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := &middlewareOIDCClient{identity: &auth.ResHubIdentity{Subject: "user:1"}}
	projection := &middlewareIdentityProjection{user: &auth.User{ID: 7, Role: auth.RoleUser, Subject: "user:1"}}
	tracker := &middlewareAccessTracker{}
	apiKeyRepo := &auth.MockApiKeyRepository{}
	authResolution := application.NewAuthResolutionService(
		auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, auth.DefaultApiKeyConfig),
		tracker,
		client,
		projection,
	)

	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
	authResolution.SetAccounts(&auth.MockAccountRepository{User: &auth.User{ID: 7, Role: auth.RoleUser, AuthVersion: 1}, Match: true})
	r.Use(ResolveCurrentUser(authResolution))
	r.GET("/login-session", func(c *gin.Context) {
		if err := SetSessionOIDC(c, "access-token", "id-token"); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		_ = SetAccountSession(c, 7, 1, false)
		c.Status(http.StatusNoContent)
	})
	r.GET("/protected", RequireAuth(), func(c *gin.Context) {
		u := auth.GetUserFromCtx(c.Request.Context())
		if u == nil || u.ID != 7 {
			c.Status(http.StatusInternalServerError)
			return
		}
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
	if client.userinfoToken != "access-token" {
		t.Fatalf("OIDC calls = client %#v projection %#v", client, projection)
	}
	if tracker.userID != 7 {
		t.Fatalf("recorded user access id = %d, want 7", tracker.userID)
	}
}

func TestResolveCurrentUserClearsInvalidOIDCSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := &middlewareOIDCClient{err: auth.ErrResHubSession}
	projection := &middlewareIdentityProjection{}
	apiKeyRepo := &auth.MockApiKeyRepository{}
	authResolution := application.NewAuthResolutionService(
		auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, auth.DefaultApiKeyConfig),
		nil,
		client,
		projection,
	)

	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
	authResolution.SetAccounts(&auth.MockAccountRepository{User: &auth.User{ID: 7, Role: auth.RoleUser, AuthVersion: 1}, Match: true})
	r.Use(ResolveCurrentUser(authResolution))
	r.GET("/login-session", func(c *gin.Context) {
		if err := SetSessionOIDC(c, "expired-access", "id-token"); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		_ = SetAccountSession(c, 7, 1, false)
		c.Status(http.StatusNoContent)
	})
	r.GET("/protected", RequireAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })

	cookieValue := captureSessionCookie(t, r)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Cookie", cookieValue)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if cookies := w.Result().Cookies(); len(cookies) == 0 || cookies[0].MaxAge >= 0 {
		t.Fatalf("expected expired session cookie, got %+v", cookies)
	}
}

func TestResolveCurrentUserLeavesOIDCStartPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	apiKeyRepo := &auth.MockApiKeyRepository{}
	authResolution := application.NewAuthResolutionService(
		auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, auth.DefaultApiKeyConfig),
		nil,
		&middlewareOIDCClient{err: auth.ErrResHubUnavailable},
		nil,
	)
	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
	authResolution.SetAccounts(&auth.MockAccountRepository{User: &auth.User{ID: 7, Role: auth.RoleUser, AuthVersion: 1}, Match: true})
	r.Use(ResolveCurrentUser(authResolution))
	r.GET("/api/auth/oidc/start", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/start", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestRequireSelfOrAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		current    *auth.User
		path       string
		wantStatus int
	}{
		{name: "unauthorized", current: nil, path: "/user/7/review", wantStatus: http.StatusUnauthorized},
		{name: "invalid user id", current: &auth.User{ID: 7, Role: auth.RoleUser}, path: "/user/not-a-number/review", wantStatus: http.StatusBadRequest},
		{name: "forbidden", current: &auth.User{ID: 7, Role: auth.RoleUser}, path: "/user/8/review", wantStatus: http.StatusForbidden},
		{name: "self", current: &auth.User{ID: 7, Role: auth.RoleUser}, path: "/user/7/review", wantStatus: http.StatusNoContent},
		{name: "admin", current: &auth.User{ID: 1, Role: auth.RoleAdmin}, path: "/user/7/review", wantStatus: http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tt.current != nil {
					c.Request = c.Request.WithContext(auth.WithUser(c.Request.Context(), tt.current))
				}
				c.Next()
			})
			r.GET("/user/:userID/review", RequireSelfOrAdmin("userID"), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestRequireSuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		current    *auth.User
		wantStatus int
	}{
		{name: "unauthorized", wantStatus: http.StatusUnauthorized},
		{name: "user forbidden", current: &auth.User{ID: 7, Role: auth.RoleUser}, wantStatus: http.StatusForbidden},
		{name: "admin forbidden", current: &auth.User{ID: 2, Role: auth.RoleAdmin}, wantStatus: http.StatusForbidden},
		{name: "super admin", current: &auth.User{ID: 1, Role: auth.RoleSuperAdmin}, wantStatus: http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tt.current != nil {
					c.Request = c.Request.WithContext(auth.WithUser(c.Request.Context(), tt.current))
				}
				c.Next()
			})
			r.PUT("/admin/user/:userID/admin", RequireSuperAdmin(), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/admin/user/7/admin", nil)
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
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

func TestSetSessionOIDCResetsExistingSessionState(t *testing.T) {
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
		if err := SetSessionOIDC(c, "access-token", "id-token"); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		_ = SetAccountSession(c, 7, 1, false)
		c.Status(http.StatusNoContent)
	})
	r.POST("/protected", CSRF(), func(c *gin.Context) { c.Status(http.StatusNoContent) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/csrf", nil)
	r.ServeHTTP(w, req)
	oldCookie := firstCookie(t, w.Result())

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/login-session", nil)
	req.Header.Set("Cookie", oldCookie)
	r.ServeHTTP(w, req)
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
		if err := SetSessionOIDC(c, "access-token", "id-token"); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		_ = SetAccountSession(c, 7, 1, false)
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
	if cookies := w.Result().Cookies(); len(cookies) == 0 || cookies[0].MaxAge >= 0 {
		t.Fatalf("expected expired session cookie, got %+v", cookies)
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
	return firstCookie(t, w.Result())
}

func firstCookie(t *testing.T, res *http.Response) string {
	t.Helper()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}
	return cookies[len(cookies)-1].String()
}

type middlewareOIDCClient struct {
	identity      *auth.ResHubIdentity
	err           error
	userinfoToken string
}

func (f *middlewareOIDCClient) AuthorizationURL(context.Context, string, string, string) (string, error) {
	return "", errors.New("not used")
}

func (f *middlewareOIDCClient) Exchange(context.Context, string, string, string, string) (auth.ResHubSession, error) {
	return auth.ResHubSession{}, errors.New("not used")
}

func (f *middlewareOIDCClient) UserInfo(_ context.Context, accessToken string) (*auth.ResHubIdentity, error) {
	f.userinfoToken = accessToken
	return f.identity, f.err
}

func (f *middlewareOIDCClient) ClientCredentialsToken(context.Context, []string) (string, error) {
	return "", errors.New("not used")
}

func (f *middlewareOIDCClient) Logout(context.Context, string, string) error {
	return nil
}

type middlewareIdentityProjection struct {
	identity *auth.ResHubIdentity
	user     *auth.User
}

func (p *middlewareIdentityProjection) Upsert(_ context.Context, identity *auth.ResHubIdentity) (*auth.User, error) {
	p.identity = identity
	return p.user, nil
}

func (p *middlewareIdentityProjection) SubjectByUserID(context.Context, int) (string, error) {
	return "user:1", nil
}

type middlewareAccessTracker struct {
	userID int
}

func (t *middlewareAccessTracker) RecordUserAccess(_ context.Context, userID int, _ time.Time) error {
	t.userID = userID
	return nil
}

func (t *middlewareAccessTracker) RecordApiKeyAccess(context.Context, int64, time.Time) error {
	return nil
}

func (t *middlewareAccessTracker) Flush(context.Context) (auth.AccessFlushResult, error) {
	return auth.AccessFlushResult{}, nil
}
