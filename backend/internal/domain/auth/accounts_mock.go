//go:build test

package auth

import "context"

// MockAccountRepository supplies inert defaults; tests override relevant methods.
type MockAccountRepository struct {
	User       *User
	Credential *AccountCredential
	Accounts   []*User
	Match      bool
}

func (r *MockAccountRepository) Initialize(context.Context) (*InitialCredential, error) {
	return nil, nil
}
func (r *MockAccountRepository) FindCredential(context.Context, string) (*AccountCredential, error) {
	return r.Credential, nil
}
func (r *MockAccountRepository) FindByID(context.Context, int) (*User, error) { return r.User, nil }
func (r *MockAccountRepository) CreateLocal(context.Context, string, string, string, bool) (*User, error) {
	return r.User, nil
}
func (r *MockAccountRepository) SetPassword(context.Context, int, string, bool) error     { return nil }
func (r *MockAccountRepository) UpdateProfile(context.Context, int, string, string) error { return nil }
func (r *MockAccountRepository) ListAccounts(context.Context, AccountFilter) ([]*User, int64, error) {
	return r.Accounts, int64(len(r.Accounts)), nil
}
func (r *MockAccountRepository) SaveAccount(context.Context, *User) error { return nil }
func (r *MockAccountRepository) DeleteAccount(context.Context, int) error { return nil }
func (r *MockAccountRepository) ResHubAccounts(context.Context, *ResHubIdentity, bool) ([]*User, error) {
	if r.Accounts != nil {
		return r.Accounts, nil
	}
	if r.Match {
		return []*User{r.User}, nil
	}
	return []*User{}, nil
}
func (r *MockAccountRepository) BindAccount(context.Context, int, *ResHubIdentity) error { return nil }
func (r *MockAccountRepository) UnbindAccount(context.Context, int) error                { return nil }

var _ AccountRepository = (*MockAccountRepository)(nil)
