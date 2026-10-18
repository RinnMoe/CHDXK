package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"jcourse/config"
	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/internal/infrastructure/reshub"
	"jcourse/internal/interface/web/middleware"
)

// Consume only synthetic, loopback credentials supplied by the provider agent.
func TestOIDCRefreshAgainstLocalProviderAndRedis(t *testing.T) {
	path, addr := os.Getenv("JCOURSE_TEST_OIDC_FIXTURE"), os.Getenv("JCOURSE_TEST_REDIS_ADDR")
	if path == "" || addr == "" {
		t.Skip("local provider fixture and isolated Redis must both be configured")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("cannot read provider fixture")
	}
	var fixture struct {
		Issuer             string `json:"issuer"`
		ClientID           string `json:"client_id"`
		ClientSecret       string `json:"client_secret"`
		SecondRefreshToken string `json:"second_refresh_token"`
		Subject            string `json:"subject"`
	}
	if json.Unmarshal(data, &fixture) != nil || fixture.SecondRefreshToken == "" {
		t.Fatal("invalid provider fixture")
	}
	issuer, err := url.Parse(fixture.Issuer)
	if err != nil || (issuer.Hostname() != "127.0.0.1" && issuer.Hostname() != "localhost") {
		t.Fatal("provider fixture must use loopback")
	}
	provider, err := reshub.NewClient(config.ResHubConfig{Issuer: fixture.Issuer, ClientID: fixture.ClientID, ClientSecret: fixture.ClientSecret})
	if err != nil {
		t.Fatal("cannot configure provider")
	}
	redisClient := redis.NewClient(&redis.Options{Addr: addr})
	defer redisClient.Close()
	store, err := middleware.NewSessionStore(persistence.RedisConfig{Addr: addr}, middleware.SessionConfig{Secret: "synthetic-local-test-session-secret", MaxAge: 604800})
	if err != nil {
		t.Fatal("cannot configure test session store")
	}
	resolution := application.NewAuthResolutionService(nil, nil, provider, nil)
	resolution.SetAccounts(&auth.MockAccountRepository{User: &auth.User{ID: 7, AuthVersion: 1, Role: auth.RoleUser, Subject: fixture.Subject}, Match: true})
	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", store), middleware.RefreshOIDCSession(reshub.NewRefreshCoordinator(redisClient, provider)), middleware.ResolveCurrentUser(resolution))
	r.GET("/seed", func(c *gin.Context) {
		if err := middleware.SetSessionOIDCCredentials(c, auth.ResHubSession{AccessToken: "expired-placeholder", IDToken: "synthetic-original-id", RefreshToken: fixture.SecondRefreshToken, ExpiresAt: time.Now().Add(-time.Minute)}); err != nil {
			t.Fatal("cannot seed session")
		}
		if err := middleware.SetAccountSession(c, 7, 1, false); err != nil {
			t.Fatal("cannot seed account")
		}
	})
	writes := 0
	r.POST("/write", middleware.RequireAuth(), func(c *gin.Context) {
		writes++
		if middleware.OIDCAccessToken(c) == "expired-placeholder" || auth.GetUserFromCtx(c.Request.Context()).ID != 7 {
			t.Error("refreshed session did not resolve account")
		}
		c.Status(204)
	})
	r.POST("/logout", func(c *gin.Context) {
		if err := middleware.ClearSession(c); err != nil {
			t.Error("logout failed")
		}
		c.Status(204)
	})
	seed := httptest.NewRecorder()
	r.ServeHTTP(seed, httptest.NewRequest("GET", "/seed", nil))
	cookies := seed.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("missing seed cookie")
	}
	req := httptest.NewRequest("POST", "/write", nil)
	req.AddCookie(cookies[len(cookies)-1])
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 204 || writes != 1 {
		t.Fatalf("real provider+Redis middleware status=%d writes=%d", w.Code, writes)
	}
	renewed := w.Result().Cookies()
	if len(renewed) == 0 {
		t.Fatal("missing renewed session cookie")
	}
	cookie := renewed[len(renewed)-1]
	if strings.Contains(cookie.Value, fixture.SecondRefreshToken) {
		t.Fatal("provider credential leaked in cookie")
	}
	logout := httptest.NewRequest("POST", "/logout", nil)
	logout.AddCookie(cookie)
	out := httptest.NewRecorder()
	r.ServeHTTP(out, logout)
	if out.Code != 204 {
		t.Fatal("logout rejected")
	}
	stale := httptest.NewRequest("POST", "/write", nil)
	stale.AddCookie(cookie)
	denied := httptest.NewRecorder()
	r.ServeHTTP(denied, stale)
	if denied.Code != http.StatusUnauthorized || writes != 1 {
		t.Fatal("logout left authenticated business access")
	}
	// The isolated Redis is owned and removed by the local verification run.
}
