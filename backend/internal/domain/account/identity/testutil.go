//go:build test

package identity

import (
	"context"
	"errors"
	"time"
)

type MockRepository struct {
	NextID             int
	AccountsByID       map[int]*Account
	AccountsByEmail    map[string]*Account
	AccountsByUsername map[string]*Account
	TouchCount         int

	OnCreate         func(context.Context, *Account) error
	OnUpdate         func(context.Context, *Account) error
	OnTouchLastSeen  func(context.Context, int, time.Time) error
	OnFindByID       func(context.Context, int) (*Account, error)
	OnFindByUsername func(context.Context, string) (*Account, error)
	OnFindByEmail    func(context.Context, string) (*Account, error)
	AfterCreate      func(*Account)
}

func NewMockRepository(accounts map[string]*Account) *MockRepository {
	repo := &MockRepository{NextID: 1}
	repo.ensureMaps()
	for key, acct := range accounts {
		repo.PutAccount(key, acct)
	}
	return repo
}

func (r *MockRepository) PutAccount(key string, acct *Account) {
	r.ensureMaps()
	copy := *acct
	if copy.ID >= r.NextID {
		r.NextID = copy.ID + 1
	}
	if copy.ID != 0 {
		r.AccountsByID[copy.ID] = &copy
	}
	if copy.Email != "" {
		r.AccountsByEmail[copy.Email] = &copy
	}
	if copy.Username != "" {
		r.AccountsByUsername[copy.Username] = &copy
	}
	if key != "" {
		r.AccountsByEmail[key] = &copy
		if copy.Username == "" {
			r.AccountsByUsername[key] = &copy
		}
	}
}

func (r *MockRepository) Create(ctx context.Context, acct *Account) error {
	if r.OnCreate != nil {
		return r.OnCreate(ctx, acct)
	}
	r.ensureMaps()
	if _, ok := r.AccountsByUsername[acct.Username]; ok {
		return errors.New("duplicate account")
	}
	copy := *acct
	if copy.ID == 0 {
		copy.ID = r.NextID
		r.NextID++
	}
	r.PutAccount("", &copy)
	acct.ID = copy.ID
	if r.AfterCreate != nil {
		r.AfterCreate(acct)
	}
	return nil
}

func (r *MockRepository) Update(ctx context.Context, acct *Account) error {
	if r.OnUpdate != nil {
		return r.OnUpdate(ctx, acct)
	}
	r.PutAccount("", acct)
	return nil
}

func (r *MockRepository) TouchLastSeen(ctx context.Context, accountID int, at time.Time) error {
	if r.OnTouchLastSeen != nil {
		return r.OnTouchLastSeen(ctx, accountID, at)
	}
	r.ensureMaps()
	r.TouchCount++
	acct, ok := r.AccountsByID[accountID]
	if !ok {
		return errors.New("account not found")
	}
	acct.LastSeenAt = at
	return nil
}

func (r *MockRepository) FindByID(ctx context.Context, id int) (*Account, error) {
	if r.OnFindByID != nil {
		return r.OnFindByID(ctx, id)
	}
	r.ensureMaps()
	return copyAccount(r.AccountsByID[id]), nil
}

func (r *MockRepository) FindByUsername(ctx context.Context, username string) (*Account, error) {
	if r.OnFindByUsername != nil {
		return r.OnFindByUsername(ctx, username)
	}
	r.ensureMaps()
	return copyAccount(r.AccountsByUsername[username]), nil
}

func (r *MockRepository) FindByEmail(ctx context.Context, email string) (*Account, error) {
	if r.OnFindByEmail != nil {
		return r.OnFindByEmail(ctx, email)
	}
	r.ensureMaps()
	return copyAccount(r.AccountsByEmail[email]), nil
}

func (r *MockRepository) ensureMaps() {
	if r.NextID == 0 {
		r.NextID = 1
	}
	if r.AccountsByID == nil {
		r.AccountsByID = map[int]*Account{}
	}
	if r.AccountsByEmail == nil {
		r.AccountsByEmail = map[string]*Account{}
	}
	if r.AccountsByUsername == nil {
		r.AccountsByUsername = map[string]*Account{}
	}
}

func copyAccount(acct *Account) *Account {
	if acct == nil {
		return nil
	}
	copy := *acct
	return &copy
}
