//go:build test

package auth

import "context"

type MockBootstrapAccountRepository struct {
	User  *User
	Match bool
}

func (r *MockBootstrapAccountRepository) Ensure(context.Context, bool) (string, error) {
	return "", nil
}
func (r *MockBootstrapAccountRepository) Verify(context.Context, string) (bool, error) {
	return false, nil
}
func (r *MockBootstrapAccountRepository) Current(context.Context) (*User, error)      { return r.User, nil }
func (r *MockBootstrapAccountRepository) Bind(context.Context, *ResHubIdentity) error { return nil }
func (r *MockBootstrapAccountRepository) Matches(context.Context, string, string) (bool, error) {
	return r.Match, nil
}
