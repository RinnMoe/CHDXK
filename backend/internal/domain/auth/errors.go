package auth

import "jcourse/pkg/apperr"

var (
	ErrUserSuspended      = apperr.ErrUserSuspended
	ErrApiKeyNameRequired = apperr.ErrApiKeyNameRequired
	ErrApiKeyNotFound     = apperr.ErrApiKeyNotFound
	ErrInvalidApiKey      = apperr.ErrInvalidApiKey
	ErrInvalidApiKeyRole  = apperr.ErrInvalidApiKeyRole
)
