package reshub

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"jcourse/config"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
)

const (
	defaultPointsBasePath = "/api/integrations/v1"
	defaultDiscoveryTTL   = 5 * time.Minute
	defaultJWKSTTL        = 10 * time.Minute
	serviceTokenSkew      = 30 * time.Second
	userAuthScopes        = "openid profile res_hub_identity"
)

// Client is the xk-side OIDC Relying Party and first-party resource client.
// All provider endpoints are loaded from Discovery; only the standard OIDC
// and Bearer resource contracts are supported here.
type Client struct {
	httpClient      *http.Client
	issuer          string
	clientID        string
	secret          string
	redirectURI     string
	logoutReturnURI string
	pointsBasePath  string

	discoveryMu        sync.Mutex
	discovery          *providerMetadata
	discoveryExpiresAt time.Time

	jwksMu        sync.Mutex
	jwks          map[string]*rsa.PublicKey
	jwksExpiresAt time.Time

	serviceTokenMu sync.Mutex
	serviceTokens  map[string]cachedToken
}

type providerMetadata struct {
	Issuer                   string   `json:"issuer"`
	AuthorizationEndpoint    string   `json:"authorization_endpoint"`
	TokenEndpoint            string   `json:"token_endpoint"`
	UserInfoEndpoint         string   `json:"userinfo_endpoint"`
	JWKSURI                  string   `json:"jwks_uri"`
	EndSessionEndpoint       string   `json:"end_session_endpoint"`
	TokenEndpointAuthMethods []string `json:"token_endpoint_auth_methods_supported"`
	CodeChallengeMethods     []string `json:"code_challenge_methods_supported"`
}

type cachedToken struct {
	value     string
	expiresAt time.Time
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
}

type userInfoResponse struct {
	Subject           string `json:"sub"`
	HasPointsAccount  *bool  `json:"res_hub_has_points_account"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	IdentityType      string `json:"res_hub_identity_type"`
	Role              string `json:"res_hub_role"`
	EffectiveRole     string `json:"res_hub_effective_role"`
	PermissionMode    string `json:"res_hub_permission_mode"`
	OperatorSubject   string `json:"res_hub_operator_subject"`
}

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Algorithm string `json:"alg"`
	Use       string `json:"use"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

// NewClient creates a client for the current res-hub OIDC provider. The
// issuer is the only provider address configured locally; every other URL is
// obtained from its Discovery document.
func NewClient(conf config.ResHubConfig) (*Client, error) {
	transport, err := newTransport(conf)
	if err != nil {
		return nil, err
	}
	timeout := conf.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	pointsPath := strings.TrimSpace(conf.PointsBasePath)
	if pointsPath == "" {
		pointsPath = defaultPointsBasePath
	}
	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		issuer:          strings.TrimRight(strings.TrimSpace(conf.Issuer), "/"),
		clientID:        strings.TrimSpace(conf.ClientID),
		secret:          conf.ClientSecret,
		redirectURI:     strings.TrimSpace(conf.RedirectURI),
		logoutReturnURI: strings.TrimSpace(conf.LogoutReturnURI),
		pointsBasePath:  "/" + strings.Trim(pointsPath, "/"),
		serviceTokens:   make(map[string]cachedToken),
	}, nil
}

