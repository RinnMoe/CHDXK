package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/auth"
)

type fakeResHubOIDCClient struct {
	authorizeState     string
	authorizeChallenge string
	authorizeNonce     string
	exchangeCode       string
	exchangeVerifier   string
	exchangeNonce      string
	exchangeRedirect   string
	userinfoToken      string
	logoutIDToken      string
	logoutReturn       string
	identity           *auth.ResHubIdentity
}

func (f *fakeResHubOIDCClient) AuthorizationURL(_ context.Context, state, challenge, nonce string) (string, error) {
	f.authorizeState = state
	f.authorizeChallenge = challenge
	f.authorizeNonce = nonce
	return "https://hub.example.test/oidc/authorize", nil
}

func (f *fakeResHubOIDCClient) Exchange(_ context.Context, code, verifier, nonce, redirect string) (auth.ResHubSession, error) {
	f.exchangeCode = code
	f.exchangeVerifier = verifier
	f.exchangeNonce = nonce
	f.exchangeRedirect = redirect
	return auth.ResHubSession{
		AccessToken: "access-token",
		IDToken:     "id-token",
		SessionID:   "session-1",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil
}

func (f *fakeResHubOIDCClient) UserInfo(_ context.Context, accessToken string) (*auth.ResHubIdentity, error) {
	f.userinfoToken = accessToken
	return f.identity, nil
}

func (f *fakeResHubOIDCClient) ClientCredentialsToken(context.Context, []string) (string, error) {
	return "service-token", nil
}

func (f *fakeResHubOIDCClient) Logout(_ context.Context, idTokenHint, postLogoutRedirectURI string) error {
	f.logoutIDToken = idTokenHint
	f.logoutReturn = postLogoutRedirectURI
	return nil
}

type fakeIdentityProjection struct {
	identity *auth.ResHubIdentity
	user     *auth.User
}

func (f *fakeIdentityProjection) Upsert(_ context.Context, identity *auth.ResHubIdentity) (*auth.User, error) {
	f.identity = identity
	return f.user, nil
}

func (f *fakeIdentityProjection) SubjectByUserID(context.Context, int) (string, error) {
	return "user:1", nil
}

func TestResHubOIDCServiceCompleteAndLogout(t *testing.T) {
	client := &fakeResHubOIDCClient{identity: &auth.ResHubIdentity{Subject: "user:1"}}
	projection := &fakeIdentityProjection{user: &auth.User{ID: 7, Subject: "user:1"}}
	service := NewResHubOIDCService(
		client,
		projection,
		"https://xk.example.test/api/auth/oidc/callback",
		"https://xk.example.test/",
	)

	if _, err := service.StartURL(context.Background(), "state", "verifier", "nonce"); err != nil {
		t.Fatalf("StartURL: %v", err)
	}
	if client.authorizeState != "state" || client.authorizeChallenge != PKCEChallenge("verifier") || client.authorizeNonce != "nonce" {
		t.Fatalf("authorize arguments = %q, %q, %q", client.authorizeState, client.authorizeChallenge, client.authorizeNonce)
	}

	user, session, err := service.Complete(context.Background(), "code", "verifier", "nonce")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if user.ID != 7 || session.AccessToken != "access-token" || session.IDToken != "id-token" || client.userinfoToken != "access-token" {
		t.Fatalf("completion = user %#v session %#v client %#v", user, session, client)
	}
	if client.exchangeCode != "code" || client.exchangeVerifier != "verifier" || client.exchangeNonce != "nonce" || client.exchangeRedirect != "https://xk.example.test/api/auth/oidc/callback" {
		t.Fatalf("exchange arguments = %#v", client)
	}
	if projection.identity.Subject != "user:1" || projection.identity.SessionID != "session-1" || projection.identity.ExpiresAt.IsZero() {
		t.Fatalf("projection identity = %#v", projection.identity)
	}
	if err := service.Logout(context.Background(), session.IDToken); err != nil || client.logoutIDToken != session.IDToken || client.logoutReturn != "https://xk.example.test/" {
		t.Fatalf("Logout = %v client=%#v", err, client)
	}
}

func TestSafeReturnPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "/course/1?q=review#top", want: "/course/1?q=review#top"},
		{input: "", want: "/"},
		{input: "//attacker.example/", want: "/"},
		{input: "https://attacker.example/", want: "/"},
		{input: "/login", want: "/"},
		{input: "/login?redirect=/course", want: "/"},
	}
	for _, tt := range tests {
		if got := SafeReturnPath(tt.input); got != tt.want {
			t.Errorf("SafeReturnPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResHubOIDCServiceRequiresDependencies(t *testing.T) {
	service := NewResHubOIDCService(nil, nil, "", "")
	if _, _, err := service.Complete(context.Background(), "code", "verifier", "nonce"); !errors.Is(err, auth.ErrResHubUnavailable) {
		t.Fatalf("Complete error = %v", err)
	}
}
