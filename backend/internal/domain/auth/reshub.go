package auth

import (
	"context"
	"time"

	"jcourse/pkg/apperr"
)

// ResHubIdentity is the current identity snapshot returned by res-hub.
// Account role and account state are request-scoped, not a local authorization
// cache. CHDXK authorizes by IdentityRole; res-hub's selected permission mode
// and EffectiveRole do not downgrade an administrator account here.
type ResHubIdentity struct {
	Issuer           string
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
	AccessToken  string
	RefreshToken string
	IDToken      string
	SessionID    string
	ExpiresAt    time.Time
}

// ResHubSessionRefresher rotates a provider credential without exposing it to clients.
type ResHubSessionRefresher interface {
	Refresh(ctx context.Context, refreshToken string) (ResHubSession, error)
}

var (
	ErrResHubUnavailable = apperr.ServiceUnavailable("主站服务暂不可用，请稍后重试")
	ErrResHubSession     = apperr.Unauthorized("主站登录状态已失效，请重新登录")
	// A single-use refresh may have been consumed but its response was lost.
	ErrResHubRefreshUncertain = apperr.Unauthorized("登录续期结果无法确认，请重新登录")
	ErrResHubState            = apperr.BadRequest("单点登录状态无效")
	ErrIdentityNotFound       = apperr.NotFound("本站主体映射不存在")
)

// ResHubOIDCClient is the standard OIDC contract consumed by xk. Discovery,
// authorization-code exchange, userinfo, client credentials and logout are
// implemented by the infrastructure adapter.
type ResHubOIDCClient interface {
	AuthorizationURL(ctx context.Context, state, codeChallenge, nonce string) (string, error)
	Exchange(ctx context.Context, code, codeVerifier, nonce, redirectURI string) (ResHubSession, error)
	UserInfo(ctx context.Context, accessToken string) (*ResHubIdentity, error)
	ClientCredentialsToken(ctx context.Context, scopes []string) (string, error)
	Logout(ctx context.Context, idTokenHint, postLogoutRedirectURI string) error
}

// IdentityProjectionRepository maps an external subject to xk's local
// integer user id used by existing course and review tables.
type IdentityProjectionRepository interface {
	Upsert(ctx context.Context, identity *ResHubIdentity) (*User, error)
	SubjectByUserID(ctx context.Context, userID int) (string, error)
}