func (c *Client) AuthorizationURL(ctx context.Context, state, codeChallenge, nonce string) (string, error) {
	if c == nil || c.clientID == "" || c.redirectURI == "" || state == "" || codeChallenge == "" || nonce == "" {
		return "", auth.ErrResHubState
	}
	metadata, err := c.loadDiscovery(ctx)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(metadata.AuthorizationEndpoint)
	if err != nil || u.Fragment != "" {
		return "", auth.ErrResHubUnavailable
	}
	query := u.Query()
	query.Set("client_id", c.clientID)
	query.Set("redirect_uri", c.redirectURI)
	query.Set("response_type", "code")
	query.Set("scope", userAuthScopes)
	query.Set("state", state)
	query.Set("nonce", nonce)
	query.Set("code_challenge", codeChallenge)
	query.Set("code_challenge_method", "S256")
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func (c *Client) Exchange(ctx context.Context, code, codeVerifier, nonce, redirectURI string) (auth.ResHubSession, error) {
	if c == nil || c.clientID == "" || c.secret == "" {
		return auth.ResHubSession{}, auth.ErrResHubUnavailable
	}
	if strings.TrimSpace(code) == "" || strings.TrimSpace(codeVerifier) == "" || strings.TrimSpace(nonce) == "" {
		return auth.ResHubSession{}, auth.ErrResHubState
	}
	if strings.TrimSpace(redirectURI) == "" {
		redirectURI = c.redirectURI
	}
	metadata, err := c.loadDiscovery(ctx)
	if err != nil {
		return auth.ResHubSession{}, err
	}
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {codeVerifier},
	}
	status, data, err := c.postForm(ctx, metadata.TokenEndpoint, form, true)
	if err != nil {
		return auth.ResHubSession{}, auth.ErrResHubUnavailable
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return auth.ResHubSession{}, c.mapAuthError(status, remoteErrorFromBody(status, data))
	}
	var response tokenResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return auth.ResHubSession{}, auth.ErrResHubState
	}
	if response.AccessToken == "" || !strings.EqualFold(response.TokenType, "Bearer") || response.IDToken == "" {
		return auth.ResHubSession{}, auth.ErrResHubState
	}
	claims, err := c.verifyIDToken(ctx, response.IDToken, response.AccessToken, nonce)
	if err != nil {
		return auth.ResHubSession{}, err
	}
	expiresAt := unixClaim(claims, "exp")
	if response.ExpiresIn > 0 && response.ExpiresIn <= 86400 {
		accessExpiry := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
		if expiresAt.IsZero() || accessExpiry.Before(expiresAt) {
			expiresAt = accessExpiry
		}
	}
	return auth.ResHubSession{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		IDToken:      response.IDToken,
		SessionID:    stringClaim(claims, "sid"),
		ExpiresAt:    expiresAt,
	}, nil
}

func (c *Client) UserInfo(ctx context.Context, accessToken string) (*auth.ResHubIdentity, error) {
	if c == nil || strings.TrimSpace(accessToken) == "" {
		return nil, auth.ErrResHubSession
	}
	metadata, err := c.loadDiscovery(ctx)
	if err != nil {
		return nil, err
	}
	status, data, err := c.do(ctx, http.MethodGet, metadata.UserInfoEndpoint, nil, http.Header{
		"Accept":        {"application/json"},
		"Authorization": {"Bearer " + accessToken},
	})
	if err != nil {
		return nil, auth.ErrResHubUnavailable
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, c.mapAuthError(status, remoteErrorFromBody(status, data))
	}
	var response userInfoResponse
	if err := json.Unmarshal(data, &response); err != nil || strings.TrimSpace(response.Subject) == "" {
		return nil, auth.ErrResHubSession
	}
	username := strings.TrimSpace(response.Name)
	if username == "" {
		username = strings.TrimSpace(response.PreferredUsername)
	}
	hasPoints := strings.HasPrefix(response.Subject, "user:")
	if response.HasPointsAccount != nil {
		hasPoints = *response.HasPointsAccount
	}
	return &auth.ResHubIdentity{
		HasPointsAccount: hasPoints,
		Issuer:           c.issuer,
		Subject:          response.Subject,
		Username:         username,
		IdentityType:     response.IdentityType,
		IdentityRole:     response.Role,
		EffectiveRole:    response.EffectiveRole,
		PermissionMode:   response.PermissionMode,
		ImpersonatedBy:   response.OperatorSubject,
	}, nil
}

