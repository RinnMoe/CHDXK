package middleware

import (
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func TestSessionStoreMergesConcurrentChanges(t *testing.T) {
	for _, refreshFirst := range []bool{true, false} {
		t.Run(map[bool]string{true: "state-after-refresh", false: "refresh-after-state"}[refreshFirst], func(t *testing.T) {
			first, second := revocationTestStores(t)
			request := httptest.NewRequest("GET", "/", nil)
			seed, err := first.New(request, "merge_test")
			if err != nil {
				t.Fatal("create session")
			}
			deadline := time.Now().Add(time.Hour).Unix()
			seed.Values = map[any]any{
				sessionKeyOIDCAccessToken: "old-access", sessionKeyOIDCIDToken: "id-token",
				sessionKeyOIDCRefreshToken: "old-refresh", sessionKeyOIDCExpiresAt: time.Now().Add(-time.Minute).Unix(),
				sessionKeyOIDCDeadline: deadline, sessionKeyOIDCNonce: "consumed-nonce",
			}
			response := httptest.NewRecorder()
			if err := seed.Save(request, response); err != nil {
				t.Fatal("save seed")
			}
			revocationTestCleanupID(t, first, seed.ID)
			cookie := response.Result().Cookies()[0]
			stateRequest := httptest.NewRequest("GET", "/", nil)
			stateRequest.AddCookie(cookie)
			refreshRequest := httptest.NewRequest("GET", "/", nil)
			refreshRequest.AddCookie(cookie)
			state, err := first.New(stateRequest, "merge_test")
			if err != nil {
				t.Fatal("load state request")
			}
			refresh, err := second.New(refreshRequest, "merge_test")
			if err != nil {
				t.Fatal("load refresh request")
			}
			state.Values[sessionKeyOIDCState] = "new-state"
			delete(state.Values, sessionKeyOIDCNonce)
			refresh.Values[sessionKeyOIDCAccessToken] = "new-access"
			refresh.Values[sessionKeyOIDCRefreshToken] = "new-refresh"
			refresh.Values[sessionKeyOIDCExpiresAt] = time.Now().Add(30 * time.Minute).Unix()
			refresh.Values[sessionKeyCSRFToken] = "new-csrf"
			saveState := func() {
				if err := state.Save(stateRequest, httptest.NewRecorder()); err != nil {
					t.Fatal("save state")
				}
			}
			saveRefresh := func() {
				if err := refresh.Save(refreshRequest, httptest.NewRecorder()); err != nil {
					t.Fatal("save refresh")
				}
			}
			if refreshFirst {
				saveRefresh()
				saveState()
			} else {
				saveState()
				saveRefresh()
			}
			check, err := second.New(refreshRequest, "merge_test")
			if err != nil {
				t.Fatal("reload merged session")
			}
			if check.Values[sessionKeyOIDCAccessToken] != "new-access" || check.Values[sessionKeyOIDCRefreshToken] != "new-refresh" ||
				check.Values[sessionKeyOIDCState] != "new-state" || check.Values[sessionKeyCSRFToken] != "new-csrf" || check.Values[sessionKeyOIDCDeadline] != deadline {
				t.Fatal("concurrent write lost refreshed credentials or independent state")
			}
			if _, exists := check.Values[sessionKeyOIDCNonce]; exists {
				t.Fatal("concurrent save restored consumed nonce")
			}
			// Another save in the same request must use its updated baseline,
			// preserving concurrently added state while retaining explicit deletes.
			check.Values["other-state"] = "retained"
			if err := check.Save(refreshRequest, httptest.NewRecorder()); err != nil {
				t.Fatal("save peer state")
			}
			delete(state.Values, sessionKeyOIDCState)
			if err := state.Save(stateRequest, httptest.NewRecorder()); err != nil {
				t.Fatal("repeat state save")
			}
			check, err = second.New(refreshRequest, "merge_test")
			if err != nil {
				t.Fatal("reload repeated save")
			}
			if _, exists := check.Values[sessionKeyOIDCState]; exists {
				t.Fatal("explicit deletion lost")
			}
			if check.Values["other-state"] != "retained" || check.Values[sessionKeyOIDCRefreshToken] != "new-refresh" {
				t.Fatal("repeated save lost concurrent updates")
			}
		})
	}
}

func TestSessionStoreStaleAccountRotationRetainsFreshCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	first, second := revocationTestStores(t)
	request := httptest.NewRequest("GET", "/", nil)
	seed, err := first.New(request, "rotate_merge_test")
	if err != nil {
		t.Fatal("create seed")
	}
	deadline := time.Now().Add(10 * time.Minute).Unix()
	seed.Values = map[any]any{sessionKeyOIDCAccessToken: "old-access", sessionKeyOIDCIDToken: "id-token", sessionKeyOIDCRefreshToken: "old-refresh", sessionKeyOIDCExpiresAt: time.Now().Add(time.Hour).Unix(), sessionKeyOIDCDeadline: deadline, "account_id": 7, "account_version": 1, sessionKeyCSRFToken: "old-csrf"}
	response := httptest.NewRecorder()
	if err := seed.Save(request, response); err != nil {
		t.Fatal("save seed")
	}
	revocationTestCleanupID(t, first, seed.ID)
	cookie := response.Result().Cookies()[0]
	r := gin.New()
	r.Use(sessions.Sessions("rotate_merge_test", first))
	r.GET("/select", func(c *gin.Context) {
		_ = OIDCAccessToken(c) // Load the account chooser's old snapshot.
		peerRequest := httptest.NewRequest("GET", "/", nil)
		peerRequest.AddCookie(cookie)
		peer, err := second.New(peerRequest, "rotate_merge_test")
		if err != nil {
			t.Fatal("load peer")
		}
		peer.Values[sessionKeyOIDCAccessToken] = "new-access"
		peer.Values[sessionKeyOIDCRefreshToken] = "new-refresh"
		if err := peer.Save(peerRequest, httptest.NewRecorder()); err != nil {
			t.Fatal("refresh peer")
		}
		if err := SetAccountSession(c, 8, 1, true); err != nil {
			t.Fatal("rotate stale account snapshot")
		}
		revocationTestCleanupID(t, first, sessions.Default(c).ID())
	})
	selectRequest := httptest.NewRequest("GET", "/select", nil)
	selectRequest.AddCookie(cookie)
	out := httptest.NewRecorder()
	r.ServeHTTP(out, selectRequest)
	cookies := out.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("missing account cookie")
	}
	rotatedCookie := cookies[len(cookies)-1]
	if rotatedCookie.MaxAge > 600 {
		t.Fatal("rotation extended absolute lifetime")
	}
	reload := httptest.NewRequest("GET", "/", nil)
	reload.AddCookie(rotatedCookie)
	current, err := second.New(reload, "rotate_merge_test")
	if err != nil {
		t.Fatal("reload new account")
	}
	if current.ID == seed.ID || current.Values["account_id"] != 8 || current.Values[sessionKeyOIDCAccessToken] != "new-access" || current.Values[sessionKeyOIDCRefreshToken] != "new-refresh" || current.Values[sessionKeyOIDCDeadline] != deadline {
		t.Fatalf("rotation state: newID=%v account=%v accessCurrent=%v refreshCurrent=%v deadlinePreserved=%v", current.ID != seed.ID, current.Values["account_id"] == 8, current.Values[sessionKeyOIDCAccessToken] == "new-access", current.Values[sessionKeyOIDCRefreshToken] == "new-refresh", current.Values[sessionKeyOIDCDeadline] == deadline)
	}
	if _, exists := current.Values[sessionKeyCSRFToken]; exists {
		t.Fatal("rotation retained old CSRF")
	}
}

