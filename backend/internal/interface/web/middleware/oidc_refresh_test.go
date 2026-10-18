package middleware

import (
	"context"
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

func TestOIDCRefreshPreservesSessionAndRevalidatesIdentity(t *testing.T) {
	for _, rotate := range []bool{false, true} {
		t.Run(map[bool]string{false: "browser", true: "native-account-selection"}[rotate], func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			provider := &middlewareOIDCClient{identity: &auth.ResHubIdentity{Subject: "user:7"}}
			resolution := application.NewAuthResolutionService(nil, nil, provider, nil)
			resolution.SetAccounts(&auth.MockAccountRepository{User: &auth.User{ID: 7, AuthVersion: 2, Role: auth.RoleUser}, Match: true})
			refreshes, writes := 0, 0
			refresher := &auth.MockResHubSessionRefresher{OnRefresh: func(_ context.Context, token string) (auth.ResHubSession, error) {
				refreshes++
				if token != "refresh-old" {
					t.Error("refresh credential lost during account rotation")
				}
				return auth.ResHubSession{AccessToken: "access-new", RefreshToken: "refresh-new", ExpiresAt: time.Now().Add(time.Hour)}, nil
			}}
			r := gin.New()
			store := cookie.NewStore([]byte("test-refresh-secret"))
			store.Options(sessions.Options{Path: "/", MaxAge: 604800, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
			r.Use(sessions.Sessions("jcourse_session", store), RefreshOIDCSession(refresher), ResolveCurrentUser(resolution))
			deadline := time.Now().Add(2 * time.Hour).Unix()
			r.GET("/seed", func(c *gin.Context) {
				if err := SetSessionOIDCCredentials(c, auth.ResHubSession{AccessToken: "access-old", IDToken: "id-original", RefreshToken: "refresh-old", ExpiresAt: time.Now().Add(-time.Minute)}); err != nil {
					t.Fatal(err)
				}
				s := sessions.Default(c)
				s.Set(sessionKeyOIDCDeadline, deadline)
				if err := SetAccountSession(c, 7, 2, rotate); err != nil {
					t.Fatal(err)
				}
				s.Set(sessionKeyCSRFToken, "csrf-original")
				if err := s.Save(); err != nil {
					t.Fatal(err)
				}
			})
			r.POST("/write", RequireAuth(), func(c *gin.Context) {
				writes++
				s := sessions.Default(c)
				if OIDCAccessToken(c) != "access-new" || OIDCIDToken(c) != "id-original" || s.Get(sessionKeyCSRFToken) != "csrf-original" || s.Get(sessionKeyOIDCDeadline) != deadline {
					t.Error("refresh replaced unrelated session data")
				}
				c.Status(204)
			})
			seed := httptest.NewRecorder()
			r.ServeHTTP(seed, httptest.NewRequest("GET", "/seed", nil))
			request := httptest.NewRequest("POST", "/write", nil)
			cookies := seed.Result().Cookies()
			request.AddCookie(cookies[len(cookies)-1])
			response := httptest.NewRecorder()
			r.ServeHTTP(response, request)
			if response.Code != 204 || refreshes != 1 || writes != 1 || provider.userinfoToken != "access-new" {
				t.Fatalf("status=%d refreshes=%d writes=%d; renewed identity not used", response.Code, refreshes, writes)
			}
			updated := response.Result().Cookies()
			if len(updated) == 0 || updated[0].MaxAge > 7200 || updated[0].MaxAge <= 0 || !updated[0].Secure || !updated[0].HttpOnly {
				t.Fatal("cookie flags or absolute lifetime lost")
			}
			provider.err = auth.ErrResHubSession
			revoked := httptest.NewRequest("POST", "/write", nil)
			revoked.AddCookie(updated[len(updated)-1])
			rejected := httptest.NewRecorder()
			r.ServeHTTP(rejected, revoked)
			if rejected.Code != 401 || writes != 1 || refreshes != 1 {
				t.Fatal("renewal bypassed later provider revocation or replayed a write")
			}
		})
	}
}

func TestOIDCRefreshFailuresAndCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name     string
		refresh  bool
		expiry   time.Duration
		deadline time.Duration
		err      error
		path     string
		status   int
		cleared  bool
		calls    int
	}{
		{"legacy", false, -time.Hour, time.Hour, nil, "/protected", 204, false, 0},
		{"fresh", true, time.Hour, 2 * time.Hour, nil, "/protected", 204, false, 0},
		{"outage", true, -time.Hour, time.Hour, auth.ErrResHubUnavailable, "/protected", 503, false, 1},
		{"revoked", true, -time.Hour, time.Hour, auth.ErrResHubSession, "/protected", 401, true, 1},
		{"uncertain", true, -time.Hour, time.Hour, auth.ErrResHubRefreshUncertain, "/protected", 401, true, 1},
		{"absolute-limit", true, time.Hour, -time.Minute, nil, "/protected", 401, true, 0},
		{"login-recovery", true, -time.Hour, -time.Minute, nil, "/api/auth/login", 204, false, 0},
		{"logout", true, -time.Hour, -time.Minute, nil, "/api/auth/logout", 204, false, 0},
		{"pending-accounts", true, -time.Hour, time.Hour, auth.ErrResHubUnavailable, "/api/auth/accounts", 503, false, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			r := gin.New()
			r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-failure-secret"))), RefreshOIDCSession(&auth.MockResHubSessionRefresher{OnRefresh: func(context.Context, string) (auth.ResHubSession, error) {
				calls++
				return auth.ResHubSession{}, tc.err
			}}))
			r.GET("/seed", func(c *gin.Context) {
				_ = SetSessionOIDC(c, "access", "id")
				if tc.refresh {
					s := sessions.Default(c)
					s.Set(sessionKeyOIDCRefreshToken, "refresh")
					s.Set(sessionKeyOIDCExpiresAt, time.Now().Add(tc.expiry).Unix())
					s.Set(sessionKeyOIDCDeadline, time.Now().Add(tc.deadline).Unix())
					_ = s.Save()
				}
			})
			r.GET(tc.path, func(c *gin.Context) { c.Status(204) })
			seed := httptest.NewRecorder()
			r.ServeHTTP(seed, httptest.NewRequest("GET", "/seed", nil))
			req := httptest.NewRequest("GET", tc.path, nil)
			cookies := seed.Result().Cookies()
			req.AddCookie(cookies[len(cookies)-1])
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			cleared := false
			for _, c := range w.Result().Cookies() {
				if c.MaxAge < 0 {
					cleared = true
				}
			}
			if w.Code != tc.status || calls != tc.calls || cleared != tc.cleared {
				t.Fatalf("status=%d calls=%d cleared=%v", w.Code, calls, cleared)
			}
		})
	}
}
