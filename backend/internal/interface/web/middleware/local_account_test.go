package middleware

import (
	"github.com/gin-gonic/gin"
	"jcourse/internal/domain/auth"
	"net/http/httptest"
	"testing"
)

func TestPasswordAndDelegatedPermissionGates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		role   string
		caps   []string
		force  bool
		path   string
		status int
	}{
		{auth.RoleSuperAdmin, nil, true, "/api/admin/users", 403}, {auth.RoleSuperAdmin, nil, true, "/api/auth/password", 204},
		{auth.RoleAdmin, nil, false, "/api/admin/users", 403}, {auth.RoleAdmin, []string{auth.PermissionUsers}, false, "/api/admin/users", 204},
		{auth.RoleUser, []string{auth.PermissionUsers}, false, "/api/admin/users", 403}, {auth.RoleSuperAdmin, nil, false, "/api/admin/users", 204},
	} {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Request = c.Request.WithContext(auth.WithUser(c.Request.Context(), &auth.User{ID: 8, Role: tc.role, Permissions: tc.caps, MustChangePassword: tc.force}))
		})
		r.Use(RequirePasswordChanged())
		r.GET("/api/admin/users", RequirePermission(auth.PermissionUsers), func(c *gin.Context) { c.Status(204) })
		r.GET("/api/auth/password", func(c *gin.Context) { c.Status(204) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		if w.Code != tc.status {
			t.Fatalf("role=%s caps=%v force=%v path=%s: got %d want %d", tc.role, tc.caps, tc.force, tc.path, w.Code, tc.status)
		}
	}
}