func TestMergeSessionChangesKeepsCredentialBundleTogether(t *testing.T) {
	original := map[any]any{sessionKeyOIDCAccessToken: "old", sessionKeyOIDCRefreshToken: "old-refresh", sessionKeyOIDCExpiresAt: int64(1)}
	incoming := map[any]any{sessionKeyOIDCAccessToken: "intermediate", sessionKeyOIDCRefreshToken: "intermediate-refresh", sessionKeyOIDCExpiresAt: int64(2), sessionKeyOIDCState: "state"}
	latest := map[any]any{sessionKeyOIDCAccessToken: "latest", sessionKeyOIDCRefreshToken: "latest-refresh", sessionKeyOIDCExpiresAt: int64(3)}
	merged := mergeSessionChanges(original, incoming, latest)
	if merged[sessionKeyOIDCState] != "state" {
		t.Fatal("lost independent state")
	}
	delete(merged, sessionKeyOIDCState)
	if !reflect.DeepEqual(merged, latest) {
		t.Fatal("mixed or downgraded credential generations")
	}
}

func TestSessionStoreExpiredCookieStartsFreshSession(t *testing.T) {
	first, second := revocationTestStores(t)
	request := httptest.NewRequest("GET", "/", nil)
	seed, err := first.New(request, "expired_merge_test")
	if err != nil {
		t.Fatal("create seed")
	}
	seed.Values["account_id"] = 7
	w := httptest.NewRecorder()
	if err := seed.Save(request, w); err != nil {
		t.Fatal("save seed")
	}
	revocationTestCleanupID(t, first, seed.ID)
	conn := first.pool.Get()
	defer conn.Close()
	if _, err := conn.Do("DEL", sessionRedisPrefix+seed.ID); err != nil {
		t.Fatal("expire Redis session")
	}
	expiredRequest := httptest.NewRequest("GET", "/", nil)
	expiredRequest.AddCookie(w.Result().Cookies()[0])
	fresh, err := second.Get(expiredRequest, "expired_merge_test")
	if err != nil {
		t.Fatal("load expired session")
	}
	if fresh.ID != "" || len(fresh.Values) != 0 {
		t.Fatal("expired session retained identity")
	}
	fresh.Values[sessionKeyCSRFToken] = "new-csrf"
	if err := fresh.Save(expiredRequest, httptest.NewRecorder()); err != nil {
		t.Fatal("save fresh anonymous session")
	}
	revocationTestCleanupID(t, second, fresh.ID)
	again, err := second.Get(expiredRequest, "expired_merge_test")
	if err != nil || again != fresh {
		t.Fatal("repeated Get lost request snapshot")
	}
	again.Values[sessionKeyOIDCState] = "state"
	if err := again.Save(expiredRequest, httptest.NewRecorder()); err != nil {
		t.Fatal("repeated Get broke later save")
	}
	if again.ID == seed.ID || again.Values[sessionKeyCSRFToken] != "new-csrf" {
		t.Fatal("fresh session reused expired ID or lost state")
	}
}
