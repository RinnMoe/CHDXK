package controller

import (
	"context"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type chooserClient struct{ auth.ResHubOIDCClient }

func (*chooserClient) UserInfo(context.Context, string) (*auth.ResHubIdentity, error) {
	return &auth.ResHubIdentity{Issuer: "https://hub.test", Subject: "user:8"}, nil
}

type bindingClient struct{ chooserClient }

func (*bindingClient) Exchange(context.Context, string, string, string, string) (auth.ResHubSession, error) {
	return auth.ResHubSession{AccessToken: "bound-token", IDToken: "bound-id-token"}, nil
}

type bindingRepository struct {
	auth.MockAccountRepository
	binds, projections int
}

func (r *bindingRepository) BindAccount(_ context.Context, id int, identity *auth.ResHubIdentity) error {
	r.binds++
	r.User.Subject = identity.Subject
	r.User.AuthVersion++
	return nil
}
func (r *bindingRepository) ResHubAccounts(context.Context, *auth.ResHubIdentity, bool) ([]*auth.User, error) {
	r.projections++
	return []*auth.User{r.User}, nil
}

func TestBindingCallbackValidatesStateAndAccountVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name       string
		version    int
		forced     bool
		wantStatus int
	}{{"valid", 3, false, 302}, {"stale account", 2, false, 401}, {"initial password", 3, true, 401}} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &bindingRepository{MockAccountRepository: auth.MockAccountRepository{User: &auth.User{ID: 8, Role: auth.RoleUser, AuthVersion: 3, MustChangePassword: tc.forced}}}
			client := &bindingClient{}
			ctrl := NewAccountController(application.NewAccountQueryService(), application.NewResHubOIDCService(client, nil, "", ""), application.NewLocalAccountService(repo, nil, nil, nil), nil)
			r := gin.New()
			r.Use(sessions.Sessions("binding_session", cookie.NewStore([]byte("binding-secret"))))
			r.GET("/setup", func(c *gin.Context) {
				_ = middleware.SetOIDCBindingTarget(c, 8, tc.version)
				_ = middleware.SetOIDCState(c, "expected-state", "verifier", "nonce", "/settings")
				c.Status(204)
			})
			r.GET("/callback", ctrl.OIDCCallback)
			r.GET("/session", func(c *gin.Context) {
				c.JSON(200, gin.H{"id": middleware.AccountSessionID(c), "version": middleware.AccountSessionVersion(c)})
			})
			var session *http.Cookie
			request := func(path string) *httptest.ResponseRecorder {
				req := httptest.NewRequest("GET", path, nil)
				if session != nil {
					req.AddCookie(session)
				}
				out := httptest.NewRecorder()
				r.ServeHTTP(out, req)
				for _, updated := range out.Result().Cookies() {
					session = updated
				}
				return out
			}
			request("/setup")
			request("/callback?state=wrong-state&code=code")
			if repo.binds != 0 {
				t.Fatal("state mismatch changed binding")
			}
			out := request("/callback?state=expected-state&code=code")
			if out.Code != tc.wantStatus {
				t.Fatalf("callback status %d", out.Code)
			}
			if tc.wantStatus == 302 {
				if repo.binds != 1 || repo.projections != 0 || out.Header().Get("Location") != "/settings" {
					t.Fatal("binding created extra account or lost binding target")
				}
				if state := request("/session"); !strings.Contains(state.Body.String(), `"id":8`) || !strings.Contains(state.Body.String(), `"version":4`) {
					t.Fatal("current session did not adopt new binding version")
				}
			} else if repo.binds != 0 {
				t.Fatal("invalid account bound identity")
			}
		})
	}
}
func TestAccountChoiceRequiresBindingBeforeBusinessAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &auth.MockAccountRepository{User: &auth.User{ID: 8, Role: auth.RoleUser, AuthVersion: 1}, Accounts: []*auth.User{{ID: 7, Role: auth.RoleUser, AuthVersion: 1}, {ID: 8, Role: auth.RoleUser, AuthVersion: 1}}}
	client := &chooserClient{}
	resolution := application.NewAuthResolutionService(nil, nil, client, nil)
	resolution.SetAccounts(repo)
	ctrl := NewAccountController(application.NewAccountQueryService(), application.NewResHubOIDCService(client, nil, "", ""), application.NewLocalAccountService(repo, nil, nil, nil), resolution)
	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("chooser-secret"))))
	r.GET("/setup", func(c *gin.Context) { _ = middleware.SetSessionOIDC(c, "token", "id-token"); c.Status(204) })
	r.Use(middleware.ResolveCurrentUser(resolution))
	r.GET("/api/auth/accounts", ctrl.Accounts)
	r.POST("/api/auth/accounts/select", ctrl.SelectAccount)
	r.GET("/protected", middleware.RequireAuth(), func(c *gin.Context) { c.Status(204) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/setup", nil))
	cookies := w.Result().Cookies()
	session := cookies[len(cookies)-1]
	request := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(session)
		out := httptest.NewRecorder()
		r.ServeHTTP(out, req)
		return out
	}
	if out := request("GET", "/protected", ""); out.Code != 401 {
		t.Fatalf("pending selection reached business API: %d", out.Code)
	}
	if out := request("GET", "/api/auth/accounts", ""); out.Code != 200 {
		t.Fatalf("account choices unavailable: %d", out.Code)
	}
	if out := request("POST", "/api/auth/accounts/select", `{"id":9}`); out.Code != 403 {
		t.Fatalf("unbound selection accepted: %d", out.Code)
	}
	selected := request("POST", "/api/auth/accounts/select", `{"id":8}`)
	if selected.Code != http.StatusOK {
		t.Fatalf("selection failed: %d", selected.Code)
	}
	cookies = selected.Result().Cookies()
	session = cookies[len(cookies)-1]
	if out := request("GET", "/protected", ""); out.Code != 204 {
		t.Fatalf("selected account not authenticated: %d", out.Code)
	}
}
