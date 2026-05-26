//go:build test

package verification

import (
	"context"
	"time"
)

type MockCodeRepository struct {
	Saved        map[string]Code
	CooldownTill map[string]time.Time

	OnReserveSend func(context.Context, string, time.Duration) (time.Duration, error)
	OnSave        func(context.Context, Code, time.Duration) error
	OnGet         func(context.Context, string) (*Code, error)
	OnDelete      func(context.Context, string) error
}

func NewMockCodeRepository() *MockCodeRepository {
	return &MockCodeRepository{Saved: map[string]Code{}, CooldownTill: map[string]time.Time{}}
}

func (r *MockCodeRepository) ReserveSend(ctx context.Context, email string, interval time.Duration) (time.Duration, error) {
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

func (r *MockCodeRepository) Save(ctx context.Context, code Code, ttl time.Duration) error {
	if r.OnSave != nil {
		return r.OnSave(ctx, code, ttl)
	}
	r.ensureMaps()
	r.Saved[code.Email] = code
	return nil
}

func (r *MockCodeRepository) Get(ctx context.Context, email string) (*Code, error) {
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

func (r *MockCodeRepository) Delete(ctx context.Context, email string) error {
	if r.OnDelete != nil {
		return r.OnDelete(ctx, email)
	}
	r.ensureMaps()
	delete(r.Saved, email)
	return nil
}

func (r *MockCodeRepository) ensureMaps() {
	if r.Saved == nil {
		r.Saved = map[string]Code{}
	}
	if r.CooldownTill == nil {
		r.CooldownTill = map[string]time.Time{}
	}
}
