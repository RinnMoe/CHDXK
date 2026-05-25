package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"jcourse/internal/domain/auth"
)

const (
	HeaderAuthorization = "Authorization"
	PrefixBearer        = "Bearer "
)

func SystemAPIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(HeaderAuthorization)
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		if !strings.HasPrefix(header, PrefixBearer) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}
		if bearerToken(c) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}

		apiKey := auth.GetApiKeyFromCtx(c.Request.Context())
		if apiKey == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}
		if !apiKey.IsSystem() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "api key role is not allowed"})
			return
		}
		c.Next()
	}
}

func bearerToken(c *gin.Context) string {
	header := c.GetHeader(HeaderAuthorization)
	if header == "" {
		return ""
	}
	if !strings.HasPrefix(header, PrefixBearer) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, PrefixBearer))
}
