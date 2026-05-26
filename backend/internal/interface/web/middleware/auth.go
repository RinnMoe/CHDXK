package middleware

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
)

func ResolveCurrentUser(authResolution *application.AuthResolutionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth.GetUserFromCtx(c.Request.Context()) != nil || auth.GetApiKeyFromCtx(c.Request.Context()) != nil {
			c.Next()
			return
		}

		resolved, err := authResolution.Resolve(c.Request.Context(), bearerToken(c), sessionUserID(c))
		if err != nil {
			if errors.Is(err, auth.ErrUserSuspended) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
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

func sessionUserID(c *gin.Context) int {
	s := sessions.Default(c)
	userID, ok := sessionInt(s, sessionKeyUserID)
	if !ok || userID == 0 {
		return 0
	}
	return userID
}
