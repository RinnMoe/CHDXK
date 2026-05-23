package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"jcourse/internal/domain/auth"
)

const anonymousUserID = 0

type userLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type UserRateLimiter struct {
	rate            rate.Limit
	burst           int
	cleanupInterval time.Duration
	idleTTL         time.Duration
	now             func() time.Time

	mu       sync.Mutex
	limiters map[int]*userLimiter
}

func NewUserRateLimiter(rps rate.Limit, burst int) *UserRateLimiter {
	return &UserRateLimiter{
		rate:            rps,
		burst:           burst,
		cleanupInterval: time.Minute,
		idleTTL:         10 * time.Minute,
		now:             time.Now,
		limiters:        make(map[int]*userLimiter),
	}
}

func UserIDRateLimit() gin.HandlerFunc {
	return NewUserRateLimiter(rate.Limit(10), 10).Middleware()
}

func (r *UserRateLimiter) Middleware() gin.HandlerFunc {
	if r.cleanupInterval > 0 && r.idleTTL > 0 {
		go r.cleanupLoop()
	}

	return func(c *gin.Context) {
		limiter := r.limiterFor(userIDFromContext(c))
		if !limiter.Allow() {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func (r *UserRateLimiter) limiterFor(userID int) *rate.Limiter {
	now := r.now()
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.limiters[userID]
	if !ok {
		entry = &userLimiter{limiter: rate.NewLimiter(r.rate, r.burst)}
		r.limiters[userID] = entry
	}
	entry.lastSeen = now
	return entry.limiter
}

func (r *UserRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(r.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		r.cleanup()
	}
}

func (r *UserRateLimiter) cleanup() {
	cutoff := r.now().Add(-r.idleTTL)
	r.mu.Lock()
	defer r.mu.Unlock()

	for userID, entry := range r.limiters {
		if entry.lastSeen.Before(cutoff) {
			delete(r.limiters, userID)
		}
	}
}

func userIDFromContext(c *gin.Context) int {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil || u.ID <= 0 {
		return anonymousUserID
	}
	return u.ID
}