func (c *Client) ClientCredentialsToken(ctx context.Context, scopes []string) (string, error) {
	if c == nil || c.clientID == "" || c.secret == "" {
		return "", auth.ErrResHubUnavailable
	}
	scopes = normalizeScopes(scopes)
	if len(scopes) == 0 {
		return "", auth.ErrResHubState
	}
	cacheKey := strings.Join(scopes, " ")

	c.serviceTokenMu.Lock()
	defer c.serviceTokenMu.Unlock()
	if cached, ok := c.serviceTokens[cacheKey]; ok && time.Now().Add(serviceTokenSkew).Before(cached.expiresAt) {
		return cached.value, nil
	}

	metadata, err := c.loadDiscovery(ctx)
	if err != nil {
		return "", err
	}
	form := url.Values{
		"grant_type": {"client_credentials"},
		"scope":      {cacheKey},
	}
	status, data, err := c.postForm(ctx, metadata.TokenEndpoint, form, true)
	if err != nil {
		return "", auth.ErrResHubUnavailable
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return "", c.mapAuthError(status, remoteErrorFromBody(status, data))
	}
	var response tokenResponse
	if err := json.Unmarshal(data, &response); err != nil || response.AccessToken == "" || !strings.EqualFold(response.TokenType, "Bearer") {
		return "", auth.ErrResHubState
	}
	if response.ExpiresIn > 0 {
		c.serviceTokens[cacheKey] = cachedToken{
			value:     response.AccessToken,
			expiresAt: time.Now().Add(time.Duration(response.ExpiresIn) * time.Second),
		}
	}
	return response.AccessToken, nil
}

func (c *Client) Logout(ctx context.Context, idTokenHint, postLogoutRedirectURI string) error {
	if c == nil || strings.TrimSpace(idTokenHint) == "" {
		return nil
	}
	metadata, err := c.loadDiscovery(ctx)
	if err != nil {
		return err
	}
	u, err := url.Parse(metadata.EndSessionEndpoint)
	if err != nil {
		return auth.ErrResHubUnavailable
	}
	query := u.Query()
	query.Set("id_token_hint", idTokenHint)
	query.Set("client_id", c.clientID)
	if strings.TrimSpace(postLogoutRedirectURI) == "" {
		postLogoutRedirectURI = c.logoutReturnURI
	}
	if strings.TrimSpace(postLogoutRedirectURI) != "" {
		query.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	}
	u.RawQuery = query.Encode()
	status, data, err := c.do(ctx, http.MethodGet, u.String(), nil, http.Header{"Accept": {"application/json"}})
	if err != nil {
		return auth.ErrResHubUnavailable
	}
	if status < http.StatusOK || status >= http.StatusBadRequest {
		return c.mapAuthError(status, remoteErrorFromBody(status, data))
	}
	return nil
}

func (c *Client) Balance(ctx context.Context, subject string) (int, error) {
	if strings.TrimSpace(subject) == "" {
		return 0, auth.ErrIdentityNotFound
	}
	path := c.pointsBasePath + "/points/" + url.PathEscape(subject)
	status, data, err := c.requestPoint(ctx, http.MethodGet, path, []string{"points.read"}, nil)
	if err != nil {
		return 0, c.mapPointError(status, err)
	}
	var response struct {
		Balance int `json:"balance"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return 0, point.ErrPointUnavailable
	}
	return response.Balance, nil
}

func (c *Client) ListOperations(ctx context.Context, subject string, page, pageSize int) ([]point.ExternalOperation, int64, error) {
	if strings.TrimSpace(subject) == "" {
		return nil, 0, auth.ErrIdentityNotFound
	}
	query := url.Values{"subject": {subject}}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		query.Set("page_size", strconv.Itoa(pageSize))
	}
	path := c.pointsBasePath + "/points/ledger?" + query.Encode()
	status, data, err := c.requestPoint(ctx, http.MethodGet, path, []string{"points.read"}, nil)
	if err != nil {
		return nil, 0, c.mapPointError(status, err)
	}
	var response struct {
		Operations []operationResponse `json:"operations"`
		Total      int64               `json:"total"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, 0, point.ErrPointUnavailable
	}
	items := make([]point.ExternalOperation, len(response.Operations))
	for i := range response.Operations {
		items[i] = response.Operations[i].operation()
	}
	return items, response.Total, nil
}

