package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	redigo "github.com/gomodule/redigo/redis"
	gsessions "github.com/gorilla/sessions"

	"jcourse/internal/domain/auth"
	"jcourse/internal/infrastructure/persistence"
)

const (
	sessionKeyOIDCAccessToken  = "oidc_access_token"
	sessionKeyOIDCIDToken      = "oidc_id_token"
	sessionKeyOIDCRefreshToken = "oidc_refresh_token"
	sessionKeyOIDCExpiresAt    = "oidc_expires_at"
	sessionKeyOIDCDeadline     = "oidc_deadline"
	sessionKeyOIDCState        = "oidc_state"
	sessionKeyOIDCVerifier     = "oidc_verifier"
	sessionKeyOIDCNonce        = "oidc_nonce"
	sessionKeyOIDCReturn       = "oidc_return"
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
	pool := &redigo.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigo.Conn, error) {
			return redigo.Dial("tcp", redisConf.Addr,
				redigo.DialUsername(redisConf.Username), redigo.DialPassword(redisConf.Password), redigo.DialDatabase(redisConf.DB),
				redigo.DialConnectTimeout(time.Second), redigo.DialReadTimeout(time.Second), redigo.DialWriteTimeout(time.Second))
		},
		TestOnBorrow: func(conn redigo.Conn, _ time.Time) error { _, err := conn.Do("PING"); return err },
	}
	store, err := redis.NewStoreWithPool(pool, []byte(sessionConf.Secret))
	if err != nil {
		_ = pool.Close()
		return nil, err
	}
	if err := redis.SetKeyPrefix(store, sessionRedisPrefix); err != nil {
		_ = pool.Close()
		return nil, err
	}
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   sessionConf.MaxAge,
		HttpOnly: true,
		Secure:   sessionConf.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	retention := time.Duration(sessionConf.MaxAge) * time.Second
	if retention <= 0 {
		retention = time.Duration(DefaultSessionConfig.MaxAge) * time.Second
	}
	return &revokedSessionStore{base: store, pool: pool, retention: retention}, nil
}

func SetSessionOIDC(c *gin.Context, accessToken, idToken string) error {
	return SetSessionOIDCCredentials(c, auth.ResHubSession{AccessToken: accessToken, IDToken: idToken})
}

// SetSessionOIDCCredentials also records an absolute local lifetime. Refreshing
// provider credentials must not silently turn a seven-day session into an unlimited one.
func SetSessionOIDCCredentials(c *gin.Context, credentials auth.ResHubSession) error {
	if credentials.AccessToken == "" || credentials.IDToken == "" {
		return fmt.Errorf("invalid oidc session")
	}
	if err := resetSession(c); err != nil {
		return err
	}
	s := sessions.Default(c)
	s.Set(sessionKeyOIDCAccessToken, credentials.AccessToken)
	s.Set(sessionKeyOIDCIDToken, credentials.IDToken)
	if credentials.RefreshToken != "" && !credentials.ExpiresAt.IsZero() {
		s.Set(sessionKeyOIDCRefreshToken, credentials.RefreshToken)
		s.Set(sessionKeyOIDCExpiresAt, credentials.ExpiresAt.Unix())
		if accessor, ok := s.(gorillaSessionAccessor); ok && accessor.Session().Options != nil && accessor.Session().Options.MaxAge > 0 {
			s.Set(sessionKeyOIDCDeadline, time.Now().Add(time.Duration(accessor.Session().Options.MaxAge)*time.Second).Unix())
		}
	}
	return s.Save()
}

func OIDCAccessToken(c *gin.Context) string {
	s := sessions.Default(c)
	token, _ := sessionString(s, sessionKeyOIDCAccessToken)
	return token
}

func SetNativePending(c *gin.Context, ttl time.Duration) error {
	s := sessions.Default(c)
	s.Set("native_pending_until", time.Now().Add(ttl).Unix())
	return s.Save()
}

