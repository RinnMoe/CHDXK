package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"jcourse/internal/domain/auth"
)

const csrfHeader = "X-CSRF-Token"

func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth.GetApiKeyFromCtx(c.Request.Context()) != nil {
			c.Next()
			return
		}

		s := sessions.Default(c)

		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			token := csrfToken(s)
			c.Header(csrfHeader, token)
			c.Next()
		default:
			expected, ok := s.Get(sessionKeyCSRFToken).(string)
			if !ok || expected == "" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "csrf token missing"})
				return
			}
			actual := c.GetHeader(csrfHeader)
			if subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) != 1 {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "csrf token mismatch"})
				return
			}
			c.Next()
		}
	}
}

func csrfToken(s sessions.Session) string {
	if token, ok := s.Get(sessionKeyCSRFToken).(string); ok && token != "" {
		return token
	}
	token := generateToken()
	s.Set(sessionKeyCSRFToken, token)
	_ = s.Save()
	return token
}

func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
