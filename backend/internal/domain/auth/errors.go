package auth

import "errors"

var (
	ErrUserSuspended      = errors.New("user is suspended")
	ErrApiKeyNameRequired = errors.New("api key name is required")
	ErrApiKeyNotFound     = errors.New("api key not found")
	ErrInvalidApiKeyRole  = errors.New("invalid api key role")
)
