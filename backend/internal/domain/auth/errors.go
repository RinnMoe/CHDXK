package auth

import "jcourse/pkg/apperr"

var (
	ErrUserSuspended       = apperr.ErrUserSuspended
	ErrApiKeyNameRequired  = apperr.ErrApiKeyNameRequired
	ErrApiKeyLimitExceeded = apperr.ErrApiKeyLimitExceeded
	ErrApiKeyNotFound      = apperr.ErrApiKeyNotFound
	ErrInvalidApiKey       = apperr.ErrInvalidApiKey
	ErrInvalidApiKeyRole   = apperr.ErrInvalidApiKeyRole
)
