package reshub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"jcourse/config"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
)

const defaultPointsBasePath = "/api/integrations/v1"

// Client is the xk-side HTTP adapter for the first-party res-hub contract.
// It never exposes the opaque integration token to callers outside the
// backend and deliberately keeps endpoint construction in one package.
type Client struct {
	httpClient     *http.Client
	baseURL        string
	authorizeURL   string
	clientID       string
	secret         string
	redirectURI    string
	pointsBasePath string
}

func NewClient(conf config.ResHubConfig) *Client {
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
			Timeout: timeout,
			// Integration endpoints are private APIs. Do not follow a redirect
			// that could forward the client Basic credentials elsewhere.
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		baseURL:        strings.TrimRight(strings.TrimSpace(conf.InternalBaseURL), "/"),
		authorizeURL:   strings.TrimSpace(conf.AuthorizeURL),
		clientID:       strings.TrimSpace(conf.ClientID),
		secret:         conf.ClientSecret,
		redirectURI:    strings.TrimSpace(conf.RedirectURI),
		pointsBasePath: "/" + strings.Trim(pointsPath, "/"),
	}
}

func (c *Client) AuthorizeURL(state, codeChallenge string) (string, error) {
	if c == nil || c.authorizeURL == "" || c.clientID == "" || c.redirectURI == "" {
		return "", auth.ErrResHubUnavailable
	}
	u, err := url.Parse(c.authorizeURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", auth.ErrResHubUnavailable
	}
	setAuthorizeQuery := func(q url.Values) {
		q.Set("client_id", c.clientID)
		q.Set("redirect_uri", c.redirectURI)
		q.Set("response_type", "code")
		q.Set("state", state)
		q.Set("code_challenge", codeChallenge)
		q.Set("code_challenge_method", "S256")
	}
	if u.Fragment != "" {
		fragmentURL, fragmentErr := url.Parse(u.Fragment)
		if fragmentErr != nil || fragmentURL.Path == "" {
			return "", auth.ErrResHubUnavailable
		}
		fragmentQuery := fragmentURL.Query()
		setAuthorizeQuery(fragmentQuery)
		fragmentURL.RawQuery = fragmentQuery.Encode()
		u.Fragment = fragmentURL.String()
	} else {
		query := u.Query()
		setAuthorizeQuery(query)
		u.RawQuery = query.Encode()
	}
	return u.String(), nil
}

func (c *Client) Exchange(ctx context.Context, code, codeVerifier, redirectURI string) (auth.ResHubSession, error) {
	if c == nil || c.baseURL == "" || c.clientID == "" || c.secret == "" {
		return auth.ResHubSession{}, auth.ErrResHubUnavailable
	}
	if strings.TrimSpace(code) == "" || strings.TrimSpace(codeVerifier) == "" {
		return auth.ResHubSession{}, auth.ErrResHubState
	}
	if strings.TrimSpace(redirectURI) == "" {
		redirectURI = c.redirectURI
	}
	var response tokenResponse
	status, err := c.requestJSON(ctx, http.MethodPost, "/api/sso/v1/token", map[string]any{
		"client_id":     c.clientID,
		"code":          code,
		"redirect_uri":  redirectURI,
		"code_verifier": codeVerifier,
	}, &response)
	if err != nil {
		return auth.ResHubSession{}, c.mapAuthError(status, err)
	}
	token := firstNonEmpty(response.Token, response.AccessToken, response.SessionToken)
	if token == "" {
		return auth.ResHubSession{}, auth.ErrResHubState
	}
	return auth.ResHubSession{Token: token, ExpiresAt: response.ExpiresAt}, nil
}

func (c *Client) Introspect(ctx context.Context, token string) (*auth.ResHubIdentity, error) {
	if strings.TrimSpace(token) == "" {
		return nil, auth.ErrResHubSession
	}
	var response identityResponse
	status, err := c.requestJSON(ctx, http.MethodPost, "/api/sso/v1/introspect", map[string]any{
		"token": token,
	}, &response)
	if err != nil {
		return nil, c.mapAuthError(status, err)
	}
	if response.Active != nil && !*response.Active {
		return nil, auth.ErrResHubSession
	}
	identity := response.identity()
	if identity.Subject == "" {
		return nil, auth.ErrResHubSession
	}
	return identity, nil
}

func (c *Client) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	status, err := c.requestJSON(ctx, http.MethodPost, "/api/sso/v1/logout", map[string]any{
		"token": token,
	}, nil)
	if err != nil {
		return c.mapAuthError(status, err)
	}
	return nil
}

