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
	sessionKeyUserID           = "user_id"
	sessionKeyAuthHash         = "auth_hash"
	sessionKeyIntegrationToken = "integration_token"
	sessionKeySSOState         = "sso_state"
	sessionKeySSOVerifier      = "sso_verifier"
	sessionKeySSOReturn        = "sso_return"
	sessionKeyCSRFToken        = "csrf_token"
	sessionRedisPrefix         = "jcourse:session:"
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
	return SetSessionUser(c, userID, "")
}

func SetSessionUser(c *gin.Context, userID int, authHash string) error {
	if err := resetSession(c); err != nil {
		return err
	}
	s := sessions.Default(c)
	s.Set(sessionKeyUserID, userID)
	if authHash != "" {
		s.Set(sessionKeyAuthHash, authHash)
	}
	return s.Save()
}

func SetSessionIntegration(c *gin.Context, userID int, token string) error {
	if userID <= 0 || token == "" {
		return fmt.Errorf("invalid integration session")
	}
	if err := resetSession(c); err != nil {
		return err
	}
	s := sessions.Default(c)
	s.Set(sessionKeyUserID, userID)
	s.Set(sessionKeyIntegrationToken, token)
	return s.Save()
}

func IntegrationToken(c *gin.Context) string {
	s := sessions.Default(c)
	token, _ := sessionString(s, sessionKeyIntegrationToken)
	return token
}

func SetSSOState(c *gin.Context, state, verifier, returnPath string) error {
	s := sessions.Default(c)
	s.Set(sessionKeySSOState, state)
	s.Set(sessionKeySSOVerifier, verifier)
	s.Set(sessionKeySSOReturn, returnPath)
	return s.Save()
}

func TakeSSOState(c *gin.Context, expectedState ...string) (state, verifier, returnPath string, ok bool) {
	s := sessions.Default(c)
	state, stateOK := sessionString(s, sessionKeySSOState)
	verifier, verifierOK := sessionString(s, sessionKeySSOVerifier)
	returnPath, returnOK := sessionString(s, sessionKeySSOReturn)
	if !stateOK || !verifierOK || !returnOK {
		return "", "", "", false
	}
	if len(expectedState) > 0 && state != expectedState[0] {
		return state, verifier, returnPath, false
	}
	s.Delete(sessionKeySSOState)
	s.Delete(sessionKeySSOVerifier)
	s.Delete(sessionKeySSOReturn)
	_ = s.Save()
	return state, verifier, returnPath, true
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

func sessionString(s sessions.Session, key string) (string, bool) {
	v := s.Get(key)
	if v == nil {
		return "", false
	}
	value, ok := v.(string)
	return value, ok
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
		gs.Values = map[any]any{}
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
