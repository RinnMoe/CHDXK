//go:build test

package account

import (
	"context"
	"errors"
	"time"
)

type MockAccountRepository struct {
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

func NewMockAccountRepository(accounts map[string]*Account) *MockAccountRepository {
	repo := &MockAccountRepository{NextID: 1}
	repo.ensureMaps()
	for key, acct := range accounts {
		repo.PutAccount(key, acct)
	}
	return repo
}

func (r *MockAccountRepository) PutAccount(key string, acct *Account) {
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

func (r *MockAccountRepository) Create(ctx context.Context, acct *Account) error {
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

func (r *MockAccountRepository) Update(ctx context.Context, acct *Account) error {
	if r.OnUpdate != nil {
		return r.OnUpdate(ctx, acct)
	}
	r.PutAccount("", acct)
	return nil
}

func (r *MockAccountRepository) TouchLastSeen(ctx context.Context, userID int, at time.Time) error {
	if r.OnTouchLastSeen != nil {
		return r.OnTouchLastSeen(ctx, userID, at)
	}
	r.ensureMaps()
	r.TouchCount++
	acct, ok := r.AccountsByID[userID]
	if !ok {
		return errors.New("account not found")
	}
	acct.LastSeenAt = at
	return nil
}

func (r *MockAccountRepository) FindByID(ctx context.Context, id int) (*Account, error) {
	if r.OnFindByID != nil {
		return r.OnFindByID(ctx, id)
	}
	r.ensureMaps()
	return copyAccount(r.AccountsByID[id]), nil
}

func (r *MockAccountRepository) FindByUsername(ctx context.Context, username string) (*Account, error) {
	if r.OnFindByUsername != nil {
		return r.OnFindByUsername(ctx, username)
	}
	r.ensureMaps()
	return copyAccount(r.AccountsByUsername[username]), nil
}

func (r *MockAccountRepository) FindByEmail(ctx context.Context, email string) (*Account, error) {
	if r.OnFindByEmail != nil {
		return r.OnFindByEmail(ctx, email)
	}
	r.ensureMaps()
	return copyAccount(r.AccountsByEmail[email]), nil
}

func (r *MockAccountRepository) ensureMaps() {
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

type MockVerificationCodeRepository struct {
	Saved        map[string]VerificationCode
	CooldownTill map[string]time.Time

	OnReserveSend func(context.Context, string, time.Duration) (time.Duration, error)
	OnSave        func(context.Context, VerificationCode, time.Duration) error
	OnGet         func(context.Context, string) (*VerificationCode, error)
	OnDelete      func(context.Context, string) error
}

func NewMockVerificationCodeRepository() *MockVerificationCodeRepository {
	return &MockVerificationCodeRepository{Saved: map[string]VerificationCode{}, CooldownTill: map[string]time.Time{}}
}

func (r *MockVerificationCodeRepository) ReserveSend(ctx context.Context, email string, interval time.Duration) (time.Duration, error) {
	if r.OnReserveSend != nil {
		return r.OnReserveSend(ctx, email, interval)
	}
	r.ensureMaps()
	now := time.Now()
	if till := r.CooldownTill[email]; till.After(now) {
		return time.Until(till), nil
	}
	r.CooldownTill[email] = now.Add(interval)
	return 0, nil
}

func (r *MockVerificationCodeRepository) Save(ctx context.Context, code VerificationCode, ttl time.Duration) error {
	if r.OnSave != nil {
		return r.OnSave(ctx, code, ttl)
	}
	r.ensureMaps()
	r.Saved[code.Email] = code
	return nil
}

func (r *MockVerificationCodeRepository) Get(ctx context.Context, email string) (*VerificationCode, error) {
	if r.OnGet != nil {
		return r.OnGet(ctx, email)
	}
	r.ensureMaps()
	code, ok := r.Saved[email]
	if !ok {
		return nil, nil
	}
	return &code, nil
}

func (r *MockVerificationCodeRepository) Delete(ctx context.Context, email string) error {
	if r.OnDelete != nil {
		return r.OnDelete(ctx, email)
	}
	r.ensureMaps()
	delete(r.Saved, email)
	return nil
}

func (r *MockVerificationCodeRepository) ensureMaps() {
	if r.Saved == nil {
		r.Saved = map[string]VerificationCode{}
	}
	if r.CooldownTill == nil {
		r.CooldownTill = map[string]time.Time{}
	}
}

type MockLoginAttemptRepository struct {
	Counts map[string]int

	OnIncrement func(context.Context, string) (int, error)
	OnGet       func(context.Context, string) (int, error)
	OnReset     func(context.Context, string) error
}

func NewMockLoginAttemptRepository(counts map[string]int) *MockLoginAttemptRepository {
	if counts == nil {
		counts = map[string]int{}
	}
	return &MockLoginAttemptRepository{Counts: counts}
}

func (r *MockLoginAttemptRepository) Increment(ctx context.Context, email string) (int, error) {
	if r.OnIncrement != nil {
		return r.OnIncrement(ctx, email)
	}
	r.ensureMap()
	r.Counts[email]++
	return r.Counts[email], nil
}

func (r *MockLoginAttemptRepository) Get(ctx context.Context, email string) (int, error) {
	if r.OnGet != nil {
		return r.OnGet(ctx, email)
	}
	r.ensureMap()
	return r.Counts[email], nil
}

func (r *MockLoginAttemptRepository) Reset(ctx context.Context, email string) error {
	if r.OnReset != nil {
		return r.OnReset(ctx, email)
	}
	r.ensureMap()
	r.Counts[email] = 0
	return nil
}

func (r *MockLoginAttemptRepository) ensureMap() {
	if r.Counts == nil {
		r.Counts = map[string]int{}
	}
}