func (c *Client) Balance(ctx context.Context, subject string) (int, error) {
	if strings.TrimSpace(subject) == "" {
		return 0, auth.ErrIdentityNotFound
	}
	var response balanceResponse
	path := c.pointsBasePath + "/points/" + url.PathEscape(subject)
	status, err := c.requestJSON(ctx, http.MethodGet, path, nil, &response)
	if err != nil {
		return 0, c.mapPointError(status, err)
	}
	return response.value(), nil
}

func (c *Client) ListOperations(ctx context.Context, subject string, page, pageSize int) ([]point.ExternalOperation, int64, error) {
	if strings.TrimSpace(subject) == "" {
		return nil, 0, auth.ErrIdentityNotFound
	}
	q := url.Values{}
	q.Set("subject", subject)
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		q.Set("page_size", strconv.Itoa(pageSize))
	}
	var response operationsResponse
	status, err := c.requestJSON(ctx, http.MethodGet, c.pointsBasePath+"/points/ledger?"+q.Encode(), nil, &response)
	if err != nil {
		return nil, 0, c.mapPointError(status, err)
	}
	remoteItems := response.items()
	items := make([]point.ExternalOperation, len(remoteItems))
	for i := range remoteItems {
		items[i] = remoteItems[i].operation()
	}
	return items, response.total(), nil
}

func (c *Client) Change(ctx context.Context, subject string, amount int, reason, operationKey string) (point.ExternalOperation, error) {
	if strings.TrimSpace(subject) == "" || strings.TrimSpace(operationKey) == "" {
		return point.ExternalOperation{}, point.ErrOperationConflict
	}
	var response operationResponse
	status, err := c.requestJSON(ctx, http.MethodPost, c.pointsBasePath+"/points/operations", map[string]any{
		"subject":       subject,
		"amount":        amount,
		"reason":        reason,
		"operation_key": operationKey,
	}, &response)
	if err != nil {
		return point.ExternalOperation{}, c.mapPointError(status, err)
	}
	return response.operation(), nil
}

func (c *Client) FindOperation(ctx context.Context, subject, operationKey string) (*point.ExternalOperation, error) {
	if strings.TrimSpace(subject) == "" || strings.TrimSpace(operationKey) == "" {
		return nil, auth.ErrIdentityNotFound
	}
	var response operationResponse
	path := c.pointsBasePath + "/points/operations/" + url.PathEscape(operationKey) + "?subject=" + url.QueryEscape(subject)
	status, err := c.requestJSON(ctx, http.MethodGet, path, nil, &response)
	if err != nil {
		if status == http.StatusNotFound {
			return nil, nil
		}
		return nil, c.mapPointError(status, err)
	}
	op := response.operation()
	if op.ID == "" {
		return nil, nil
	}
	return &op, nil
}

func (c *Client) Reverse(ctx context.Context, subject, operationID string) (point.ExternalOperation, error) {
	if strings.TrimSpace(subject) == "" || strings.TrimSpace(operationID) == "" {
		return point.ExternalOperation{}, point.ErrOperationConflict
	}
	var response operationResponse
	path := c.pointsBasePath + "/points/operations/" + url.PathEscape(operationID) + "/reverse"
	status, err := c.requestJSON(ctx, http.MethodPost, path, nil, &response)
	if err != nil {
		return point.ExternalOperation{}, c.mapPointError(status, err)
	}
	return response.operation(), nil
}

func (c *Client) requestJSON(ctx context.Context, method, path string, body any, result any) (int, error) {
	if c == nil || c.httpClient == nil || c.baseURL == "" {
		return 0, fmt.Errorf("res-hub client is not configured")
	}
	endpoint, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return 0, err
	}
	if strings.Contains(path, "?") {
		// url.JoinPath preserves the query in normal cases, but explicitly keep
		// it here for paths containing escaped operation keys.
		if parsed, parseErr := url.Parse(c.baseURL + path); parseErr == nil {
			endpoint = parsed.String()
		}
	}

	var reader io.Reader
	if body != nil {
		payload, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return 0, marshalErr
		}
		reader = strings.NewReader(string(payload))
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.SetBasicAuth(c.clientID, c.secret)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	if err != nil {
		return res.StatusCode, err
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return res.StatusCode, remoteErrorFromBody(res.StatusCode, data)
	}
	if result == nil || len(data) == 0 {
		return res.StatusCode, nil
	}
	if err := decodeEnvelope(data, result); err != nil {
		return res.StatusCode, err
	}
	return res.StatusCode, nil
}