func (c *Client) Change(ctx context.Context, subject string, amount int, reason, operationKey string) (point.ExternalOperation, error) {
	if strings.TrimSpace(subject) == "" || strings.TrimSpace(operationKey) == "" {
		return point.ExternalOperation{}, point.ErrOperationConflict
	}
	body := map[string]any{
		"subject":       subject,
		"amount":        amount,
		"reason":        reason,
		"operation_key": operationKey,
	}
	status, data, err := c.requestPoint(ctx, http.MethodPost, c.pointsBasePath+"/points/operations", []string{"points.write"}, body)
	if err != nil {
		return point.ExternalOperation{}, c.mapPointError(status, err)
	}
	var response operationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return point.ExternalOperation{}, point.ErrPointUnavailable
	}
	return response.operation(), nil
}

func (c *Client) FindOperation(ctx context.Context, subject, operationKey string) (*point.ExternalOperation, error) {
	if strings.TrimSpace(subject) == "" || strings.TrimSpace(operationKey) == "" {
		return nil, auth.ErrIdentityNotFound
	}
	path := c.pointsBasePath + "/points/operations/" + url.PathEscape(operationKey)
	status, data, err := c.requestPoint(ctx, http.MethodGet, path, []string{"points.read"}, nil)
	if err != nil {
		if status == http.StatusNotFound {
			return nil, nil
		}
		return nil, c.mapPointError(status, err)
	}
	var response operationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, point.ErrPointUnavailable
	}
	if response.OperationID == "" {
		return nil, nil
	}
	operation := response.operation()
	return &operation, nil
}

func (c *Client) Reverse(ctx context.Context, subject, operationID string) (point.ExternalOperation, error) {
	if strings.TrimSpace(subject) == "" || strings.TrimSpace(operationID) == "" {
		return point.ExternalOperation{}, point.ErrOperationConflict
	}
	path := c.pointsBasePath + "/points/operations/" + url.PathEscape(operationID) + "/reverse"
	status, data, err := c.requestPoint(ctx, http.MethodPost, path, []string{"points.write"}, nil)
	if err != nil {
		return point.ExternalOperation{}, c.mapPointError(status, err)
	}
	var response operationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return point.ExternalOperation{}, point.ErrPointUnavailable
	}
	return response.operation(), nil
}

func (c *Client) requestPoint(ctx context.Context, method, path string, scopes []string, body any) (int, []byte, error) {
	token, err := c.ClientCredentialsToken(ctx, scopes)
	if err != nil {
		return 0, nil, err
	}
	var reader io.Reader
	if body != nil {
		payload, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return 0, nil, marshalErr
		}
		reader = bytes.NewReader(payload)
	}
	status, data, err := c.do(ctx, method, c.resourceURL(path), reader, http.Header{
		"Accept":        {"application/json"},
		"Authorization": {"Bearer " + token},
		"Content-Type":  {"application/json"},
	})
	if err != nil {
		return 0, nil, point.ErrPointUnavailable
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return status, data, remoteErrorFromBody(status, data)
	}
	return status, data, nil
}

func (c *Client) postForm(ctx context.Context, endpoint string, form url.Values, basicAuth bool) (int, []byte, error) {
	headers := http.Header{
		"Accept":       {"application/json"},
		"Content-Type": {"application/x-www-form-urlencoded"},
	}
	if basicAuth {
		return c.do(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()), headersWithBasic(headers, c.clientID, c.secret))
	}
	return c.do(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()), headers)
}

func headersWithBasic(headers http.Header, username, password string) http.Header {
	if headers == nil {
		headers = make(http.Header)
	}
	headers.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	return headers
}

func (c *Client) do(ctx context.Context, method, endpoint string, body io.Reader, headers http.Header) (int, []byte, error) {
	if c == nil || c.httpClient == nil {
		return 0, nil, auth.ErrResHubUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return 0, nil, err
	}
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	response, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return response.StatusCode, nil, err
	}
	return response.StatusCode, data, nil
}

