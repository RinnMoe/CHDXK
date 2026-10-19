package middleware_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"jcourse/internal/domain/auth"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/internal/infrastructure/reshub"
	"jcourse/internal/interface/web/middleware"
)

func TestOIDCStateWritePreservesRefreshAfterResultCacheExpires(t *testing.T) {
	addr := os.Getenv("JCOURSE_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("JCOURSE_TEST_REDIS_ADDR is not set")
	}
	gin.SetMode(gin.TestMode)
	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()
	store, err := middleware.NewSessionStore(persistence.RedisConfig{Addr: addr}, middleware.SessionConfig{Secret: "isolated-state-refresh-test-secret", MaxAge: 3600})
	if err != nil {
		t.Fatal("create isolated session store")
	}
	oldToken := "test-state-grant-" + time.Now().Format("150405.000000000")
	digest := sha256.Sum256([]byte(oldToken))
	key := "jcourse:oidc-refresh:" + hex.EncodeToString(digest[:])
	var sessionID string
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = client.Del(ctx, key, key+":lock", key+":consumed", "jcourse:session:"+sessionID).Err()
	}()
	var calls atomic.Int32
	refresher := reshub.NewRefreshCoordinator(client, &auth.MockResHubSessionRefresher{OnRefresh: func(context.Context, string) (auth.ResHubSession, error) {
		calls.Add(1)
		return auth.ResHubSession{AccessToken: "new-access", RefreshToken: "new-refresh", ExpiresAt: time.Now().Add(time.Hour)}, nil
	}})
	router := gin.New()
	router.Use(sessions.Sessions("state_refresh_test", store), middleware.RefreshOIDCSession(refresher))
	router.GET("/seed", func(c *gin.Context) {
		if err := middleware.SetSessionOIDCCredentials(c, auth.ResHubSession{AccessToken: "expired-access", IDToken: "original-id", RefreshToken: oldToken, ExpiresAt: time.Now().Add(-time.Minute)}); err != nil {
			t.Fatal("seed session")
		}
		sessionID = sessions.Default(c).ID()
	})
	loaded, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	router.GET("/api/auth/oidc/start", func(c *gin.Context) {
		_ = middleware.OIDCAccessToken(c)
		close(loaded)
		<-release
		if err := middleware.SetOIDCState(c, "new-state", "verifier", "nonce", "/"); err != nil {
			t.Error("save OIDC state")
		}
	})
	router.GET("/protected", func(c *gin.Context) { c.Status(204) })
	seed := httptest.NewRecorder()
	router.ServeHTTP(seed, httptest.NewRequest("GET", "/seed", nil))
	cookies := seed.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("seed did not create cookie")
	}
	cookie := cookies[len(cookies)-1]
	stateRequest := httptest.NewRequest("GET", "/api/auth/oidc/start", nil)
	stateRequest.AddCookie(cookie)
	go func() { defer close(done); router.ServeHTTP(httptest.NewRecorder(), stateRequest) }()
	select {
	case <-loaded:
	case <-time.After(3 * time.Second):
		t.Fatal("state request did not load")
	}
	request := httptest.NewRequest("GET", "/protected", nil)
	request.AddCookie(cookie)
	renewed := httptest.NewRecorder()
	router.ServeHTTP(renewed, request)
	close(release)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("state write did not finish")
	}
	if renewed.Code != 204 {
		t.Fatal("initial refresh failed")
	}
	// Model natural expiry of only this grant's two-minute result cache.
	if err := client.Del(context.Background(), key).Err(); err != nil {
		t.Fatal("expire result cache")
	}
	later := httptest.NewRequest("GET", "/protected", nil)
	later.AddCookie(cookie)
	result := httptest.NewRecorder()
	router.ServeHTTP(result, later)
	if result.Code != 204 || calls.Load() != 1 {
		t.Fatalf("business status=%d provider calls=%d: state save restored a spent refresh grant", result.Code, calls.Load())
	}
}