func (c *Client) mapAuthError(status int, err error) error {
	switch remoteCode(err) {
	case "integration_disabled", "integration_misconfigured", "invalid_client", "private_network_required":
		// These responses describe the xk/res-hub integration or network
		// configuration. Treating them as an end-user session expiry would
		// cause a misleading login loop.
		return auth.ErrResHubUnavailable
	case "invalid_grant", "invalid_code", "invalid_code_verifier", "invalid_redirect_uri", "invalid_code_challenge_method":
		return auth.ErrResHubState
	case "invalid_token", "invalid_session", "initial_password_required":
		return auth.ErrResHubSession
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return auth.ErrResHubSession
	}
	if status >= http.StatusInternalServerError || status == 0 {
		return auth.ErrResHubUnavailable
	}
	return auth.ErrResHubState
}

func (c *Client) mapPointError(status int, err error) error {
	switch remoteCode(err) {
	case "no_points_account", "no_point_account", "point_account_not_found", "user_not_found", "subject_not_found":
		return point.ErrNoPointAccount
	case "operation_not_found":
		return point.ErrOperationNotFound
	case "insufficient_points", "insufficient_balance":
		return point.ErrInsufficientBalance
	case "idempotency_conflict", "operation_conflict":
		return point.ErrOperationConflict
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return point.ErrPointUnavailable
	}
	if status == http.StatusNotFound {
		return point.ErrNoPointAccount
	}
	if status >= http.StatusInternalServerError || status == 0 {
		return point.ErrPointUnavailable
	}
	return err
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
		Data    struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	_ = json.Unmarshal(data, &payload)
	var detail struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	detailText := ""
	if len(payload.Detail) > 0 && string(payload.Detail) != "null" {
		if json.Unmarshal(payload.Detail, &detail) != nil {
			_ = json.Unmarshal(payload.Detail, &detailText)
		}
	}
	code := firstNonEmpty(payload.Code, detail.Code, payload.Data.Code, payload.Error)
	message := firstNonEmpty(payload.Message, detail.Message, detailText, payload.Error)
	return &remoteError{status: status, code: strings.ToLower(strings.TrimSpace(code)), message: message}
}

func remoteCode(err error) string {
	var remote *remoteError
	if !asRemoteError(err, &remote) {
		return ""
	}
	return remote.code
}

func asRemoteError(err error, target **remoteError) bool {
	if err == nil {
		return false
	}
	if remote, ok := err.(*remoteError); ok {
		*target = remote
		return true
	}
	return false
}

type tokenResponse struct {
	Token        string    `json:"token"`
	AccessToken  string    `json:"access_token"`
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type identityResponse struct {
	Active            *bool             `json:"active"`
	Subject           string            `json:"subject"`
	SubjectID         string            `json:"subject_id"`
	Username          string            `json:"username"`
	Nickname          string            `json:"nickname"`
	Email             string            `json:"email"`
	IdentityRole      string            `json:"identity_role"`
	EffectiveRole     string            `json:"effective_role"`
	Role              string            `json:"role"`
	PermissionMode    string            `json:"permission_mode"`
	ViolationCount    int               `json:"violation_count"`
	InitialPassword   bool              `json:"initial_password"`
	IsDefaultPassword bool              `json:"is_default_password"`
	SessionID         string            `json:"session_id"`
	ExpiresAt         time.Time         `json:"expires_at"`
	IdentityType      string            `json:"identity_type"`
	ImpersonatedBy    string            `json:"impersonated_by"`
	OperatorSubjectID string            `json:"operator_subject_id"`
	Session           identitySession   `json:"session"`
	NestedIdentity    *identityResponse `json:"identity"`
	HasPointsAccount  bool              `json:"has_points_account"`
	Operator          *identityOperator `json:"operator"`
}

func (r identityResponse) identity() *auth.ResHubIdentity {
	if r.Subject == "" && r.SubjectID == "" && r.NestedIdentity != nil {
		return r.NestedIdentity.identity()
	}
	subject := firstNonEmpty(r.Subject, r.SubjectID)
	identityRole := firstNonEmpty(r.IdentityRole, r.Role)
	sessionID := firstNonEmpty(r.SessionID, r.Session.ID)
	expiresAt := firstTime(r.ExpiresAt, r.Session.ExpiresAt.Time)
	impersonatedBy := firstNonEmpty(r.ImpersonatedBy, r.OperatorSubjectID)
	if impersonatedBy == "" && r.Operator != nil {
		impersonatedBy = r.Operator.Subject
	}
	return &auth.ResHubIdentity{
		Subject: subject, Username: firstNonEmpty(r.Username, r.Nickname), Email: r.Email,
		IdentityRole: identityRole, EffectiveRole: firstNonEmpty(r.EffectiveRole, r.Role),
		PermissionMode: r.PermissionMode,
		ViolationCount: r.ViolationCount, InitialPassword: r.InitialPassword || r.IsDefaultPassword, SessionID: sessionID,
		HasPointsAccount: r.HasPointsAccount,
		ExpiresAt:        expiresAt, IdentityType: r.IdentityType, ImpersonatedBy: impersonatedBy,
	}
}

type identitySession struct {
	ID        string     `json:"id"`
	ExpiresAt timeOrUnix `json:"expires_at"`
}

type identityOperator struct {
	Subject string `json:"subject"`
}

type balanceResponse struct {
	Balance int `json:"balance"`
	Points  int `json:"points"`
	Total   int `json:"total"`
}

func (r balanceResponse) value() int {
	if r.Balance != 0 || (r.Points == 0 && r.Total == 0) {
		return r.Balance
	}
	if r.Points != 0 {
		return r.Points
	}
	return r.Total
}

type operationsResponse struct {
	Items      []operationResponse `json:"items"`
	Records    []operationResponse `json:"records"`
	Operations []operationResponse `json:"operations"`
	Total      int64               `json:"total"`
}

func (r operationsResponse) items() []operationResponse {
	if len(r.Items) > 0 {
		return r.Items
	}
	if len(r.Records) > 0 {
		return r.Records
	}
	return r.Operations
}

func (r operationsResponse) total() int64 {
	if r.Total != 0 {
		return r.Total
	}
	return int64(len(r.items()))
}

func (r *operationsResponse) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) > 0 && data[0] == '[' {
		return json.Unmarshal(data, &r.Items)
	}
	type responseAlias operationsResponse
	var decoded responseAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*r = operationsResponse(decoded)
	return nil
}

