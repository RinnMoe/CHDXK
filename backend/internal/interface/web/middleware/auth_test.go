package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"jcourse/internal/domain/auth"
)

func TestAuthReusesOptionalAuthUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &authMiddlewareUserRepo{user: &auth.User{ID: 7, Role: auth.RoleUser}}
	currentUserSvc := auth.NewCurrentUserService(repo)
	r := gin.New()
	r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
	r.Use(OptionalAuth(currentUserSvc))
	r.GET("/login-session", func(c *gin.Context) {
		if err := SetSessionUserID(c, repo.user.ID); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	r.GET("/protected", Auth(currentUserSvc), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	cookieValue := captureSessionCookie(t, r)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Cookie", cookieValue)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if repo.findByIDCalls != 1 {
		t.Fatalf("FindByID calls = %d, want 1", repo.findByIDCalls)
	}
}

func captureSessionCookie(t *testing.T, r http.Handler) string {
	t.Helper()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login-session", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("session setup status = %d, want %d", w.Code, http.StatusNoContent)
	}
	res := w.Result()
	defer res.Body.Close()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}
	return cookies[0].String()
}

type authMiddlewareUserRepo struct {
	user          *auth.User
	findByIDCalls int
}

func (r *authMiddlewareUserRepo) Create(context.Context, *auth.User) error { return nil }

func (r *authMiddlewareUserRepo) Update(context.Context, *auth.User) error { return nil }

func (r *authMiddlewareUserRepo) TouchLastSeen(context.Context, int, time.Time) error { return nil }

func (r *authMiddlewareUserRepo) FindByID(_ context.Context, id int) (*auth.User, error) {
	r.findByIDCalls++
	if r.user == nil || r.user.ID != id {
		return nil, nil
	}
	copy := *r.user
	return &copy, nil
}

func (r *authMiddlewareUserRepo) FindByUsername(context.Context, string) (*auth.User, error) {
	return nil, nil
}

func (r *authMiddlewareUserRepo) FindByEmail(context.Context, string) (*auth.User, error) {
	return nil, nil
}
