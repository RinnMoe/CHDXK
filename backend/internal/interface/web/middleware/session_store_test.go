package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	gsessions "github.com/gorilla/sessions"
	"jcourse/internal/domain/auth"
	"jcourse/internal/infrastructure/persistence"
)

func revocationTestStores(t *testing.T) (*revokedSessionStore, *revokedSessionStore) {
	t.Helper()
	addr := os.Getenv("JCOURSE_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("JCOURSE_TEST_REDIS_ADDR is not set")
	}
	newStore := func() *revokedSessionStore {
		store, err := NewSessionStore(persistence.RedisConfig{Addr: addr}, SessionConfig{Secret: "session-revocation-integration-secret", MaxAge: 3600})
		if err != nil {
			t.Fatal("create isolated Redis session store")
		}
		wrapped := store.(*revokedSessionStore)
		t.Cleanup(func() { _ = wrapped.pool.Close() })
		return wrapped
	}
	return newStore(), newStore()
}

func revocationTestCleanupID(t *testing.T, store *revokedSessionStore, id string) {
	t.Helper()
	t.Cleanup(func() {
		conn := store.pool.Get()
		defer conn.Close()
		digest := sha256.Sum256([]byte(id))
		key := "jcourse:session-revocation:" + hex.EncodeToString(digest[:])
		_, _ = conn.Do("DEL", sessionRedisPrefix+id, key, key+":lock")
	})
}

func TestSessionStoreSlowRefreshCannotRestoreDeletedSession(t *testing.T) {
	for _, operation := range []string{"logout", "account-rotation"} {
		t.Run(operation, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			first, second := revocationTestStores(t)
			seedRouter := gin.New()
			seedRouter.Use(sessions.Sessions("revocation_test", first))
			var oldID string
			seedRouter.GET("/seed", func(c *gin.Context) {
				if err := SetSessionOIDCCredentials(c, auth.ResHubSession{AccessToken: "test-access", IDToken: "test-id", RefreshToken: "test-refresh", ExpiresAt: time.Now().Add(-time.Minute)}); err != nil {
					t.Fatal("seed provider session")
				}
				if err := SetAccountSession(c, 7, 1, false); err != nil {
					t.Fatal("seed account session")
				}
				oldID = sessions.Default(c).ID()
				c.Status(204)
			})
			seed := httptest.NewRecorder()
			seedRouter.ServeHTTP(seed, httptest.NewRequest("GET", "/seed", nil))
			revocationTestCleanupID(t, first, oldID)
			cookies := seed.Result().Cookies()
			cookie := cookies[len(cookies)-1]
			loaded, release := make(chan struct{}), make(chan struct{})
			businessCalls := 0
			slow := gin.New()
			slow.Use(sessions.Sessions("revocation_test", first), RefreshOIDCSession(&auth.MockResHubSessionRefresher{OnRefresh: func(context.Context, string) (auth.ResHubSession, error) {
				close(loaded)
				<-release
				return auth.ResHubSession{AccessToken: "test-refreshed-access", RefreshToken: "test-refreshed-grant", ExpiresAt: time.Now().Add(time.Hour)}, nil
			}}))
			slow.GET("/refresh", func(c *gin.Context) {
				businessCalls++
				c.Status(204)
			})
			slowRequest := httptest.NewRequest("GET", "/refresh", nil)
			slowRequest.AddCookie(cookie)
			slowResponse := httptest.NewRecorder()
			done := make(chan struct{})
			go func() { defer close(done); slow.ServeHTTP(slowResponse, slowRequest) }()
			select {
			case <-loaded:
			case <-time.After(3 * time.Second):
				t.Fatal("slow request did not load")
			}
			mutator := gin.New()
			mutator.Use(sessions.Sessions("revocation_test", second))
			var mutationErr error
			var newID string
			mutator.GET("/mutate", func(c *gin.Context) {
				if operation == "logout" {
					mutationErr = ClearSession(c)
				} else {
					mutationErr = SetAccountSession(c, 8, 1, true)
					newID = sessions.Default(c).ID()
				}
				c.Status(204)
			})
			mutation := httptest.NewRequest("GET", "/mutate", nil)
			mutation.AddCookie(cookie)
			mutationResponse := httptest.NewRecorder()
			mutator.ServeHTTP(mutationResponse, mutation)
			close(release)
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				t.Fatal("slow save did not finish")
			}
			if mutationErr != nil {
				t.Fatal("logout/rotation failed")
			}
			if slowResponse.Code != http.StatusUnauthorized || businessCalls != 0 {
				t.Fatal("stale refresh reached the business handler or did not return unauthorized")
			}
			if len(slowResponse.Result().Cookies()) != 0 {
				t.Fatal("stale refresh emitted a session cookie")
			}
			conn := first.pool.Get()
			defer conn.Close()
			if exists, err := redis.Bool(conn.Do("EXISTS", sessionRedisPrefix+oldID)); err != nil || exists {
				t.Fatal("deleted session was recreated")
			}
			if operation == "account-rotation" {
				revocationTestCleanupID(t, first, newID)
				if newID == "" || newID == oldID {
					t.Fatal("account rotation did not create a fresh ID")
				}
				if exists, err := redis.Bool(conn.Do("EXISTS", sessionRedisPrefix+newID)); err != nil || !exists {
					t.Fatal("new account session was lost")
				}
			}
		})
	}
}