func (c *Client) loadDiscovery(ctx context.Context) (*providerMetadata, error) {
	if c == nil || c.issuer == "" {
		return nil, auth.ErrResHubUnavailable
	}
	c.discoveryMu.Lock()
	if c.discovery != nil && time.Now().Before(c.discoveryExpiresAt) {
		metadata := *c.discovery
		c.discoveryMu.Unlock()
		return &metadata, nil
	}
	c.discoveryMu.Unlock()

	endpoint := strings.TrimRight(c.issuer, "/") + "/.well-known/openid-configuration"
	status, data, err := c.do(ctx, http.MethodGet, endpoint, nil, http.Header{"Accept": {"application/json"}})
	if err != nil {
		return nil, auth.ErrResHubUnavailable
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, auth.ErrResHubUnavailable
	}
	var metadata providerMetadata
	if err := json.Unmarshal(data, &metadata); err != nil || strings.TrimRight(metadata.Issuer, "/") != c.issuer {
		return nil, auth.ErrResHubUnavailable
	}
	if !validEndpoint(metadata.AuthorizationEndpoint) || !validEndpoint(metadata.TokenEndpoint) ||
		!validEndpoint(metadata.UserInfoEndpoint) || !validEndpoint(metadata.JWKSURI) || !validEndpoint(metadata.EndSessionEndpoint) {
		return nil, auth.ErrResHubUnavailable
	}
	if !containsString(metadata.TokenEndpointAuthMethods, "client_secret_basic") || !containsString(metadata.CodeChallengeMethods, "S256") {
		return nil, auth.ErrResHubUnavailable
	}
	c.discoveryMu.Lock()
	c.discovery = &metadata
	c.discoveryExpiresAt = time.Now().Add(defaultDiscoveryTTL)
	c.discoveryMu.Unlock()
	return &metadata, nil
}

func (c *Client) loadJWKS(ctx context.Context, force bool) (map[string]*rsa.PublicKey, error) {
	if c == nil {
		return nil, auth.ErrResHubUnavailable
	}
	c.jwksMu.Lock()
	if !force && c.jwks != nil && time.Now().Before(c.jwksExpiresAt) {
		keys := c.jwks
		c.jwksMu.Unlock()
		return keys, nil
	}
	c.jwksMu.Unlock()
	metadata, err := c.loadDiscovery(ctx)
	if err != nil {
		return nil, err
	}
	status, data, err := c.do(ctx, http.MethodGet, metadata.JWKSURI, nil, http.Header{"Accept": {"application/json"}})
	if err != nil || status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, auth.ErrResHubUnavailable
	}
	var response jwksResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, auth.ErrResHubUnavailable
	}
	keys := make(map[string]*rsa.PublicKey, len(response.Keys))
	for _, item := range response.Keys {
		if item.KeyType != "RSA" || item.KeyID == "" || item.Modulus == "" || item.Exponent == "" ||
			(item.Algorithm != "" && item.Algorithm != "RS256") || (item.Use != "" && item.Use != "sig") {
			continue
		}
		modulus, err := decodeBase64URL(item.Modulus)
		if err != nil || len(modulus) == 0 {
			continue
		}
		exponentBytes, err := decodeBase64URL(item.Exponent)
		if err != nil || len(exponentBytes) == 0 || len(exponentBytes) > 4 {
			continue
		}
		exponent := 0
		for _, value := range exponentBytes {
			exponent = exponent<<8 | int(value)
		}
		if exponent <= 0 {
			continue
		}
		modulusInt := new(big.Int).SetBytes(modulus)
		if modulusInt.Sign() <= 0 {
			continue
		}
		keys[item.KeyID] = &rsa.PublicKey{N: modulusInt, E: exponent}
	}
	if len(keys) == 0 {
		return nil, auth.ErrResHubUnavailable
	}
	c.jwksMu.Lock()
	c.jwks = keys
	c.jwksExpiresAt = time.Now().Add(defaultJWKSTTL)
	c.jwksMu.Unlock()
	return keys, nil
}

