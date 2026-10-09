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

func TestCourseDataAndMaintenancePermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, permission := range []string{auth.PermissionCourseData, auth.PermissionMaintenance} {
		for _, tc := range []struct {
			name   string
			user   *auth.User
			status int
		}{
			{"anonymous", nil, 401},
			{"admin", &auth.User{Role: auth.RoleAdmin}, 403},
			{"delegated admin", &auth.User{Role: auth.RoleAdmin, Permissions: []string{permission}}, 204},
			{"regular user", &auth.User{Role: auth.RoleUser, Permissions: []string{permission}}, 403},
			{"superadmin", &auth.User{Role: auth.RoleSuperAdmin}, 204},
			{"unrelated delegation", &auth.User{Role: auth.RoleAdmin, Permissions: []string{auth.PermissionSettings}}, 403},
		} {
			t.Run(permission+"/"+tc.name, func(t *testing.T) {
				r := gin.New()
				r.Use(func(c *gin.Context) {
					if tc.user != nil {
						c.Request = c.Request.WithContext(auth.WithUser(c.Request.Context(), tc.user))
					}
				})
				r.GET("/restricted", RequirePermission(permission), func(c *gin.Context) { c.Status(204) })
				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest("GET", "/restricted", nil))
				if w.Code != tc.status {
					t.Fatalf("got %d want %d", w.Code, tc.status)
				}
			})
		}
	}
}
