package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"strings"

	"jcourse/internal/domain/auth"
)

type ResHubOIDCService struct {
	client          auth.ResHubOIDCClient
	projection      auth.IdentityProjectionRepository
	redirectURI     string
	logoutReturnURI string
}

func NewResHubOIDCService(client auth.ResHubOIDCClient, projection auth.IdentityProjectionRepository, redirectURI, logoutReturnURI string) *ResHubOIDCService {
	return &ResHubOIDCService{
		client:          client,
		projection:      projection,
		redirectURI:     strings.TrimSpace(redirectURI),
		logoutReturnURI: strings.TrimSpace(logoutReturnURI),
	}
}

func (s *ResHubOIDCService) StartURL(ctx context.Context, state, verifier, nonce string) (string, error) {
	if s == nil || s.client == nil {
		return "", auth.ErrResHubUnavailable
	}
	if strings.TrimSpace(state) == "" || strings.TrimSpace(verifier) == "" || strings.TrimSpace(nonce) == "" {
		return "", auth.ErrResHubState
	}
	return s.client.AuthorizationURL(ctx, state, PKCEChallenge(verifier), nonce)
}

func (s *ResHubOIDCService) Complete(ctx context.Context, code, verifier, nonce string) (*auth.User, auth.ResHubSession, error) {
	if s == nil || s.client == nil || s.projection == nil {
		return nil, auth.ResHubSession{}, auth.ErrResHubUnavailable
	}
	session, err := s.client.Exchange(ctx, code, verifier, nonce, s.redirectURI)
	if err != nil {
		return nil, auth.ResHubSession{}, err
	}
	snapshot, err := s.client.UserInfo(ctx, session.AccessToken)
	if err != nil {
		return nil, auth.ResHubSession{}, err
	}
	snapshot.SessionID = session.SessionID
	snapshot.ExpiresAt = session.ExpiresAt
	user, err := s.projection.Upsert(ctx, snapshot)
	if err != nil {
		return nil, auth.ResHubSession{}, err
	}
	return user, session, nil
}

func (s *ResHubOIDCService) Logout(ctx context.Context, idTokenHint string) error {
	if s == nil || s.client == nil || strings.TrimSpace(idTokenHint) == "" {
		return nil
	}
	return s.client.Logout(ctx, idTokenHint, s.logoutReturnURI)
}

func NewOIDCState() (string, error) {
	return randomURLToken(32)
}

func NewOIDCNonce() (string, error) {
	return randomURLToken(32)
}

func NewPKCEVerifier() (string, error) {
	return randomURLToken(48)
}

func PKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func SafeReturnPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return "/"
	}
	u, err := url.Parse(value)
	if err != nil || u.IsAbs() || u.Host != "" || u.Path == "/login" {
		return "/"
	}
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	if u.Fragment != "" {
		path += "#" + u.EscapedFragment()
	}
	return path
}

func randomURLToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *ResHubOIDCService) CompleteIdentity(ctx context.Context, code, verifier, nonce string) (*auth.ResHubIdentity, auth.ResHubSession, error) {
	if s == nil || s.client == nil {
		return nil, auth.ResHubSession{}, auth.ErrResHubUnavailable
	}
	session, err := s.client.Exchange(ctx, code, verifier, nonce, s.redirectURI)
	if err != nil {
		return nil, session, err
	}
	identity, err := s.client.UserInfo(ctx, session.AccessToken)
	return identity, session, err
}

func (s *ResHubOIDCService) Identity(ctx context.Context, token string) (*auth.ResHubIdentity, error) {
	return s.client.UserInfo(ctx, token)
}