func (c *Client) verifyIDToken(ctx context.Context, token, accessToken, nonce string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, auth.ErrResHubState
	}
	headerData, err := decodeBase64URL(parts[0])
	if err != nil {
		return nil, auth.ErrResHubState
	}
	var header struct {
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
	}
	if json.Unmarshal(headerData, &header) != nil || header.Algorithm != "RS256" || header.KeyID == "" {
		return nil, auth.ErrResHubState
	}
	claimData, err := decodeBase64URL(parts[1])
	if err != nil {
		return nil, auth.ErrResHubState
	}
	var claims map[string]any
	if json.Unmarshal(claimData, &claims) != nil {
		return nil, auth.ErrResHubState
	}
	keys, err := c.loadJWKS(ctx, false)
	if err != nil {
		return nil, err
	}
	key, ok := keys[header.KeyID]
	if !ok {
		keys, err = c.loadJWKS(ctx, true)
		if err != nil {
			return nil, err
		}
		key = keys[header.KeyID]
		if key == nil {
			return nil, auth.ErrResHubState
		}
	}
	signature, err := decodeBase64URL(parts[2])
	if err != nil {
		return nil, auth.ErrResHubState
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature); err != nil {
		return nil, auth.ErrResHubState
	}
	if strings.TrimRight(stringClaim(claims, "iss"), "/") != c.issuer || stringClaim(claims, "sub") == "" ||
		!audienceContains(claims["aud"], c.clientID) || stringClaim(claims, "nonce") != nonce ||
		stringClaim(claims, "sid") == "" {
		return nil, auth.ErrResHubState
	}
	exp := unixClaim(claims, "exp")
	if exp.IsZero() || !exp.After(time.Now()) {
		return nil, auth.ErrResHubSession
	}
	if stringClaim(claims, "at_hash") == "" || stringClaim(claims, "at_hash") != accessTokenHash(accessToken) {
		return nil, auth.ErrResHubState
	}
	return claims, nil
}

func (c *Client) resourceURL(path string) string {
	return strings.TrimRight(c.issuer, "/") + "/" + strings.TrimLeft(path, "/")
}

func (c *Client) mapAuthError(status int, err error) error {
	code := remoteCode(err)
	switch code {
	case "invalid_client", "unauthorized_client", "temporarily_unavailable", "server_error":
		return auth.ErrResHubUnavailable
	case "invalid_grant", "invalid_request", "invalid_scope", "invalid_code", "invalid_code_verifier", "invalid_redirect_uri", "invalid_code_challenge_method":
		return auth.ErrResHubState
	case "invalid_token", "invalid_session", "login_required", "access_denied":
		return auth.ErrResHubSession
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return auth.ErrResHubSession
	}
	if status == 0 || status >= http.StatusInternalServerError {
		return auth.ErrResHubUnavailable
	}
	return auth.ErrResHubState
}

