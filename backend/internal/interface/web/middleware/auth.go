package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"jcourse/internal/domain/auth"
)

func Auth(currentUserSvc *auth.AuthUserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth.GetUserFromCtx(c.Request.Context()) != nil {
			c.Next()
			return
		}

		s := sessions.Default(c)
		userID, ok := sessionInt(s, sessionKeyUserID)
		if !ok || userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		user, err := currentUserSvc.GetUser(c.Request.Context(), userID)
		if err != nil || user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		ctx := auth.WithUser(c.Request.Context(), user)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func OptionalAuth(currentUserSvc *auth.AuthUserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth.GetUserFromCtx(c.Request.Context()) != nil {
			c.Next()
			return
		}

		s := sessions.Default(c)
		userID, ok := sessionInt(s, sessionKeyUserID)
		if !ok || userID == 0 {
			c.Next()
			return
		}

		user, err := currentUserSvc.GetUser(c.Request.Context(), userID)
		if err != nil || user == nil {
			c.Next()
			return
		}

		ctx := auth.WithUser(c.Request.Context(), user)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func Admin() gin.HandlerFunc {
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
