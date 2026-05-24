package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	gsessions "github.com/gorilla/sessions"

	"jcourse/internal/infrastructure/persistence"
)

const (
	sessionKeyUserID    = "user_id"
	sessionKeyCSRFToken = "csrf_token"
	sessionRedisPrefix  = "jcourse:session:"
)

type SessionConfig struct {
	Secret string `mapstructure:"secret"`
	MaxAge int    `mapstructure:"max_age"` // seconds
	Secure bool   `mapstructure:"secure"`  // HTTPS only
}

var DefaultSessionConfig = SessionConfig{
	MaxAge: 2592000,
	Secure: false,
}

func NewSessionStore(redisConf persistence.RedisConfig, sessionConf SessionConfig) (sessions.Store, error) {
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

func SetSessionUserID(c *gin.Context, userID int) error {
	if err := resetSession(c); err != nil {
		return err
	}
	s := sessions.Default(c)
	s.Set(sessionKeyUserID, userID)
	return s.Save()
}

func ClearSession(c *gin.Context) error {
	s := sessions.Default(c)
	if accessor, ok := s.(gorillaSessionAccessor); ok {
		gs := accessor.Session()
		if gs.Options != nil {
			options := *gs.Options
			options.MaxAge = -1
			gs.Options = &options
		} else {
			gs.Options = &gsessions.Options{Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode}
		}
	} else {
		s.Options(sessions.Options{Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	}
	s.Clear()
	return s.Save()
}

type gorillaSessionAccessor interface {
	Session() *gsessions.Session
}

func resetSession(c *gin.Context) error {
	s := sessions.Default(c)
	if accessor, ok := s.(gorillaSessionAccessor); ok {
		gs := accessor.Session()
		options := gs.Options
		if gs.ID != "" {
			deleteOptions := gsessions.Options{Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode}
			if options != nil {
				deleteOptions = *options
			}
			gs.Options = &deleteOptions
			gs.Options.MaxAge = -1
			if err := gs.Save(c.Request, c.Writer); err != nil {
				return err
			}
		}
		gs.ID = ""
		gs.IsNew = true
		gs.Values = map[interface{}]interface{}{}
		gs.Options = options
		return nil
	}
	s.Clear()
	return s.Save()
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
