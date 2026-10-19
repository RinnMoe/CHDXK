package middleware

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gomodule/redigo/redis"
	gsessions "github.com/gorilla/sessions"
	"jcourse/internal/domain/auth"
)

// revokedSessionStore prevents a request loaded before logout or session-ID
// rotation from recreating the deleted session when its refresh finishes.
type revokedSessionStore struct {
	base      sessions.Store
	pool      *redis.Pool
	retention time.Duration
}

type sessionCookieSink struct {
	http.ResponseWriter
	header http.Header
}

func (w sessionCookieSink) Header() http.Header { return w.header }

func (s *revokedSessionStore) Options(options sessions.Options) { s.base.Options(options) }

func (s *revokedSessionStore) Get(r *http.Request, name string) (*gsessions.Session, error) {
	return gsessions.GetRegistry(r).Get(s, name)
}

func (s *revokedSessionStore) New(r *http.Request, name string) (*gsessions.Session, error) {
	loaded, err := s.base.New(r, name)
	if loaded == nil {
		return nil, err
	}
	// Gorilla keeps the store private; construct a session bound to this gate so
	// direct Session.Save calls made by resetSession also pass through it.
	wrapped := gsessions.NewSession(s, name)
	wrapped.ID, wrapped.Values, wrapped.Options, wrapped.IsNew = loaded.ID, loaded.Values, loaded.Options, loaded.IsNew
	return wrapped, err
}

func (s *revokedSessionStore) Save(r *http.Request, w http.ResponseWriter, session *gsessions.Session) error {
	if session.ID == "" {
		return s.base.Save(r, w, session)
	}
	digest := sha256.Sum256([]byte(session.ID))
	key := "jcourse:session-revocation:" + hex.EncodeToString(digest[:])
	var owner [16]byte
	if _, err := rand.Read(owner[:]); err != nil {
		return auth.ErrResHubUnavailable
	}
	ownerID := hex.EncodeToString(owner[:])
	// A disconnected browser must not cancel a pending logout. All Redis
	// connections are bounded by the pool's dial/read/write timeouts, including
	// the underlying store's SETEX/DEL. Its worst-case operation is below this
	// 30-second lease; acquisition itself is bounded separately.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return auth.ErrResHubUnavailable
	}
	defer conn.Close()
	for {
		locked, err := redis.String(redis.DoContext(conn, ctx, "SET", key+":lock", ownerID, "NX", "PX", 30000))
		if err == nil && locked == "OK" {
			break
		}
		if err != nil && err != redis.ErrNil {
			return auth.ErrResHubUnavailable
		}
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return auth.ErrResHubUnavailable
		case <-timer.C:
		}
	}
	defer func() {
		_, _ = conn.Do("EVAL", `if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) end return 0`, 1, key+":lock", ownerID)
	}()
	revoked, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return auth.ErrResHubUnavailable
	}
	if revoked {
		if session.Options != nil && session.Options.MaxAge <= 0 {
			// A late failure from the old session must not clear the cookie
			// issued by a concurrent account switch or fresh login. Still retry
			// server-side deletion if the original delete failed.
			return s.base.Save(r, sessionCookieSink{ResponseWriter: w, header: w.Header().Clone()}, session)
		}
		return auth.ErrResHubSession
	}
	if session.Options != nil && session.Options.MaxAge <= 0 {
		if _, err := conn.Do("SET", key, "1", "PX", s.retention.Milliseconds()); err != nil {
			return auth.ErrResHubUnavailable
		}
	}
	return s.base.Save(r, w, session)
}
