package auth

import (
	"context"
	"time"

	"jcourse/pkg/apperr"
)

// ResHubIdentity is the current identity snapshot returned by res-hub.
// Role and account state are intentionally request-scoped and must not be
// treated as a local authorization cache.
type ResHubIdentity struct {
	Subject          string
	Username         string
	Email            string
	IdentityRole     string
	EffectiveRole    string
	PermissionMode   string
	ViolationCount   int
	InitialPassword  bool
	HasPointsAccount bool
	SessionID        string
	ExpiresAt        time.Time
	IdentityType     string
	ImpersonatedBy   string
}

type ResHubSession struct {
	Token     string
	ExpiresAt time.Time
}

var (
	ErrResHubUnavailable = apperr.ServiceUnavailable("主站服务暂不可用，请稍后重试")
	ErrResHubSession     = apperr.Unauthorized("主站登录状态已失效，请重新登录")
	ErrResHubState       = apperr.BadRequest("单点登录状态无效")
	ErrIdentityNotFound  = apperr.NotFound("本站主体映射不存在")
)

// ResHubClient contains only the versioned first-party contract consumed by
// xk. Its implementation belongs to infrastructure.
type ResHubClient interface {
	Exchange(ctx context.Context, code, codeVerifier, redirectURI string) (ResHubSession, error)
	Introspect(ctx context.Context, token string) (*ResHubIdentity, error)
	Logout(ctx context.Context, token string) error
}

type ResHubAuthorizeClient interface {
	ResHubClient
	AuthorizeURL(state, codeChallenge string) (string, error)
}

// IdentityProjectionRepository maps an external subject to xk's local
// integer user id used by existing course and review tables.
type IdentityProjectionRepository interface {
	Upsert(ctx context.Context, identity *ResHubIdentity) (*User, error)
	SubjectByUserID(ctx context.Context, userID int) (string, error)
}
