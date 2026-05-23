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

func SystemAPIKeyAuth(apiKeySvc *auth.ApiKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey, ok := authenticateAPIKey(c, apiKeySvc)
		if !ok {
			return
		}
		if !apiKey.IsSystem() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "api key role is not allowed"})
			return
		}
		if err := apiKeySvc.MarkKeyUsed(c.Request.Context(), apiKey.ID); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		ctx := auth.WithApiKey(c.Request.Context(), apiKey)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func UserAPIKeyAuth(apiKeySvc *auth.ApiKeyService, authSvc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey, ok := authenticateAPIKey(c, apiKeySvc)
		if !ok {
			return
		}
		if !apiKey.IsUser() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "api key role is not allowed"})
			return
		}

		user, err := authSvc.GetUser(c.Request.Context(), apiKey.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if err := apiKeySvc.MarkKeyUsed(c.Request.Context(), apiKey.ID); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		ctx := auth.WithApiKey(c.Request.Context(), apiKey)
		ctx = auth.WithUser(ctx, user)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func authenticateAPIKey(c *gin.Context, apiKeySvc *auth.ApiKeyService) (*auth.ApiKey, bool) {
	header := c.GetHeader(HeaderAuthorization)
	if header == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return nil, false
	}

	if !strings.HasPrefix(header, PrefixBearer) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		return nil, false
	}

	key := strings.TrimSpace(strings.TrimPrefix(header, PrefixBearer))
	if key == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return nil, false
	}

	apiKey, err := apiKeySvc.ValidateKey(c.Request.Context(), key)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return nil, false
	}
	if apiKey == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return nil, false
	}
	return apiKey, true
}