func (c *Client) mapPointError(status int, err error) error {
	switch remoteCode(err) {
	case "no_points_account", "no_point_account", "point_account_not_found":
		return point.ErrNoPointAccount
	case "subject_not_found":
		return auth.ErrIdentityNotFound
	case "operation_not_found":
		return point.ErrOperationNotFound
	case "insufficient_points", "insufficient_balance":
		return point.ErrInsufficientBalance
	case "operation_conflict", "idempotency_conflict":
		return point.ErrOperationConflict
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden || status == 0 || status >= http.StatusInternalServerError {
		return point.ErrPointUnavailable
	}
	if status == http.StatusNotFound {
		return point.ErrNoPointAccount
	}
	return point.ErrPointUnavailable
}

type remoteError struct {
	status  int
	code    string
	message string
}

func (e *remoteError) Error() string {
	if e.message != "" {
		return e.message
	}
	return fmt.Sprintf("res-hub request failed with status %d", e.status)
}

func remoteErrorFromBody(status int, data []byte) error {
	var payload struct {
		Code    string          `json:"code"`
		Error   string          `json:"error"`
		Message string          `json:"message"`
		Detail  json.RawMessage `json:"detail"`
	}
	_ = json.Unmarshal(data, &payload)
	code := strings.TrimSpace(payload.Code)
	message := strings.TrimSpace(payload.Message)
	if len(payload.Detail) > 0 && string(payload.Detail) != "null" {
		var detail struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(payload.Detail, &detail) == nil {
			if code == "" {
				code = detail.Code
			}
			if message == "" {
				message = detail.Message
			}
		} else {
			var detailText string
			if json.Unmarshal(payload.Detail, &detailText) == nil && message == "" {
				message = detailText
			}
		}
	}
	if code == "" {
		code = payload.Error
	}
	if message == "" {
		message = payload.Error
	}
	return &remoteError{status: status, code: strings.ToLower(strings.TrimSpace(code)), message: message}
}

func remoteCode(err error) string {
	var remote *remoteError
	if errors.As(err, &remote) {
		return remote.code
	}
	return ""
}

type operationResponse struct {
	OperationID  string     `json:"operation_id"`
	ClientID     string     `json:"client_id"`
	Subject      string     `json:"subject"`
	Amount       int        `json:"amount"`
	Reason       string     `json:"reason"`
	OperationKey string     `json:"operation_key"`
	BalanceAfter int        `json:"balance_after"`
	CreatedAt    timeOrUnix `json:"created_at"`
	ReversalOf   string     `json:"reversal_of"`
}

func (r operationResponse) operation() point.ExternalOperation {
	return point.ExternalOperation{
		ID:           r.OperationID,
		ClientID:     r.ClientID,
		Subject:      r.Subject,
		Amount:       r.Amount,
		Reason:       r.Reason,
		OperationKey: r.OperationKey,
		BalanceAfter: r.BalanceAfter,
		OccurredAt:   r.CreatedAt.Time,
		OriginalID:   r.ReversalOf,
	}
}

func normalizeScopes(scopes []string) []string {
	set := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		if scope = strings.TrimSpace(scope); scope != "" {
			set[scope] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for scope := range set {
		result = append(result, scope)
	}
	sort.Strings(result)
	return result
}

func validEndpoint(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && u.Fragment == ""
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func decodeBase64URL(value string) ([]byte, error) {
	value = strings.TrimRight(strings.TrimSpace(value), "=")
	return base64.RawURLEncoding.DecodeString(value)
}

func accessTokenHash(accessToken string) string {
	digest := sha256.Sum256([]byte(accessToken))
	return base64.RawURLEncoding.EncodeToString(digest[:len(digest)/2])
}

func stringClaim(claims map[string]any, name string) string {
	value, _ := claims[name].(string)
	return strings.TrimSpace(value)
}

func unixClaim(claims map[string]any, name string) time.Time {
	value, ok := claims[name].(float64)
	if !ok || value <= 0 {
		return time.Time{}
	}
	return time.Unix(int64(value), 0)
}

func audienceContains(value any, expected string) bool {
	switch audience := value.(type) {
	case string:
		return audience == expected
	case []any:
		for _, item := range audience {
			if item, ok := item.(string); ok && item == expected {
				return true
			}
		}
	}
	return false
}

type timeOrUnix struct {
	time.Time
}

func (value *timeOrUnix) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		value.Time = time.Time{}
		return nil
	}
	if strings.HasPrefix(raw, "\"") {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		if strings.TrimSpace(text) == "" {
			value.Time = time.Time{}
			return nil
		}
		parsed, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			return err
		}
		value.Time = parsed
		return nil
	}
	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}
	value.Time = time.Unix(seconds, 0).UTC()
	return nil
}

var _ auth.ResHubOIDCClient = (*Client)(nil)
var _ point.ExternalPointProvider = (*Client)(nil)
