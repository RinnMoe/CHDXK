package auth

import "errors"

var (
	ErrUserSuspended       = errors.New("user is suspended")
	ErrApiKeyNameRequired  = errors.New("api key name is required")
	ErrApiKeyLimitExceeded = errors.New("api key limit exceeded")
	ErrApiKeyNotFound      = errors.New("api key not found")
	ErrInvalidApiKey       = errors.New("invalid api key")
	ErrInvalidApiKeyRole   = errors.New("invalid api key role")
)
