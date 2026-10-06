package auth

import (
	"context"
	"jcourse/pkg/apperr"
	"strings"
	"time"
)

const (
	PermissionUsers       = "users"
	PermissionRoles       = "roles"
	PermissionPoints      = "points"
	PermissionSettings    = "settings"
	PermissionMaintenance = "maintenance"
)

var AdminPermissions = []string{PermissionUsers, PermissionRoles, PermissionPoints, PermissionSettings, PermissionMaintenance}
var ErrLocalLogin = apperr.Unauthorized("账号或密码不正确")
var ErrAccountUnavailable = apperr.Forbidden("该本站账号已被停用")
var ErrMuted = apperr.Forbidden("你暂时无法发表或修改点评")
var ErrPasswordRequired = apperr.Forbidden("请先修改初始密码")

func (u *User) Can(permission string) bool {
	if u == nil {
		return false
	}
	if u.IsSuperAdmin() {
		return true
	}
	if !u.IsAdmin() {
		return false
	}
	for _, p := range u.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}
func (u *User) IsMuted() bool { return u.MutedUntil != nil && u.MutedUntil.After(time.Now()) }
func EffectiveRole(local, remote string) string {
	if local == RoleSuperAdmin {
		return RoleSuperAdmin
	}
	if local == RoleAdmin {
		return RoleAdmin
	}
	switch strings.ToLower(remote) {
	case "admin", "administrator", "super_admin", "super-admin", "superadmin":
		return RoleAdmin
	}
	return RoleUser
}

type AccountCredential struct {
	User         *User
	PasswordHash string
}
type InitialCredential struct{ Username, Password string }
type AccountFilter struct {
	Query, Role string
	Page        int
}
type AccountRepository interface {
	Initialize(context.Context) (*InitialCredential, error)
	FindCredential(context.Context, string) (*AccountCredential, error)
	FindByID(context.Context, int) (*User, error)
	CreateLocal(context.Context, string, string, string, bool) (*User, error)
	SetPassword(context.Context, int, string, bool) error
	UpdateProfile(context.Context, int, string, string) error
	ListAccounts(context.Context, AccountFilter) ([]*User, int64, error)
	SaveAccount(context.Context, *User) error
	DeleteAccount(context.Context, int) error
	ResHubAccounts(context.Context, *ResHubIdentity, bool) ([]*User, error)
	BindAccount(context.Context, int, *ResHubIdentity) error
	UnbindAccount(context.Context, int) error
}
