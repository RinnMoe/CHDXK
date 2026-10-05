package auth

import "context"

// BootstrapAccountRepository keeps the local owner separate from SSO projections.
type BootstrapAccountRepository interface {
	Ensure(context.Context, bool) (string, error)
	Verify(context.Context, string) (bool, error)
	Current(context.Context) (*User, error)
	Bind(context.Context, *ResHubIdentity) error
	Matches(context.Context, string, string) (bool, error)
}
