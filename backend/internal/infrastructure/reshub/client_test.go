package reshub

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"jcourse/config"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
)

func TestClientAuthorizeURLAndSSORequests(t *testing.T) {
	var tokenBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hasExpectedSecret(r) {
			t.Errorf("missing client authentication")
		}
		if r.URL.Path == "/api/sso/v1/token" {
			if err := json.NewDecoder(r.Body).Decode(&tokenBody); err != nil {
				t.Errorf("decode token body: %v", err)
			}
			_, _ = io.WriteString(w, `{"access_token":"session-token","expires_at":"2030-01-02T03:04:05Z"}`)
			return
		}
		if r.URL.Path == "/api/sso/v1/introspect" {
			_, _ = io.WriteString(w, `{"active":true,"subject":"user:42","subject_id":"user:42","nickname":"Ada","role":"admin","effective_role":"admin","permission_mode":"admin","has_points_account":true,"identity_type":"admin","session":{"id":"sess-1","expires_at":1893563045}}`)
			return
		}
		if r.URL.Path == "/api/sso/v1/logout" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := NewClient(config.ResHubConfig{
		AuthorizeURL:    server.URL + "/#/sso/authorize",
		InternalBaseURL: server.URL,
		ClientID:        "xk",
		ClientSecret:    "secret",
		RedirectURI:     "https://xk.example.test/api/auth/sso/callback",
	})
	authorizeURL, err := client.AuthorizeURL("state-value", "challenge-value")
	if err != nil {
		t.Fatalf("AuthorizeURL: %v", err)
	}
	parsed, err := url.Parse(authorizeURL)
	if err != nil {
		t.Fatalf("parse authorize URL: %v", err)
	}
	if !strings.Contains(authorizeURL, "redirect_uri=https%3A%2F%2Fxk.example.test%2Fapi%2Fauth%2Fsso%2Fcallback") || strings.Contains(authorizeURL, "redirect_uri=https%25") {
		t.Fatalf("authorize URL encoded redirect_uri incorrectly: %s", authorizeURL)
	}
	authorizeQuery := parsed.Query()
	if parsed.Fragment != "" {
		fragment, fragmentErr := url.Parse(parsed.Fragment)
		if fragmentErr != nil {
			t.Fatalf("parse authorize fragment: %v", fragmentErr)
		}
		if fragment.Path != "/sso/authorize" || parsed.RawQuery != "" {
			t.Fatalf("authorize URL route/query = %q / %q", fragment.Path, parsed.RawQuery)
		}
		authorizeQuery = fragment.Query()
	}
	for key, want := range map[string]string{
		"client_id":             "xk",
		"redirect_uri":          "https://xk.example.test/api/auth/sso/callback",
		"response_type":         "code",
		"state":                 "state-value",
		"code_challenge":        "challenge-value",
		"code_challenge_method": "S256",
	} {
		if got := authorizeQuery.Get(key); got != want {
			t.Errorf("authorize query %s = %q, want %q", key, got, want)
		}
	}

	session, err := client.Exchange(context.Background(), "auth-code", "verifier", "")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if session.Token != "session-token" || session.ExpiresAt.IsZero() {
		t.Fatalf("session = %#v", session)
	}
	if tokenBody["client_id"] != "xk" || tokenBody["code"] != "auth-code" || tokenBody["code_verifier"] != "verifier" {
		t.Fatalf("token body = %#v", tokenBody)
	}

	identity, err := client.Introspect(context.Background(), session.Token)
	if err != nil {
		t.Fatalf("Introspect: %v", err)
	}
	if identity.Subject != "user:42" || identity.Username != "Ada" || identity.IdentityRole != "admin" || identity.EffectiveRole != "admin" || identity.PermissionMode != "admin" || !identity.HasPointsAccount || identity.SessionID != "sess-1" || identity.ExpiresAt.Unix() != 1893563045 {
		t.Fatalf("identity = %#v", identity)
	}
	if err := client.Logout(context.Background(), session.Token); err != nil {
		t.Fatalf("Logout: %v", err)
	}
}

func TestClientPointOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/integrations/v1/points/user:42":
			_, _ = io.WriteString(w, `{"subject":"user:42","balance":17,"has_points_account":true}`)
		case r.URL.Path == "/api/integrations/v1/points/ledger":
			if r.Method != http.MethodGet {
				t.Errorf("ledger method = %s", r.Method)
			}
			_, _ = io.WriteString(w, `{"client_id":"xk","subject":"user:42","operations":[{"operation_id":"op-1","subject":"user:42","amount":2,"reason":"review_create","operation_key":"xk:key","balance_after":19,"created_at":1700000000}],"page":2,"page_size":10,"total":1}`)
		case r.URL.Path == "/api/integrations/v1/points/operations" && r.Method == http.MethodPost:
			_, _ = io.WriteString(w, `{"operation_id":"op-1","client_id":"xk","subject":"user:42","amount":2,"reason":"review_create","operation_key":"xk:key","balance_after":19,"created_at":1700000000}`)
		case r.URL.Path == "/api/integrations/v1/points/operations/op-1/reverse":
			_, _ = io.WriteString(w, `{"operation_id":"op-2","client_id":"xk","subject":"user:42","amount":-2,"reason":"review_create","operation_key":"xk:key:reverse:op-1","balance_after":17,"created_at":1700000001,"reversal_of":"op-1"}`)
		case r.URL.Path == "/api/integrations/v1/points/operations/xk:key":
			_, _ = io.WriteString(w, `{"operation_id":"op-1","client_id":"xk","subject":"user:42","amount":2,"reason":"review_create","operation_key":"xk:key","balance_after":19,"created_at":1700000000}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client := NewClient(config.ResHubConfig{InternalBaseURL: server.URL, ClientID: "xk", ClientSecret: "secret"})

	balance, err := client.Balance(context.Background(), "user:42")
	if err != nil || balance != 17 {
		t.Fatalf("Balance = %d, %v", balance, err)
	}
	operations, total, err := client.ListOperations(context.Background(), "user:42", 2, 10)
	if err != nil || total != 1 || len(operations) != 1 || operations[0].ID != "op-1" || operations[0].Reason != "review_create" || operations[0].OccurredAt.Unix() != 1700000000 {
		t.Fatalf("ListOperations = %#v, total=%d, err=%v", operations, total, err)
	}
	operation, err := client.Change(context.Background(), "user:42", 2, "review_create", "xk:key")
	if err != nil || operation.ID != "op-1" {
		t.Fatalf("Change = %#v, %v", operation, err)
	}
	found, err := client.FindOperation(context.Background(), "user:42", "xk:key")
	if err != nil || found == nil || found.ID != "op-1" {
		t.Fatalf("FindOperation = %#v, %v", found, err)
	}
	reversal, err := client.Reverse(context.Background(), "user:42", "op-1")
	if err != nil || reversal.OriginalID != "op-1" {
		t.Fatalf("Reverse = %#v, %v", reversal, err)
	}
}

func TestClientMapsPointErrorsAndUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"detail":{"code":"insufficient_points","message":"insufficient"}}`)
	}))
	defer server.Close()
	client := NewClient(config.ResHubConfig{InternalBaseURL: server.URL, ClientID: "xk", ClientSecret: "secret"})
	_, err := client.Change(context.Background(), "user:42", -2, "consume", "xk:key")
	if !errors.Is(err, point.ErrInsufficientBalance) {
		t.Fatalf("Change error = %v, want insufficient balance", err)
	}

	unavailable := NewClient(config.ResHubConfig{InternalBaseURL: "http://127.0.0.1:1", ClientID: "xk", ClientSecret: "secret"})
	_, err = unavailable.Introspect(context.Background(), "token")
	if !errors.Is(err, auth.ErrResHubUnavailable) {
		t.Fatalf("Introspect unavailable error = %v", err)
	}
}

func TestClientMapsIntegrationConfigurationErrorsAsUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"detail":{"code":"private_network_required","message":"private network required"}}`)
	}))
	defer server.Close()

	client := NewClient(config.ResHubConfig{InternalBaseURL: server.URL, ClientID: "xk", ClientSecret: "secret"})
	_, err := client.Introspect(context.Background(), "token")
	if !errors.Is(err, auth.ErrResHubUnavailable) {
		t.Fatalf("Introspect error = %v, want unavailable", err)
	}
}

func TestClientMapsMissingOperation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"detail":{"code":"operation_not_found","message":"operation not found"}}`)
	}))
	defer server.Close()

	client := NewClient(config.ResHubConfig{InternalBaseURL: server.URL, ClientID: "xk", ClientSecret: "secret"})
	_, err := client.Reverse(context.Background(), "user:42", "op-missing")
	if !errors.Is(err, point.ErrOperationNotFound) {
		t.Fatalf("Reverse error = %v, want operation not found", err)
	}
}

func TestClientRejectsInactiveIntrospectionResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"active":false,"subject":"user:42"}`)
	}))
	defer server.Close()

	client := NewClient(config.ResHubConfig{InternalBaseURL: server.URL, ClientID: "xk", ClientSecret: "secret"})
	_, err := client.Introspect(context.Background(), "token")
	if !errors.Is(err, auth.ErrResHubSession) {
		t.Fatalf("Introspect error = %v, want session error", err)
	}
}

func hasExpectedSecret(r *http.Request) bool {
	user, password, ok := r.BasicAuth()
	return ok && user == "xk" && password == "secret"
}
