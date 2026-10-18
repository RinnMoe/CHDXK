package reshub

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"jcourse/internal/domain/auth"
)

func (c *Client) Refresh(ctx context.Context, refreshToken string) (auth.ResHubSession, error) {
	if c == nil || c.clientID == "" || c.secret == "" {
		return auth.ResHubSession{}, auth.ErrResHubUnavailable
	}
	if strings.TrimSpace(refreshToken) == "" {
		return auth.ResHubSession{}, auth.ErrResHubSession
	}
	metadata, err := c.loadDiscovery(ctx)
	if err != nil {
		return auth.ResHubSession{}, err
	}
	status, data, err := c.postForm(ctx, metadata.TokenEndpoint, url.Values{
		"grant_type": {"refresh_token"}, "refresh_token": {refreshToken},
	}, true)
	if err != nil {
		return auth.ResHubSession{}, auth.ErrResHubRefreshUncertain
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		remote := remoteErrorFromBody(status, data)
		switch remoteCode(remote) {
		case "invalid_grant":
			return auth.ResHubSession{}, auth.ErrResHubSession
		case "invalid_client", "unauthorized_client", "unsupported_grant_type", "invalid_scope", "temporarily_unavailable":
			return auth.ResHubSession{}, auth.ErrResHubUnavailable
		}
		return auth.ResHubSession{}, auth.ErrResHubRefreshUncertain
	}
	var response tokenResponse
	if json.Unmarshal(data, &response) != nil || response.AccessToken == "" || response.RefreshToken == "" ||
		!strings.EqualFold(response.TokenType, "Bearer") || response.ExpiresIn <= 0 || response.ExpiresIn > 86400 {
		return auth.ResHubSession{}, auth.ErrResHubRefreshUncertain
	}
	return auth.ResHubSession{
		AccessToken: response.AccessToken, RefreshToken: response.RefreshToken,
		ExpiresAt: time.Now().Add(time.Duration(response.ExpiresIn) * time.Second),
	}, nil
}