func NativePendingExpired(c *gin.Context) bool {
	until, ok := sessions.Default(c).Get("native_pending_until").(int64)
	return ok && time.Now().Unix() >= until
}

func OIDCIDToken(c *gin.Context) string {
	s := sessions.Default(c)
	token, _ := sessionString(s, sessionKeyOIDCIDToken)
	return token
}

func SetOIDCState(c *gin.Context, state, verifier, nonce, returnPath string) error {
	s := sessions.Default(c)
	s.Set(sessionKeyOIDCState, state)
	s.Set(sessionKeyOIDCVerifier, verifier)
	s.Set(sessionKeyOIDCNonce, nonce)
	s.Set(sessionKeyOIDCReturn, returnPath)
	return s.Save()
}

func TakeOIDCState(c *gin.Context, expectedState ...string) (state, verifier, nonce, returnPath string, ok bool) {
	s := sessions.Default(c)
	state, stateOK := sessionString(s, sessionKeyOIDCState)
	verifier, verifierOK := sessionString(s, sessionKeyOIDCVerifier)
	nonce, nonceOK := sessionString(s, sessionKeyOIDCNonce)
	returnPath, returnOK := sessionString(s, sessionKeyOIDCReturn)
	if !stateOK || !verifierOK || !nonceOK || !returnOK {
		return "", "", "", "", false
	}
	if len(expectedState) > 0 && state != expectedState[0] {
		return state, verifier, nonce, returnPath, false
	}
	s.Delete(sessionKeyOIDCState)
	s.Delete(sessionKeyOIDCVerifier)
	s.Delete(sessionKeyOIDCNonce)
	s.Delete(sessionKeyOIDCReturn)
	_ = s.Save()
	return state, verifier, nonce, returnPath, true
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

func AccountSessionID(c *gin.Context) int {
	v, _ := sessions.Default(c).Get("account_id").(int)
	return v
}
func AccountSessionVersion(c *gin.Context) int {
	v, _ := sessions.Default(c).Get("account_version").(int)
	return v
}
func SetAccountSession(c *gin.Context, id, version int, rotate bool) error {
	token, idToken := OIDCAccessToken(c), OIDCIDToken(c)
	s := sessions.Default(c)
	refresh, expires, deadline := s.Get(sessionKeyOIDCRefreshToken), s.Get(sessionKeyOIDCExpiresAt), s.Get(sessionKeyOIDCDeadline)
	if rotate {
		if err := resetSession(c); err != nil {
			return err
		}
	}
	s = sessions.Default(c)
	s.Set("account_id", id)
	s.Set("account_version", version)
	if token != "" {
		s.Set(sessionKeyOIDCAccessToken, token)
		s.Set(sessionKeyOIDCIDToken, idToken)
		if refresh != nil {
			s.Set(sessionKeyOIDCRefreshToken, refresh)
		}
		if expires != nil {
			s.Set(sessionKeyOIDCExpiresAt, expires)
		}
		if deadline != nil {
			s.Set(sessionKeyOIDCDeadline, deadline)
		}
	}
	return s.Save()
}
func SetOIDCBindingTarget(c *gin.Context, id, version int) error {
	s := sessions.Default(c)
	s.Set("oidc_binding_id", id)
	s.Set("oidc_binding_version", version)
	return s.Save()
}
func TakeOIDCBindingTarget(c *gin.Context) (int, int) {
	s := sessions.Default(c)
	id, _ := s.Get("oidc_binding_id").(int)
	version, _ := s.Get("oidc_binding_version").(int)
	s.Delete("oidc_binding_id")
	s.Delete("oidc_binding_version")
	_ = s.Save()
	return id, version
}

func SetPasswordSession(c *gin.Context, id, version int) error {
	if err := resetSession(c); err != nil {
		return err
	}
	s := sessions.Default(c)
	s.Set("account_id", id)
	s.Set("account_version", version)
	return s.Save()
}
