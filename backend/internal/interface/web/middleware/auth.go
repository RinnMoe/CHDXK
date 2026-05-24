package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"

	"jcourse/config"
	"jcourse/internal/domain/auth"
)

const (
	sessionKeyUserID    = "user_id"
	sessionKeyCSRFToken = "csrf_token"
	sessionRedisPrefix  = "jcourse:session:"
)

func NewSessionStore(redisConf config.RedisConfig, sessionConf config.SessionConfig) (sessions.Store, error) {
	store, err := redis.NewStoreWithDB(10, "tcp", redisConf.Addr, redisConf.Username, redisConf.Password, fmt.Sprintf("%d", redisConf.DB), []byte(sessionConf.Secret))
	if err != nil {
		return nil, err
	}
	if err := redis.SetKeyPrefix(store, sessionRedisPrefix); err != nil {
		return nil, err
	}
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   sessionConf.MaxAge,
		HttpOnly: true,
		Secure:   sessionConf.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	return store, nil
}

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

func SetSessionUserID(c *gin.Context, userID int) error {
	s := sessions.Default(c)
	s.Set(sessionKeyUserID, userID)
	return s.Save()
}

func ClearSession(c *gin.Context) error {
	s := sessions.Default(c)
	s.Clear()
	return s.Save()
}

func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)

		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			token := s.Get(sessionKeyCSRFToken)
			if token == nil {
				token = generateToken()
				s.Set(sessionKeyCSRFToken, token)
				_ = s.Save()
			}
			c.Header("X-CSRF-Token", token.(string))
			c.Next()
		default:
			expected := s.Get(sessionKeyCSRFToken)
			if expected == nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "csrf token missing"})
				return
			}
			if c.GetHeader("X-CSRF-Token") != expected.(string) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "csrf token mismatch"})
				return
			}
			c.Next()
		}
	}
}

func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func sessionInt(s sessions.Session, key string) (int, bool) {
	v := s.Get(key)
	if v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}
