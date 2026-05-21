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

func APIKeyAuth(apiKeySvc *auth.ApiKeyService) gin.HandlerFunc {
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

		key := strings.TrimPrefix(header, PrefixBearer)
		valid, err := apiKeySvc.ValidateKey(c.Request.Context(), key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}

		c.Next()
	}
}
