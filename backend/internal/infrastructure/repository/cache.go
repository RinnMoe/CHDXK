package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"jcourse/pkg/logx"
)

const repositoryCacheTTL = 30 * time.Minute

func redisKey(domain string, parts ...any) string {
	var key strings.Builder
	key.WriteString("jcourse:" + domain)
	for _, part := range parts {
		key.WriteString(":" + fmt.Sprint(part))
	}
	return key.String()
}

func cacheKey(domain string, parts ...any) string {
	return redisKey(domain, parts...)
}

func cacheGetJSON[T any](ctx context.Context, client *redis.Client, key string) (*T, bool) {
	if client == nil {
		return nil, false
	}

	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logCacheAccessFailure(ctx, "get", key, err)
		}
		return nil, false
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		logx.Warn(ctx, "cache payload invalid", "operation", "unmarshal", "key", key, "err", err)
		cacheDelete(ctx, client, key)
		return nil, false
	}
	return &value, true
}

func cacheSetJSON(ctx context.Context, client *redis.Client, key string, value any) {
	cacheSetJSONWithTTL(ctx, client, key, value, repositoryCacheTTL)
}

func cacheSetJSONWithTTL(ctx context.Context, client *redis.Client, key string, value any, ttl time.Duration) {
	if client == nil || value == nil {
		return
	}

	data, err := json.Marshal(value)
	if err != nil {
		logx.Warn(ctx, "cache payload marshal failed", "operation", "marshal", "key", key, "err", err)
		return
	}
	if err := client.Set(ctx, key, data, ttl).Err(); err != nil {
		logCacheAccessFailure(ctx, "set", key, err)
	}
}

func cacheDelete(ctx context.Context, client *redis.Client, keys ...string) {
	if client == nil || len(keys) == 0 {
		return
	}
	if err := client.Del(ctx, keys...).Err(); err != nil {
		logx.Warn(ctx, "cache access failed", "operation", "del", "keys", keys, "err", err)
	}
}

func cacheDeletePattern(ctx context.Context, client *redis.Client, pattern string) {
	if client == nil {
		return
	}

	iter := client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 100 {
			cacheDelete(ctx, client, keys...)
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		logx.Warn(ctx, "cache access failed", "operation", "scan", "pattern", pattern, "err", err)
	}
	if len(keys) > 0 {
		cacheDelete(ctx, client, keys...)
	}
}

func logCacheAccessFailure(ctx context.Context, operation, key string, err error) {
	logx.Warn(ctx, "cache access failed", "operation", operation, "key", key, "err", err)
}
