package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"jcourse/internal/domain/auth"
)

// RefreshOIDCSession runs before identity resolution, including account choice
// and binding routes. It never retries a business request or bypasses userinfo.
func RefreshOIDCSession(refresher auth.ResHubSessionRefresher) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if bearerToken(c) != "" || (isPublicOIDCRoute(path) && path != "/api/auth/accounts" && path != "/api/auth/accounts/select" && !(path == "/api/auth/oidc/start" && c.Query("purpose") == "bind")) {
			c.Next()
			return
		}
		s := sessions.Default(c)
		refresh, _ := sessionString(s, sessionKeyOIDCRefreshToken)
		expires, _ := s.Get(sessionKeyOIDCExpiresAt).(int64)
		if refresh == "" || expires == 0 {
			c.Next()
			return
		}
		deadline, _ := s.Get(sessionKeyOIDCDeadline).(int64)
		if deadline <= time.Now().Unix() || NativePendingExpired(c) {
			_ = ClearSession(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if time.Now().Add(30*time.Second).Unix() < expires {
			c.Next()
			return
		}
		if refresher == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": auth.ErrResHubUnavailable.Error()})
			return
		}
		credentials, err := refresher.Refresh(c.Request.Context(), refresh)
		if err != nil {
			if errors.Is(err, auth.ErrResHubSession) || errors.Is(err, auth.ErrResHubRefreshUncertain) {
				_ = ClearSession(c)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			} else {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": auth.ErrResHubUnavailable.Error()})
			}
			return
		}
		if credentials.AccessToken == "" || credentials.RefreshToken == "" || !credentials.ExpiresAt.After(time.Now()) {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": auth.ErrResHubUnavailable.Error()})
			return
		}
		s.Set(sessionKeyOIDCAccessToken, credentials.AccessToken)
		s.Set(sessionKeyOIDCRefreshToken, credentials.RefreshToken)
		s.Set(sessionKeyOIDCExpiresAt, credentials.ExpiresAt.Unix())
		if accessor, ok := s.(gorillaSessionAccessor); ok && accessor.Session().Options != nil {
			options := *accessor.Session().Options
			options.MaxAge = int(deadline - time.Now().Unix())
			if options.MaxAge <= 0 {
				_ = ClearSession(c)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			accessor.Session().Options = &options
		}
		if err := s.Save(); err != nil {
			if errors.Is(err, auth.ErrResHubSession) {
				// The session was concurrently revoked/rotated. Do not emit a
				// cookie that would overwrite a newer account's browser cookie.
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": auth.ErrResHubUnavailable.Error()})
			return
		}
		c.Next()
	}
}
