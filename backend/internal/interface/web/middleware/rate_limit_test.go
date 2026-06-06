package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"jcourse/internal/domain/auth"
)

func TestGlobalRateLimiterLimitsAcrossAllRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewGlobalRateLimiter(rate.Limit(1), 2)
	r := gin.New()
	r.Use(limiter.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	if code := performAnonymousRateLimitRequest(r, "192.0.2.1"); code != http.StatusOK {
		t.Fatalf("first request status = %d, want %d", code, http.StatusOK)
	}
	if code := performRateLimitRequest(r, 7); code != http.StatusOK {
		t.Fatalf("second request status = %d, want %d", code, http.StatusOK)
	}
	if code := performAnonymousRateLimitRequest(r, "192.0.2.2"); code != http.StatusTooManyRequests {
		t.Fatalf("third request status = %d, want %d", code, http.StatusTooManyRequests)
	}
}

func TestUserRateLimiterLimitsByUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewUserRateLimiter(rate.Limit(1), 2)
	limiter.cleanupInterval = 0
	r := gin.New()
	r.Use(limiter.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	if code := performRateLimitRequest(r, 1); code != http.StatusOK {
		t.Fatalf("first request status = %d, want %d", code, http.StatusOK)
	}
	if code := performRateLimitRequest(r, 1); code != http.StatusOK {
		t.Fatalf("second request status = %d, want %d", code, http.StatusOK)
	}
	if code := performRateLimitRequest(r, 1); code != http.StatusTooManyRequests {
		t.Fatalf("third request status = %d, want %d", code, http.StatusTooManyRequests)
	}
	if code := performRateLimitRequest(r, 2); code != http.StatusOK {
		t.Fatalf("different user status = %d, want %d", code, http.StatusOK)
	}
}

func TestUserRateLimiterSkipsAdmins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewUserRateLimiter(rate.Limit(1), 2)
	limiter.cleanupInterval = 0
	r := gin.New()
	r.Use(limiter.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := range 3 {
		if code := performRateLimitRequestWithUser(r, &auth.User{ID: 1, Role: auth.RoleAdmin}); code != http.StatusOK {
			t.Fatalf("admin request %d status = %d, want %d", i+1, code, http.StatusOK)
		}
	}
	for i := range 3 {
		if code := performRateLimitRequestWithUser(r, &auth.User{ID: 2, Role: auth.RoleSuperAdmin}); code != http.StatusOK {
			t.Fatalf("super admin request %d status = %d, want %d", i+1, code, http.StatusOK)
		}
	}
}

func TestUserRateLimiterSkipsSystemUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewUserRateLimiter(rate.Limit(1), 2)
	limiter.cleanupInterval = 0
	r := gin.New()
	r.Use(limiter.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := range 3 {
		if code := performRateLimitRequestWithUser(r, &auth.User{Role: auth.RoleSystem}); code != http.StatusOK {
			t.Fatalf("system request %d status = %d, want %d", i+1, code, http.StatusOK)
		}
	}
}

func TestAdminStillSubjectToGlobalRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	globalLimiter := NewGlobalRateLimiter(rate.Limit(1), 2)
	userLimiter := NewUserRateLimiter(rate.Limit(100), 100)
	userLimiter.cleanupInterval = 0
	r := gin.New()
	r.Use(globalLimiter.Middleware())
	r.Use(userLimiter.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	admin := &auth.User{ID: 1, Role: auth.RoleAdmin}
	if code := performRateLimitRequestWithUser(r, admin); code != http.StatusOK {
		t.Fatalf("first admin request status = %d, want %d", code, http.StatusOK)
	}
	if code := performRateLimitRequestWithUser(r, admin); code != http.StatusOK {
		t.Fatalf("second admin request status = %d, want %d", code, http.StatusOK)
	}
	if code := performRateLimitRequestWithUser(r, admin); code != http.StatusTooManyRequests {
		t.Fatalf("third admin request status = %d, want %d", code, http.StatusTooManyRequests)
	}
}

func TestSystemUserStillSubjectToGlobalRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	globalLimiter := NewGlobalRateLimiter(rate.Limit(1), 2)
	userLimiter := NewUserRateLimiter(rate.Limit(100), 100)
	userLimiter.cleanupInterval = 0
	r := gin.New()
	r.Use(globalLimiter.Middleware())
	r.Use(userLimiter.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	system := &auth.User{Role: auth.RoleSystem}
	if code := performRateLimitRequestWithUser(r, system); code != http.StatusOK {
		t.Fatalf("first system request status = %d, want %d", code, http.StatusOK)
	}
	if code := performRateLimitRequestWithUser(r, system); code != http.StatusOK {
		t.Fatalf("second system request status = %d, want %d", code, http.StatusOK)
	}
	if code := performRateLimitRequestWithUser(r, system); code != http.StatusTooManyRequests {
		t.Fatalf("third system request status = %d, want %d", code, http.StatusTooManyRequests)
	}
}

