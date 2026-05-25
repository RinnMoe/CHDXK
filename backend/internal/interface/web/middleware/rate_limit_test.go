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
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if userID > 0 {
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: userID}))
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

func performAPIKeyRateLimitRequest(r http.Handler, apiKeyID int) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(auth.WithApiKey(req.Context(), &auth.ApiKey{ID: apiKeyID, Role: auth.ApiKeyRoleSystem}))
	r.ServeHTTP(w, req)
	return w.Code
}

func performUserAPIKeyRateLimitRequest(r http.Handler, apiKeyID int, userID int) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(auth.WithApiKey(req.Context(), &auth.ApiKey{ID: apiKeyID, Role: auth.ApiKeyRoleUser, UserID: userID}))
	r.ServeHTTP(w, req)
	return w.Code
}
