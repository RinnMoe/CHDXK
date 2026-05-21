package auth

import "errors"

var (
	ErrEmailNotAllowed         = errors.New("email is not allowed")
	ErrVerificationTooSoon     = errors.New("verification code sent too recently")
	ErrVerificationCodeInvalid = errors.New("verification code is invalid")
	ErrInvalidCredentials      = errors.New("invalid email or password")
	ErrPasswordRequired        = errors.New("password is required")
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrUserSuspended           = errors.New("user is suspended")
)
