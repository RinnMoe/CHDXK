package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/auth"
)

type fakeResHubClient struct {
	authorizeState     string
	authorizeChallenge string
	exchangeCode       string
	exchangeVerifier   string
	exchangeRedirect   string
	logoutToken        string
	identity           *auth.ResHubIdentity
}

func (f *fakeResHubClient) AuthorizeURL(state, challenge string) (string, error) {
	f.authorizeState = state
	f.authorizeChallenge = challenge
	return "https://hub.example.test/sso/authorize", nil
}

func (f *fakeResHubClient) Exchange(_ context.Context, code, verifier, redirect string) (auth.ResHubSession, error) {
	f.exchangeCode = code
	f.exchangeVerifier = verifier
	f.exchangeRedirect = redirect
	return auth.ResHubSession{Token: "integration-token", ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func (f *fakeResHubClient) Introspect(context.Context, string) (*auth.ResHubIdentity, error) {
	return f.identity, nil
}

func (f *fakeResHubClient) Logout(_ context.Context, token string) error {
	f.logoutToken = token
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

func TestResHubSSOServiceCompleteAndLogout(t *testing.T) {
	client := &fakeResHubClient{identity: &auth.ResHubIdentity{Subject: "user:1"}}
	projection := &fakeIdentityProjection{user: &auth.User{ID: 7, Subject: "user:1"}}
	service := NewResHubSSOService(client, projection, "https://xk.example.test/api/auth/sso/callback")

	if _, err := service.StartURL("state", "verifier"); err != nil {
		t.Fatalf("StartURL: %v", err)
	}
	if client.authorizeState != "state" || client.authorizeChallenge != PKCEChallenge("verifier") {
		t.Fatalf("authorize arguments = %q, %q", client.authorizeState, client.authorizeChallenge)
	}

	user, token, err := service.Complete(context.Background(), "code", "verifier")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if user.ID != 7 || token != "integration-token" || client.exchangeCode != "code" || client.exchangeVerifier != "verifier" {
		t.Fatalf("completion = user %#v token %q exchange %#v", user, token, client)
	}
	if client.exchangeRedirect != "https://xk.example.test/api/auth/sso/callback" || projection.identity.Subject != "user:1" {
		t.Fatalf("completion context = %#v", client)
	}
	if err := service.Logout(context.Background(), token); err != nil || client.logoutToken != token {
		t.Fatalf("Logout = %v token=%q", err, client.logoutToken)
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

func TestResHubSSOServiceRequiresDependencies(t *testing.T) {
	service := NewResHubSSOService(nil, nil, "")
	if _, _, err := service.Complete(context.Background(), "code", "verifier"); !errors.Is(err, auth.ErrResHubUnavailable) {
		t.Fatalf("Complete error = %v", err)
	}
}