func TestUserRateLimiterLimitsByAnonymousIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewUserRateLimiter(rate.Limit(1), 2)
	limiter.cleanupInterval = 0
	r := gin.New()
	r.Use(limiter.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	if code := performAnonymousRateLimitRequest(r, "192.0.2.1"); code != http.StatusOK {
		t.Fatalf("first anonymous request status = %d, want %d", code, http.StatusOK)
	}
	if code := performAnonymousRateLimitRequest(r, "192.0.2.1"); code != http.StatusOK {
		t.Fatalf("second anonymous request status = %d, want %d", code, http.StatusOK)
	}
	if code := performAnonymousRateLimitRequest(r, "192.0.2.1"); code != http.StatusTooManyRequests {
		t.Fatalf("third anonymous request status = %d, want %d", code, http.StatusTooManyRequests)
	}
}

func TestUserRateLimiterSeparatesAnonymousIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewUserRateLimiter(rate.Limit(1), 2)
	limiter.cleanupInterval = 0
	r := gin.New()
	r.Use(limiter.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	if code := performAnonymousRateLimitRequest(r, "192.0.2.1"); code != http.StatusOK {
		t.Fatalf("first anonymous request status = %d, want %d", code, http.StatusOK)
	}
	if code := performAnonymousRateLimitRequest(r, "192.0.2.1"); code != http.StatusOK {
		t.Fatalf("second anonymous request status = %d, want %d", code, http.StatusOK)
	}
	if code := performAnonymousRateLimitRequest(r, "192.0.2.2"); code != http.StatusOK {
		t.Fatalf("different anonymous ip status = %d, want %d", code, http.StatusOK)
	}
}

func TestUserRateLimiterSharesUserAPIKeyWithUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewUserRateLimiter(rate.Limit(1), 2)
	limiter.cleanupInterval = 0
	r := gin.New()
	r.Use(limiter.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	if code := performRateLimitRequest(r, 7); code != http.StatusOK {
		t.Fatalf("first user request status = %d, want %d", code, http.StatusOK)
	}
	if code := performUserAPIKeyRateLimitRequest(r, 1, 7); code != http.StatusOK {
		t.Fatalf("user api key request status = %d, want %d", code, http.StatusOK)
	}
	if code := performRateLimitRequest(r, 7); code != http.StatusTooManyRequests {
		t.Fatalf("shared user bucket status = %d, want %d", code, http.StatusTooManyRequests)
	}
	if code := performUserAPIKeyRateLimitRequest(r, 2, 8); code != http.StatusOK {
		t.Fatalf("different api key user status = %d, want %d", code, http.StatusOK)
	}
}

func TestUserRateLimiterSeparatesSystemAPIKeyFromAnonymousIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewUserRateLimiter(rate.Limit(1), 2)
	limiter.cleanupInterval = 0
	r := gin.New()
	r.Use(limiter.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	if code := performRateLimitRequest(r, 0); code != http.StatusOK {
		t.Fatalf("first anonymous request status = %d, want %d", code, http.StatusOK)
	}
	if code := performRateLimitRequest(r, 0); code != http.StatusOK {
		t.Fatalf("second anonymous request status = %d, want %d", code, http.StatusOK)
	}
	if code := performAPIKeyRateLimitRequest(r, 1); code != http.StatusOK {
		t.Fatalf("system api key status = %d, want %d", code, http.StatusOK)
	}
}

func performRateLimitRequest(r http.Handler, userID int) int {
	var u *auth.User
	if userID > 0 {
		u = &auth.User{ID: userID, Role: auth.RoleUser}
	}
	return performRateLimitRequestWithUser(r, u)
}

func performRateLimitRequestWithUser(r http.Handler, u *auth.User) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if u != nil {
		req = req.WithContext(auth.WithUser(req.Context(), u))
	}
	r.ServeHTTP(w, req)
	return w.Code
}

func performAnonymousRateLimitRequest(r http.Handler, ip string) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = ip + ":1234"
	r.ServeHTTP(w, req)
	return w.Code
}

func performAPIKeyRateLimitRequest(r http.Handler, apiKeyID int64) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(auth.WithApiKey(req.Context(), &auth.ApiKey{ID: apiKeyID, Role: auth.ApiKeyRoleSystem}))
	r.ServeHTTP(w, req)
	return w.Code
}

func performUserAPIKeyRateLimitRequest(r http.Handler, apiKeyID int64, userID int) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(auth.WithApiKey(req.Context(), &auth.ApiKey{ID: apiKeyID, Role: auth.ApiKeyRoleUser, UserID: userID}))
	r.ServeHTTP(w, req)
	return w.Code
}
