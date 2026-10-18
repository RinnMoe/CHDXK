package reshub

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"jcourse/config"
	"jcourse/internal/domain/auth"
)

func TestClientRefreshContractAndErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"rotated-without-id-token", 200, `{"access_token":"next-access","refresh_token":"next-refresh","token_type":"Bearer","expires_in":3600}`, nil},
		{"revoked", 400, `{"error":"invalid_grant"}`, auth.ErrResHubSession},
		{"client-not-enabled", 400, `{"error":"unauthorized_client"}`, auth.ErrResHubUnavailable},
		{"outage", 503, `{"error":"temporarily_unavailable"}`, auth.ErrResHubUnavailable},
		{"ambiguous-gateway", 502, `gateway error`, auth.ErrResHubRefreshUncertain},
		{"missing-rotation", 200, `{"access_token":"access","token_type":"Bearer","expires_in":3600}`, auth.ErrResHubRefreshUncertain},
		{"invalid-lifetime", 200, `{"access_token":"access","refresh_token":"refresh","token_type":"Bearer","expires_in":0}`, auth.ErrResHubRefreshUncertain},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				id, secret, ok := r.BasicAuth()
				if !ok || id != "test-client" || secret != "test-secret" || r.Method != "POST" {
					t.Error("client authentication contract violated")
				}
				if err := r.ParseForm(); err != nil {
					t.Fatal(err)
				}
				if r.PostForm.Get("grant_type") != "refresh_token" || r.PostForm.Get("refresh_token") != "old-refresh" || len(r.PostForm) != 2 {
					t.Error("unexpected refresh form")
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			client := &Client{httpClient: server.Client(), issuer: server.URL, clientID: "test-client", secret: "test-secret", discovery: &providerMetadata{TokenEndpoint: server.URL}, discoveryExpiresAt: time.Now().Add(time.Hour)}
			before := time.Now()
			got, err := client.Refresh(context.Background(), "old-refresh")
			if !errors.Is(err, tc.want) {
				t.Fatalf("error=%v want %v", err, tc.want)
			}
			if err == nil && (got.AccessToken != "next-access" || got.RefreshToken != "next-refresh" || got.IDToken != "" || got.ExpiresAt.Before(before.Add(3599*time.Second))) {
				t.Error("invalid refreshed credentials")
			}
		})
	}
}

// An explicitly supplied fixture must be a loopback-only, isolated provider.
// The fixture contains synthetic credentials and is never checked into Git.
func TestRefreshAgainstLocalResHub(t *testing.T) {
	path := os.Getenv("JCOURSE_TEST_OIDC_FIXTURE")
	if path == "" {
		t.Skip("JCOURSE_TEST_OIDC_FIXTURE is not set")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("cannot read local provider fixture")
	}
	var fixture struct {
		Issuer       string `json:"issuer"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RefreshToken string `json:"refresh_token"`
		Subject      string `json:"subject"`
	}
	if json.Unmarshal(data, &fixture) != nil {
		t.Fatal("invalid local provider fixture")
	}
	issuer, err := url.Parse(fixture.Issuer)
	if err != nil || (issuer.Hostname() != "127.0.0.1" && issuer.Hostname() != "localhost" && issuer.Hostname() != "::1") {
		t.Fatal("provider fixture must use loopback")
	}
	client, err := NewClient(config.ResHubConfig{Issuer: fixture.Issuer, ClientID: fixture.ClientID, ClientSecret: fixture.ClientSecret})
	if err != nil {
		t.Fatal("cannot configure local provider")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	first, err := client.Refresh(ctx, fixture.RefreshToken)
	if err != nil {
		t.Fatalf("first real-provider rotation failed: %v", err)
	}
	identity, err := client.UserInfo(ctx, first.AccessToken)
	if err != nil || identity == nil || (fixture.Subject != "" && identity.Subject != fixture.Subject) {
		t.Fatal("real-provider identity binding failed")
	}
	second, err := client.Refresh(ctx, first.RefreshToken)
	if err != nil || second.RefreshToken == first.RefreshToken {
		t.Fatal("real-provider second rotation failed")
	}
	if _, err := client.UserInfo(ctx, second.AccessToken); err != nil {
		t.Fatal("real-provider new access token rejected")
	}
	if _, err := client.Refresh(ctx, fixture.RefreshToken); !errors.Is(err, auth.ErrResHubSession) {
		t.Fatal("real-provider accepted consumed refresh token")
	}
	if _, err := client.UserInfo(ctx, second.AccessToken); !errors.Is(err, auth.ErrResHubSession) {
		t.Fatal("real-provider replay did not revoke rotation chain")
	}
}

func TestClientRefreshTransportFailureIsNotReplayable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.Close()
	client := &Client{httpClient: server.Client(), issuer: server.URL, clientID: "test-client", secret: "test-secret", discovery: &providerMetadata{TokenEndpoint: server.URL}, discoveryExpiresAt: time.Now().Add(time.Hour)}
	if _, err := client.Refresh(context.Background(), "old-refresh"); !errors.Is(err, auth.ErrResHubRefreshUncertain) {
		t.Fatalf("error=%v", err)
	}
}
