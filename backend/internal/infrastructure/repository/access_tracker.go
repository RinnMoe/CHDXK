package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"jcourse/internal/domain/auth"
)

var ackAccessScript = redis.NewScript(`
local removed = 0
for i = 1, #ARGV, 2 do
  local member = ARGV[i]
  local score = ARGV[i + 1]
  local current = redis.call("ZSCORE", KEYS[1], member)
  if current and tonumber(current) == tonumber(score) then
    removed = removed + redis.call("ZREM", KEYS[1], member)
  end
end
return removed
`)

type AccessTrackerRepository struct {
	db         *gorm.DB
	cache      *redis.Client
	flushLimit int
}

type accessRecord struct {
	ID    int64
	At    time.Time
	Score int64
}

func NewAccessTrackerRepository(db *gorm.DB, cache *redis.Client, flushLimit int) *AccessTrackerRepository {
	if flushLimit <= 0 {
		flushLimit = auth.DefaultAccessConfig.FlushBatchSize
	}
	return &AccessTrackerRepository{db: db, cache: cache, flushLimit: flushLimit}
}

func (r *AccessTrackerRepository) RecordUserAccess(ctx context.Context, userID int, at time.Time) error {
	return r.record(ctx, r.userKey(), int64(userID), at)
}

func (r *AccessTrackerRepository) RecordApiKeyAccess(ctx context.Context, apiKeyID int64, at time.Time) error {
	return r.record(ctx, r.apiKeyKey(), apiKeyID, at)
}

func (r *AccessTrackerRepository) Flush(ctx context.Context) (auth.AccessFlushResult, error) {
	users, err := r.readRecords(ctx, r.userKey(), r.flushLimit)
	if err != nil {
		return auth.AccessFlushResult{}, err
	}
	apiKeys, err := r.readRecords(ctx, r.apiKeyKey(), r.flushLimit)
	if err != nil {
		return auth.AccessFlushResult{}, err
	}

	var result auth.AccessFlushResult
	if len(users) > 0 {
		if err := r.flushUsers(ctx, users); err != nil {
			return result, err
		}
		if err := r.ackRecords(ctx, r.userKey(), users); err != nil {
			return result, err
		}
		result.Users = len(users)
	}
	if len(apiKeys) > 0 {
		if err := r.flushApiKeys(ctx, apiKeys); err != nil {
			return result, err
		}
		if err := r.ackRecords(ctx, r.apiKeyKey(), apiKeys); err != nil {
			return result, err
		}
		result.ApiKeys = len(apiKeys)
	}
	return result, nil
}

func (r *AccessTrackerRepository) record(ctx context.Context, key string, id int64, at time.Time) error {
	if r.cache == nil || id <= 0 || at.IsZero() {
		return nil
	}
	member := strconv.FormatInt(id, 10)
	if err := r.cache.ZAddNX(ctx, key, redis.Z{
		Score:  float64(at.UnixMilli()),
		Member: member,
	}).Err(); err != nil {
		logCacheAccessFailure(ctx, "zaddnx", key, err)
		return err
	}
	if err := r.cache.ZAddGT(ctx, key, redis.Z{
		Score:  float64(at.UnixMilli()),
		Member: member,
	}).Err(); err != nil {
		logCacheAccessFailure(ctx, "zaddgt", key, err)
		return err
	}
	return nil
}

func (r *AccessTrackerRepository) readRecords(ctx context.Context, key string, limit int) ([]accessRecord, error) {
	if r.cache == nil {
		return nil, nil
	}
	items, err := r.cache.ZRangeWithScores(ctx, key, 0, int64(limit)-1).Result()
	if err != nil {
		logCacheAccessFailure(ctx, "zrange_with_scores", key, err)
		return nil, err
	}
	records := make([]accessRecord, 0, len(items))
	for _, item := range items {
		id, err := strconv.ParseInt(fmt.Sprint(item.Member), 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		score := int64(item.Score)
		records = append(records, accessRecord{
			ID:    id,
			At:    time.UnixMilli(score),
			Score: score,
		})
	}
	return records, nil
}

func (r *AccessTrackerRepository) flushUsers(ctx context.Context, records []accessRecord) error {
	if len(records) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, record := range records {
			if err := tx.WithContext(ctx).
				Model(&UserEntity{}).
				Where("id = ? AND last_seen_at < ?", record.ID, record.At).
				Update("last_seen_at", record.At).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (r *AccessTrackerRepository) flushApiKeys(ctx context.Context, records []accessRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, record := range records {
			if err := tx.WithContext(ctx).
				Model(&ApiKeyEntity{}).
				Where("id = ? AND (last_used_at IS NULL OR last_used_at < ?)", record.ID, record.At).
				Update("last_used_at", record.At).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *AccessTrackerRepository) ackRecords(ctx context.Context, key string, records []accessRecord) error {
	if r.cache == nil || len(records) == 0 {
		return nil
	}
	args := make([]any, 0, len(records)*2)
	for _, record := range records {
		args = append(args, strconv.FormatInt(record.ID, 10), strconv.FormatInt(record.Score, 10))
	}
	if err := ackAccessScript.Run(ctx, r.cache, []string{key}, args...).Err(); err != nil {
		logCacheAccessFailure(ctx, "eval_ack_access", key, err)
		return err
	}
	return nil
}

func (r *AccessTrackerRepository) userKey() string {
	return redisKey("auth", "access", "user")
}

func (r *AccessTrackerRepository) apiKeyKey() string {
	return redisKey("auth", "access", "api_key")
}

var _ auth.AccessTracker = (*AccessTrackerRepository)(nil)
