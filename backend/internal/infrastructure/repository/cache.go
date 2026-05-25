package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
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
		return nil, false
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
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
		return
	}
	_ = client.Set(ctx, key, data, ttl).Err()
}

func cacheDelete(ctx context.Context, client *redis.Client, keys ...string) {
	if client == nil || len(keys) == 0 {
		return
	}
	_ = client.Del(ctx, keys...).Err()
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
	if len(keys) > 0 {
		cacheDelete(ctx, client, keys...)
	}
}
