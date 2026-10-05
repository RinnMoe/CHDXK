package middleware

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/pkg/apperr"
)

func ResolveCurrentUser(authResolution *application.AuthResolutionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// OIDC start and callback must remain reachable when the browser carries
		// an expired access token. Otherwise an old token can prevent the user
		// from starting a fresh authorization flow or consuming its code.
		if isPublicOIDCRoute(c.Request.URL.Path) {
			c.Next()
			return
		}
		if auth.GetUserFromCtx(c.Request.Context()) != nil || auth.GetApiKeyFromCtx(c.Request.Context()) != nil {
			c.Next()
			return
		}

		resolved, err := authResolution.Resolve(c.Request.Context(), bearerToken(c), OIDCAccessToken(c))
		if err != nil {
			if errors.Is(err, auth.ErrResHubSession) {
				_ = ClearSession(c)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			if errors.Is(err, auth.ErrUserSuspended) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			if errors.Is(err, auth.ErrInvalidApiKeyRole) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			if errors.Is(err, auth.ErrResHubUnavailable) {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		ctx := c.Request.Context()
		if resolved.ApiKey != nil {
			ctx = auth.WithApiKey(ctx, resolved.ApiKey)
		}
		if resolved.User != nil {
			ctx = auth.WithUser(ctx, resolved.User)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func isPublicOIDCRoute(path string) bool {
	return path == "/api/auth/oidc/start" || path == "/api/auth/oidc/callback" || path == "/api/site-copy" || path == "/api/site-pages" || path == "/api/site.webmanifest" || path == "/api/system-settings"
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth.GetUserFromCtx(c.Request.Context()) == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := auth.GetUserFromCtx(c.Request.Context())
		if u == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if !u.IsAdmin() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := auth.GetUserFromCtx(c.Request.Context())
		if u == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if !u.IsSuperAdmin() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": apperr.ErrRequireSuperAdmin.Error()})
			return
		}
		c.Next()
	}
}

func RequireSelfOrAdmin(userIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := auth.GetUserFromCtx(c.Request.Context())
		if u == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, err := strconv.Atoi(c.Param(userIDParam))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}

		if u.ID != userID && !u.IsAdmin() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
