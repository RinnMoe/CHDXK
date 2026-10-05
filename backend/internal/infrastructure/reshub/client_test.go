package reshub

import (
	"context"
	"crypto"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"jcourse/config"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
)

func TestClientUsesStandardOIDCAndBearerPointAPI(t *testing.T) {
	privateKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}
	var serviceTokenRequests atomic.Int32
	var tokenRequestBody url.Values
	var logoutQuery url.Values

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			writeJSON(t, w, map[string]any{
				"issuer":                                server.URL,
				"authorization_endpoint":                server.URL + "/oidc/authorize",
				"token_endpoint":                        server.URL + "/oidc/token",
				"userinfo_endpoint":                     server.URL + "/oidc/userinfo",
				"jwks_uri":                              server.URL + "/oidc/jwks.json",
				"end_session_endpoint":                  server.URL + "/oidc/logout",
				"token_endpoint_auth_methods_supported": []string{"client_secret_basic"},
				"code_challenge_methods_supported":      []string{"S256"},
			})
		case "/oidc/jwks.json":
			writeJSON(t, w, jwksFor(privateKey))
		case "/oidc/token":
			user, password, ok := r.BasicAuth()
			if !ok || user != "xk" || password != "secret" {
				http.Error(w, `{"error":"invalid_client"}`, http.StatusUnauthorized)
				return
			}
			body, _ := io.ReadAll(r.Body)
			tokenRequestBody, _ = url.ParseQuery(string(body))
			if tokenRequestBody.Get("grant_type") == "authorization_code" {
				accessToken := "user-access-token"
				writeJSON(t, w, map[string]any{
					"access_token": accessToken,
					"token_type":   "Bearer",
					"expires_in":   3600,
					"scope":        "openid profile res_hub_identity",
					"id_token":     signedIDToken(t, privateKey, server.URL, "xk", "nonce-value", accessToken),
				})
				return
			}
			if tokenRequestBody.Get("grant_type") == "client_credentials" {
				serviceTokenRequests.Add(1)
				writeJSON(t, w, map[string]any{
					"access_token": "service-token",
					"token_type":   "Bearer",
					"expires_in":   3600,
					"scope":        tokenRequestBody.Get("scope"),
				})
				return
			}
			http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
		case "/oidc/userinfo":
			if r.Header.Get("Authorization") != "Bearer user-access-token" {
				http.Error(w, `{"error":"invalid_token"}`, http.StatusUnauthorized)
				return
			}
			writeJSON(t, w, map[string]any{
				"sub":                     "user:42",
				"name":                    "Ada",
				"preferred_username":      "ada",
				"res_hub_identity_type":   "member",
				"res_hub_role":            "member",
				"res_hub_effective_role":  "member",
				"res_hub_permission_mode": "member",
			})
		case "/oidc/logout":
			logoutQuery = r.URL.Query()
			w.Header().Set("Location", logoutQuery.Get("post_logout_redirect_uri"))
			w.WriteHeader(http.StatusFound)
		case "/api/integrations/v1/points/user:42":
			assertBearer(t, r)
			writeJSON(t, w, map[string]any{"subject": "user:42", "balance": 17, "has_points_account": true})
		case "/api/integrations/v1/points/ledger":
			assertBearer(t, r)
			writeJSON(t, w, map[string]any{
				"client_id": "xk", "subject": "user:42", "page": 2, "page_size": 10, "total": 1,
				"operations": []any{map[string]any{
					"operation_id": "op-1", "client_id": "xk", "subject": "user:42", "amount": 2,
					"reason": "review_create", "operation_key": "xk:key", "balance_after": 19,
					"created_at": 1790121600,
				}},
			})
		case "/api/integrations/v1/points/operations":
			assertBearer(t, r)
			if r.Method != http.MethodPost {
				http.Error(w, `{"detail":{"code":"invalid_request"}}`, http.StatusBadRequest)
				return
			}
			writeJSON(t, w, map[string]any{
				"operation_id": "op-1", "client_id": "xk", "subject": "user:42", "amount": 2,
				"reason": "review_create", "operation_key": "xk:key", "balance_after": 19,
				"created_at": 1790121600,
			})
		case "/api/integrations/v1/points/operations/xk:key":
			assertBearer(t, r)
			writeJSON(t, w, map[string]any{
				"operation_id": "op-1", "client_id": "xk", "subject": "user:42", "amount": 2,
				"reason": "review_create", "operation_key": "xk:key", "balance_after": 19,
				"created_at": 1790121600,
			})
		case "/api/integrations/v1/points/operations/op-1/reverse":
			assertBearer(t, r)
			writeJSON(t, w, map[string]any{
				"operation_id": "op-2", "client_id": "xk", "subject": "user:42", "amount": -2,
				"reason": "review_create", "operation_key": "xk:key:reverse:op-1", "balance_after": 17,
				"created_at": 1790121601, "reversal_of": "op-1",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, config.ResHubConfig{
		Issuer:          server.URL,
		ClientID:        "xk",
		ClientSecret:    "secret",
		RedirectURI:     "https://xk.example.test/api/auth/oidc/callback",
		LogoutReturnURI: "https://xk.example.test/",
	})

	authorizeURL, err := client.AuthorizationURL(context.Background(), "state-value", "challenge-value", "nonce-value")
	if err != nil {
		t.Fatalf("AuthorizationURL: %v", err)
	}
	parsed, err := url.Parse(authorizeURL)
	if err != nil {
		t.Fatalf("parse authorize URL: %v", err)
	}
	if parsed.Fragment != "" {
		t.Fatalf("authorization endpoint unexpectedly uses fragment: %s", authorizeURL)
	}
	for key, want := range map[string]string{
		"client_id":             "xk",
		"redirect_uri":          "https://xk.example.test/api/auth/oidc/callback",
		"response_type":         "code",
		"scope":                 "openid profile res_hub_identity",
		"state":                 "state-value",
		"nonce":                 "nonce-value",
		"code_challenge":        "challenge-value",
		"code_challenge_method": "S256",
	} {
		if got := parsed.Query().Get(key); got != want {
			t.Errorf("authorize query %s = %q, want %q", key, got, want)
		}
	}

	session, err := client.Exchange(context.Background(), "auth-code", "verifier", "nonce-value", "")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if session.AccessToken != "user-access-token" || session.IDToken == "" || session.SessionID != "session-1" || session.ExpiresAt.IsZero() {
		t.Fatalf("session = %#v", session)
	}
	if tokenRequestBody.Get("grant_type") != "authorization_code" || tokenRequestBody.Get("code") != "auth-code" || tokenRequestBody.Get("code_verifier") != "verifier" || tokenRequestBody.Get("redirect_uri") != "https://xk.example.test/api/auth/oidc/callback" {
		t.Fatalf("token body = %#v", tokenRequestBody)
	}

	identity, err := client.UserInfo(context.Background(), session.AccessToken)
	if err != nil {
		t.Fatalf("UserInfo: %v", err)
	}
	if identity.Issuer != server.URL || identity.Subject != "user:42" || identity.Username != "Ada" || identity.IdentityRole != "member" || !identity.HasPointsAccount {
		t.Fatalf("identity = %#v", identity)
	}
	if err := client.Logout(context.Background(), session.IDToken, "https://xk.example.test/"); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if logoutQuery.Get("id_token_hint") != session.IDToken || logoutQuery.Get("client_id") != "xk" || logoutQuery.Get("post_logout_redirect_uri") != "https://xk.example.test/" {
		t.Fatalf("logout query = %#v", logoutQuery)
	}

	balance, err := client.Balance(context.Background(), "user:42")
	if err != nil || balance != 17 {
		t.Fatalf("Balance = %d, %v", balance, err)
	}
	operations, total, err := client.ListOperations(context.Background(), "user:42", 2, 10)
	if err != nil || total != 1 || len(operations) != 1 || operations[0].ID != "op-1" || operations[0].OccurredAt.IsZero() {
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
	if serviceTokenRequests.Load() != 2 {
		t.Fatalf("client credentials requests = %d, want one per scope", serviceTokenRequests.Load())
	}
}

func TestClientMapsStandardPointErrors(t *testing.T) {
	privateKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			writeJSON(t, w, map[string]any{
				"issuer": server.URL, "token_endpoint": server.URL + "/oidc/token",
				"authorization_endpoint": server.URL + "/oidc/authorize", "userinfo_endpoint": server.URL + "/oidc/userinfo",
				"jwks_uri": server.URL + "/oidc/jwks.json", "end_session_endpoint": server.URL + "/oidc/logout",
				"token_endpoint_auth_methods_supported": []string{"client_secret_basic"},
				"code_challenge_methods_supported":      []string{"S256"},
			})
		case "/oidc/jwks.json":
			writeJSON(t, w, jwksFor(privateKey))
		case "/oidc/token":
			writeJSON(t, w, map[string]any{"access_token": "service-token", "token_type": "Bearer", "expires_in": 3600})
		case "/api/integrations/v1/points/operations":
			assertBearer(t, r)
			w.WriteHeader(http.StatusConflict)
			_, _ = io.WriteString(w, `{"detail":{"code":"insufficient_points","message":"insufficient"}}`)
		case "/api/integrations/v1/points/operations/op-missing/reverse":
			assertBearer(t, r)
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `{"detail":{"code":"operation_not_found","message":"missing"}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client := mustNewClient(t, config.ResHubConfig{Issuer: server.URL, ClientID: "xk", ClientSecret: "secret"})
	_, err = client.Change(context.Background(), "user:42", -2, "consume", "xk:key")
	if !errors.Is(err, point.ErrInsufficientBalance) {
		t.Fatalf("Change error = %v, want insufficient balance", err)
	}
	_, err = client.Reverse(context.Background(), "user:42", "op-missing")
	if !errors.Is(err, point.ErrOperationNotFound) {
		t.Fatalf("Reverse error = %v, want operation not found", err)
	}
}

func TestClientRejectsUnavailableDiscovery(t *testing.T) {
	client := mustNewClient(t, config.ResHubConfig{Issuer: "http://127.0.0.1:1", ClientID: "xk", ClientSecret: "secret"})
	_, err := client.UserInfo(context.Background(), "access-token")
	if !errors.Is(err, auth.ErrResHubUnavailable) {
		t.Fatalf("UserInfo unavailable error = %v", err)
	}
}

func assertBearer(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.Header.Get("Authorization"); got != "Bearer service-token" {
		t.Errorf("resource authorization = %q, want Bearer service-token", got)
	}
	if user, _, ok := r.BasicAuth(); ok || user != "" {
		t.Errorf("resource request unexpectedly used Basic auth")
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Errorf("write JSON: %v", err)
	}
}

func jwksFor(privateKey *rsa.PrivateKey) map[string]any {
	publicKey := &privateKey.PublicKey
	exponent := bigEndianBytes(publicKey.E)
	return map[string]any{"keys": []any{map[string]any{
		"kty": "RSA", "use": "sig", "alg": "RS256", "kid": "test-key",
		"n": base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
		"e": base64.RawURLEncoding.EncodeToString(exponent),
	}}}
}

func signedIDToken(t *testing.T, privateKey *rsa.PrivateKey, issuer, audience, nonce, accessToken string) string {
	t.Helper()
	encode := func(value any) string {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal token segment: %v", err)
		}
		return base64.RawURLEncoding.EncodeToString(data)
	}
	header := encode(map[string]string{"alg": "RS256", "kid": "test-key", "typ": "JWT"})
	now := time.Now().Unix()
	digest := sha256.Sum256([]byte(accessToken))
	claims := map[string]any{
		"iss": issuer, "sub": "user:42", "aud": audience,
		"exp": now + 3600, "iat": now, "auth_time": now,
		"nonce": nonce, "sid": "session-1",
		"at_hash": base64.RawURLEncoding.EncodeToString(digest[:16]),
	}
	payload := encode(claims)
	signingInput := header + "." + payload
	signatureDigest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(cryptorand.Reader, privateKey, crypto.SHA256, signatureDigest[:])
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func bigEndianBytes(value int) []byte {
	return new(big.Int).SetInt64(int64(value)).Bytes()
}

func mustNewClient(t testing.TB, conf config.ResHubConfig) *Client {
	t.Helper()
	client, err := NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestUserInfoPointsCapability(t *testing.T) {
	for _, tc := range []struct {
		subject    string
		capability any
		want       bool
	}{{"user:42", nil, true}, {"admin:root", nil, false}, {"user:42", false, false}, {"admin:root", true, true}} {
		t.Run(tc.subject, func(t *testing.T) {
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/.well-known/openid-configuration" {
					writeJSON(t, w, map[string]any{"issuer": server.URL, "authorization_endpoint": server.URL + "/authorize", "token_endpoint": server.URL + "/token", "userinfo_endpoint": server.URL + "/userinfo", "jwks_uri": server.URL + "/jwks", "end_session_endpoint": server.URL + "/logout", "token_endpoint_auth_methods_supported": []string{"client_secret_basic"}, "code_challenge_methods_supported": []string{"S256"}})
					return
				}
				body := map[string]any{"sub": tc.subject, "name": "积分用户"}
				if tc.capability != nil {
					body["res_hub_has_points_account"] = tc.capability
				}
				writeJSON(t, w, body)
			}))
			defer server.Close()
			client, err := NewClient(config.ResHubConfig{Issuer: server.URL, ClientID: "xk", ClientSecret: "test-secret"})
			if err != nil {
				t.Fatal(err)
			}
			user, err := client.UserInfo(context.Background(), "test-token")
			if err != nil {
				t.Fatal(err)
			}
			if user.HasPointsAccount != tc.want {
				t.Fatalf("capability = %v", user.HasPointsAccount)
			}
		})
	}
}
