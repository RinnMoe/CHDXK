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
