package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"jcourse/internal/domain/auth"
)

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

func TestUserRateLimiterSharesAnonymousUserID(t *testing.T) {
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
	if code := performRateLimitRequest(r, 0); code != http.StatusTooManyRequests {
		t.Fatalf("third anonymous request status = %d, want %d", code, http.StatusTooManyRequests)
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