func TestSessionStoreCancelledLogoutStillRevokes(t *testing.T) {
	first, second := revocationTestStores(t)
	request := httptest.NewRequest("GET", "/", nil)
	session, err := first.New(request, "cancelled_logout_test")
	if err != nil {
		t.Fatal("create session")
	}
	session.Values["account_id"] = 7
	w := httptest.NewRecorder()
	if err := session.Save(request, w); err != nil {
		t.Fatal("save initial session")
	}
	revocationTestCleanupID(t, first, session.ID)
	request2 := httptest.NewRequest("GET", "/", nil)
	request2.AddCookie(w.Result().Cookies()[0])
	stale, err := second.New(request2, "cancelled_logout_test")
	if err != nil {
		t.Fatal("load stale session")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	session.Options.MaxAge = -1
	if err := session.Save(request.WithContext(ctx), httptest.NewRecorder()); err != nil {
		t.Fatal("cancelled request prevented logout")
	}
	if err := stale.Save(request2, httptest.NewRecorder()); !errors.Is(err, auth.ErrResHubSession) {
		t.Fatal("cancelled logout failed to revoke old session")
	}
}

type delayedSessionStore struct {
	sessions.Store
	entered chan struct{}
	release chan struct{}
}

func (s *delayedSessionStore) Save(r *http.Request, w http.ResponseWriter, session *gsessions.Session) error {
	close(s.entered)
	<-s.release
	return s.Store.Save(r, w, session)
}

func TestSessionStoreLogoutSerializesWithWriteAlreadyInProgress(t *testing.T) {
	first, second := revocationTestStores(t)
	request := httptest.NewRequest("GET", "/", nil)
	session, err := first.New(request, "locked_logout_test")
	if err != nil {
		t.Fatal("create session")
	}
	session.Values["account_id"] = 7
	seed := httptest.NewRecorder()
	if err := session.Save(request, seed); err != nil {
		t.Fatal("save initial session")
	}
	revocationTestCleanupID(t, first, session.ID)
	request2 := httptest.NewRequest("GET", "/", nil)
	request2.AddCookie(seed.Result().Cookies()[0])
	logout, err := second.New(request2, "locked_logout_test")
	if err != nil {
		t.Fatal("load logout session")
	}
	delayed := &delayedSessionStore{Store: first.base, entered: make(chan struct{}), release: make(chan struct{})}
	first.base = delayed
	written, deleted := make(chan error, 1), make(chan error, 1)
	go func() { written <- session.Save(request, httptest.NewRecorder()) }()
	select {
	case <-delayed.entered:
	case <-time.After(3 * time.Second):
		t.Fatal("save did not acquire gate")
	}
	logout.Options.MaxAge = -1
	go func() { deleted <- logout.Save(request2, httptest.NewRecorder()) }()
	select {
	case <-deleted:
		t.Fatal("logout bypassed a save holding the shared session lock")
	case <-time.After(60 * time.Millisecond):
	}
	close(delayed.release)
	if err := <-written; err != nil {
		t.Fatal("first write failed")
	}
	if err := <-deleted; err != nil {
		t.Fatal("serialized logout failed")
	}
	conn := first.pool.Get()
	defer conn.Close()
	if exists, err := redis.Bool(conn.Do("EXISTS", sessionRedisPrefix+session.ID)); err != nil || exists {
		t.Fatal("write raced past logout deletion")
	}
}
