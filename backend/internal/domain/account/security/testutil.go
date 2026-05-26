//go:build test

package security

import "context"

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
