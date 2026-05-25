package auth

import (
	"context"
	"time"
)

type AccessConfig struct {
	FlushCron        string
	SchedulerEnabled bool
	FlushBatchSize   int
}

var DefaultAccessConfig = AccessConfig{
	FlushCron:        "*/5 * * * *",
	SchedulerEnabled: true,
	FlushBatchSize:   1000,
}

type AccessFlushResult struct {
	Users   int
	ApiKeys int
}

type AccessTracker interface {
	RecordUserAccess(ctx context.Context, userID int, at time.Time) error
	RecordApiKeyAccess(ctx context.Context, apiKeyID int, at time.Time) error
	Flush(ctx context.Context) (AccessFlushResult, error)
}
