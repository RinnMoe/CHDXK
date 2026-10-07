package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
)

type nativeTestClient struct {
	chooserClient
	revoked         bool
	redirect, nonce string
}

func (s *nativeTestClient) Exchange(_ context.Context, _, _, nonce, redirect string) (auth.ResHubSession, error) {
	s.redirect, s.nonce = redirect, nonce
	if nonce != "native_state_0123456789" {
		return auth.ResHubSession{}, auth.ErrResHubState
	}
	return auth.ResHubSession{AccessToken: "native-token", IDToken: "native-id"}, nil
}
func (s *nativeTestClient) UserInfo(ctx context.Context, token string) (*auth.ResHubIdentity, error) {
	if s.revoked {
		return nil, auth.ErrResHubSession
	}
	return s.chooserClient.UserInfo(ctx, token)
}

func TestNativeExchangeAndAccountSelection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, mode := range []string{"single", "multiple", "banned", "expired"} {
		t.Run(mode, func(t *testing.T) {
			user := &auth.User{ID: 8, Role: auth.RoleAdmin, AuthVersion: 1}
			repo := &auth.MockAccountRepository{User: user, Accounts: []*auth.User{user}}
			if mode == "multiple" || mode == "expired" {
				repo.Accounts = append(repo.Accounts, &auth.User{ID: 9, Role: auth.RoleUser, AuthVersion: 1})
			}
			if mode == "banned" {
				user.Suspend(time.Hour)
			}
			provider := &nativeTestClient{}
			resolution := application.NewAuthResolutionService(nil, nil, provider, nil)
			resolution.SetAccounts(repo)
			ctrl := NewAccountController(application.NewAccountQueryService(), application.NewResHubOIDCService(provider, nil, "https://web.example/callback", ""), application.NewLocalAccountService(repo, nil, nil, nil), resolution)
			r := gin.New()
			r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("native-test-secret"))))
			r.GET("/expire", func(c *gin.Context) { _ = middleware.SetNativePending(c, -time.Second); c.Status(204) })
			r.Use(middleware.ResolveCurrentUser(resolution), middleware.CSRF())
			r.GET("/api/auth/csrf", func(c *gin.Context) { c.Status(204) })
			r.POST("/api/auth/native/exchange", ctrl.NativeExchange)
			r.GET("/api/auth/accounts", ctrl.Accounts)
			r.POST("/api/auth/accounts/select", ctrl.SelectAccount)
			r.GET("/protected", middleware.RequireAuth(), func(c *gin.Context) { c.Status(204) })
			var session *http.Cookie
			csrf := ""
			request := func(method, path, body string) *httptest.ResponseRecorder {
				req := httptest.NewRequest(method, path, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				if session != nil {
					req.AddCookie(session)
				}
				if csrf != "" {
					req.Header.Set("X-CSRF-Token", csrf)
				}
				out := httptest.NewRecorder()
				r.ServeHTTP(out, req)
				for _, c := range out.Result().Cookies() {
					if c.Name == "jcourse_session" {
						session = c
					}
				}
				if value := out.Header().Get("X-CSRF-Token"); value != "" {
					csrf = value
				}
				return out
			}
			body := `{"code":"` + strings.Repeat("c", 43) + `","code_verifier":"` + strings.Repeat("v", 64) + `","state":"native_state_0123456789"}`
			if out := request("POST", "/api/auth/native/exchange", body); out.Code != 403 {
				t.Fatal("exchange bypassed CSRF")
			}
			request("GET", "/api/auth/csrf", "")
			if out := request("POST", "/api/auth/native/exchange", strings.Replace(body, "native_state_0123456789", "wrong_state_0123456789", 1)); out.Code != 401 {
				t.Fatalf("wrong nonce accepted: %d", out.Code)
			}
			out := request("POST", "/api/auth/native/exchange", body)
			if provider.redirect != application.NativeRedirectURI || provider.nonce != "native_state_0123456789" {
				t.Fatal("native redirect or nonce lost")
			}
			if mode == "banned" {
				if out.Code != 403 {
					t.Fatalf("banned account accepted: %d", out.Code)
				}
				return
			}
			if out.Code != 200 {
				t.Fatalf("exchange failed: %d %s", out.Code, out.Body.String())
			}
			var response map[string]any
			_ = json.Unmarshal(out.Body.Bytes(), &response)
			request("GET", "/api/auth/csrf", "") // Exchange rotates the session and CSRF token.
			if mode == "multiple" || mode == "expired" {
				if response["status"] != "account_selection_required" || response["expires_in"] != float64(300) {
					t.Fatal(response)
				}
				if out := request("GET", "/protected", ""); out.Code != 401 {
					t.Fatal("pending session reached business API")
				}
				if mode == "expired" {
					request("GET", "/expire", "")
					if out := request("GET", "/api/auth/accounts", ""); out.Code != 401 {
						t.Fatal("expired choice accepted")
					}
					return
				}
				if out := request("POST", "/api/auth/accounts/select", `{"id":10}`); out.Code != 403 {
					t.Fatal("unbound account accepted")
				}
				if out := request("POST", "/api/auth/accounts/select", `{"id":8}`); out.Code != 200 {
					t.Fatalf("selection failed: %s", out.Body.String())
				}
			} else if response["status"] != "authenticated" {
				t.Fatal(response)
			}
			if out := request("GET", "/protected", ""); out.Code != 204 {
				t.Fatalf("business access failed: %d", out.Code)
			}
			provider.revoked = true
			if out := request("GET", "/protected", ""); out.Code != 401 {
				t.Fatal("revoked parent session remained authenticated")
			}
		})
	}
}
