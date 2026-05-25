//go:build test

package stat

import "context"

type MockDailyStatCommandRepository struct {
	Saved *DailyStat

	OnUpsert func(context.Context, *DailyStat) error
}

func (r *MockDailyStatCommandRepository) Upsert(ctx context.Context, daily *DailyStat) error {
	if r.OnUpsert != nil {
		return r.OnUpsert(ctx, daily)
	}
	copy := *daily
	r.Saved = &copy
	return nil
}
