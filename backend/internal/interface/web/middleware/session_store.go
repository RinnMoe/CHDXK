package middleware

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"maps"
	"net/http"
	"reflect"
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

// Bound to one request's Gorilla session; no snapshot data is persisted.
type sessionSnapshotStore struct {
	*revokedSessionStore
	id     string
	values map[any]any
}

type sessionSnapshotContextKey struct {
	store *revokedSessionStore
	name  string
}

var oidcCredentialKeys = []string{
	sessionKeyOIDCAccessToken, sessionKeyOIDCIDToken, sessionKeyOIDCRefreshToken,
	sessionKeyOIDCExpiresAt, sessionKeyOIDCDeadline,
}

type sessionCookieSink struct {
	http.ResponseWriter
	header http.Header
}

func (w sessionCookieSink) Header() http.Header { return w.header }

func (s *revokedSessionStore) Options(options sessions.Options) { s.base.Options(options) }

func (s *revokedSessionStore) Get(r *http.Request, name string) (*gsessions.Session, error) {
	key := sessionSnapshotContextKey{store: s, name: name}
	snapshot, _ := r.Context().Value(key).(*sessionSnapshotStore)
	if snapshot == nil {
		snapshot = &sessionSnapshotStore{revokedSessionStore: s}
		*r = *r.WithContext(context.WithValue(r.Context(), key, snapshot))
	}
	// Gorilla rebinds the returned session to the store passed to Registry.Get.
	return gsessions.GetRegistry(r).Get(snapshot, name)
}

func (s *revokedSessionStore) New(r *http.Request, name string) (*gsessions.Session, error) {
	return (&sessionSnapshotStore{revokedSessionStore: s}).New(r, name)
}

func (s *sessionSnapshotStore) New(r *http.Request, name string) (*gsessions.Session, error) {
	loaded, err := s.base.New(r, name)
	if loaded == nil {
		return nil, err
	}
	// Gorilla keeps the store private; construct a session bound to this gate so
	// direct Session.Save calls made by resetSession also pass through it.
	s.id, s.values = loaded.ID, maps.Clone(loaded.Values)
	wrapped := gsessions.NewSession(s, name)
	wrapped.ID, wrapped.Values, wrapped.Options, wrapped.IsNew = loaded.ID, loaded.Values, loaded.Options, loaded.IsNew
	if wrapped.IsNew {
		// A cookie whose Redis entry expired starts a fresh anonymous session.
		wrapped.ID, s.id = "", ""
	}
	return wrapped, err
}

func (s *revokedSessionStore) Save(r *http.Request, w http.ResponseWriter, session *gsessions.Session) (saveErr error) {
	snapshot, _ := session.Store().(*sessionSnapshotStore)
	defer func() {
		if saveErr == nil && snapshot != nil {
			snapshot.id, snapshot.values = session.ID, maps.Clone(session.Values)
		}
	}()
	if session.ID == "" {
		if err := boundSessionDeadline(session); err != nil {
			return err
		}
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
	if snapshot != nil && snapshot.id == session.ID && (session.Options.MaxAge > 0 || len(session.Values) > 0) {
		data, err := redis.Bytes(conn.Do("GET", sessionRedisPrefix+session.ID))
		if err == redis.ErrNil {
			return auth.ErrResHubSession
		}
		if err != nil {
			return auth.ErrResHubUnavailable
		}
		var latest map[any]any
		// NewSessionStore uses RediStore's default Gob session serializer.
		if gob.NewDecoder(bytes.NewReader(data)).Decode(&latest) != nil {
			return auth.ErrResHubUnavailable
		}
		session.Values = mergeSessionChanges(snapshot.values, session.Values, latest)
	}
	if err := boundSessionDeadline(session); err != nil {
		return err
	}
	if session.Options != nil && session.Options.MaxAge <= 0 {
		if _, err := conn.Do("SET", key, "1", "PX", s.retention.Milliseconds()); err != nil {
			return auth.ErrResHubUnavailable
		}
	}
	return s.base.Save(r, w, session)
}

func boundSessionDeadline(session *gsessions.Session) error {
	deadline, _ := session.Values[sessionKeyOIDCDeadline].(int64)
	if deadline == 0 || session.Options == nil || session.Options.MaxAge <= 0 {
		return nil
	}
	remaining := int(deadline - time.Now().Unix())
	if remaining <= 0 {
		return auth.ErrResHubSession
	}
	if remaining < session.Options.MaxAge {
		options := *session.Options
		options.MaxAge = remaining
		session.Options = &options
	}
	return nil
}

// Session values written by this application are scalar values. Apply only the
// caller's changes, retaining unrelated concurrent writes and explicit deletions.
func mergeSessionChanges(original, incoming, latest map[any]any) map[any]any {
	merged := maps.Clone(latest)
	if merged == nil {
		merged = make(map[any]any)
	}
	for key, value := range incoming {
		before, existed := original[key]
		if !existed || !reflect.DeepEqual(before, value) {
			merged[key] = value
		}
	}
	for key := range original {
		if _, exists := incoming[key]; !exists {
			delete(merged, key)
		}
	}
	// Concurrent refreshes can finish out of order. If another request has
	// already advanced the credentials, use that entire authoritative bundle.
	for _, key := range oidcCredentialKeys {
		before, hadBefore := original[key]
		current, hasCurrent := latest[key]
		if hadBefore != hasCurrent || !reflect.DeepEqual(before, current) {
			for _, credentialKey := range oidcCredentialKeys {
				if value, exists := latest[credentialKey]; exists {
					merged[credentialKey] = value
				} else {
					delete(merged, credentialKey)
				}
			}
			break
		}
	}
	return merged
}
