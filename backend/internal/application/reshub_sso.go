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

type ResHubSSOService struct {
	client      auth.ResHubAuthorizeClient
	projection  auth.IdentityProjectionRepository
	redirectURI string
}

func NewResHubSSOService(client auth.ResHubAuthorizeClient, projection auth.IdentityProjectionRepository, redirectURI string) *ResHubSSOService {
	return &ResHubSSOService{client: client, projection: projection, redirectURI: strings.TrimSpace(redirectURI)}
}

func (s *ResHubSSOService) StartURL(state, verifier string) (string, error) {
	if s == nil || s.client == nil {
		return "", auth.ErrResHubUnavailable
	}
	if strings.TrimSpace(state) == "" || strings.TrimSpace(verifier) == "" {
		return "", auth.ErrResHubState
	}
	return s.client.AuthorizeURL(state, PKCEChallenge(verifier))
}

func (s *ResHubSSOService) Complete(ctx context.Context, code, verifier string) (*auth.User, string, error) {
	if s == nil || s.client == nil || s.projection == nil {
		return nil, "", auth.ErrResHubUnavailable
	}
	session, err := s.client.Exchange(ctx, code, verifier, s.redirectURI)
	if err != nil {
		return nil, "", err
	}
	snapshot, err := s.client.Introspect(ctx, session.Token)
	if err != nil {
		return nil, "", err
	}
	user, err := s.projection.Upsert(ctx, snapshot)
	if err != nil {
		return nil, "", err
	}
	return user, session.Token, nil
}

func (s *ResHubSSOService) Logout(ctx context.Context, token string) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Logout(ctx, token)
}

func NewSSOState() (string, error) {
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