type operationResponse struct {
	ID                  stringOrNumber `json:"id"`
	OperationID         stringOrNumber `json:"operation_id"`
	ClientID            string         `json:"client_id"`
	Subject             string         `json:"subject"`
	Amount              int            `json:"amount"`
	Reason              string         `json:"reason"`
	Description         string         `json:"description"`
	OperationKey        string         `json:"operation_key"`
	BalanceAfter        int            `json:"balance_after"`
	Balance             int            `json:"balance"`
	OccurredAt          timeOrUnix     `json:"occurred_at"`
	CreatedAt           timeOrUnix     `json:"created_at"`
	ReversalOf          stringOrNumber `json:"reversal_of"`
	OriginalID          stringOrNumber `json:"original_id"`
	OriginalOperationID stringOrNumber `json:"original_operation_id"`
}

func (r operationResponse) operation() point.ExternalOperation {
	return point.ExternalOperation{
		ID: firstNonEmpty(string(r.ID), string(r.OperationID)), ClientID: r.ClientID, Subject: r.Subject,
		Amount: r.Amount, Reason: r.Reason, Description: r.Description, OperationKey: r.OperationKey,
		BalanceAfter: firstInt(r.BalanceAfter, r.Balance), OccurredAt: firstTime(r.OccurredAt.Time, r.CreatedAt.Time),
		OriginalID: firstNonEmpty(string(r.ReversalOf), string(r.OriginalID), string(r.OriginalOperationID)),
	}
}

type timeOrUnix struct {
	time.Time
}

func (v *timeOrUnix) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		v.Time = time.Time{}
		return nil
	}
	if strings.HasPrefix(raw, "\"") {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		if strings.TrimSpace(value) == "" {
			v.Time = time.Time{}
			return nil
		}
		parsed, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return err
		}
		v.Time = parsed
		return nil
	}
	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}
	v.Time = time.Unix(seconds, 0).UTC()
	return nil
}

type stringOrNumber string

func (v *stringOrNumber) UnmarshalJSON(data []byte) error {
	value := strings.TrimSpace(string(data))
	if value == "" || value == "null" {
		*v = ""
		return nil
	}
	if strings.HasPrefix(value, "\"") {
		var decoded string
		if err := json.Unmarshal(data, &decoded); err != nil {
			return err
		}
		*v = stringOrNumber(decoded)
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return err
	}
	*v = stringOrNumber(number.String())
	return nil
}

func decodeEnvelope(data []byte, result any) error {
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	if len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		return json.Unmarshal(envelope.Data, result)
	}
	return json.Unmarshal(data, result)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func firstTime(values ...time.Time) time.Time {
	for _, value := range values {
		if !value.IsZero() {
			return value
		}
	}
	return time.Time{}
}

var _ auth.ResHubClient = (*Client)(nil)
var _ point.ExternalPointProvider = (*Client)(nil)
