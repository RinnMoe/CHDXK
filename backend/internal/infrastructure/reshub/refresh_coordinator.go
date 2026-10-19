package reshub

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"jcourse/internal/domain/auth"
)

// RefreshCoordinator shares a rotation result across API processes. Concurrent
// requests carrying the same old session must never reuse a single-use grant.
type RefreshCoordinator struct {
	redis     *redis.Client
	provider  auth.ResHubSessionRefresher
	retention time.Duration
}

func NewRefreshCoordinator(client *redis.Client, provider auth.ResHubSessionRefresher, lifetime ...time.Duration) *RefreshCoordinator {
	retention := 30 * 24 * time.Hour
	if len(lifetime) > 0 && lifetime[0] > 0 {
		retention = lifetime[0]
	}
	return &RefreshCoordinator{redis: client, provider: provider, retention: retention}
}

func (r *RefreshCoordinator) Refresh(ctx context.Context, token string) (auth.ResHubSession, error) {
	if r == nil || r.redis == nil || r.provider == nil || token == "" {
		return auth.ResHubSession{}, auth.ErrResHubUnavailable
	}
	digest := sha256.Sum256([]byte(token))
	key := "jcourse:oidc-refresh:" + hex.EncodeToString(digest[:])
	var owner [16]byte
	if _, err := rand.Read(owner[:]); err != nil {
		return auth.ResHubSession{}, auth.ErrResHubUnavailable
	}
	lockOwner := hex.EncodeToString(owner[:])
	waitCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	for {
		cached, err := r.redis.Get(waitCtx, key).Bytes()
		if err == nil {
			var result auth.ResHubSession
			if json.Unmarshal(cached, &result) != nil {
				return result, auth.ErrResHubUnavailable
			}
			return result, nil
		}
		if !errors.Is(err, redis.Nil) {
			return auth.ResHubSession{}, auth.ErrResHubUnavailable
		}
		locked, err := r.redis.SetNX(waitCtx, key+":lock", lockOwner, 30*time.Second).Result()
		if err != nil {
			return auth.ResHubSession{}, auth.ErrResHubUnavailable
		}
		if locked {
			defer func() {
				unlockCtx, done := context.WithTimeout(context.Background(), time.Second)
				defer done()
				_ = r.redis.Eval(unlockCtx, `if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) end return 0`, []string{key + ":lock"}, lockOwner).Err()
			}()
			// The prior owner can publish between our GET and SET NX.
			cached, err = r.redis.Get(waitCtx, key).Bytes()
			if err == nil {
				var result auth.ResHubSession
				if json.Unmarshal(cached, &result) != nil {
					return result, auth.ErrResHubUnavailable
				}
				return result, nil
			}
			if !errors.Is(err, redis.Nil) {
				return auth.ResHubSession{}, auth.ErrResHubUnavailable
			}
			// Mark the grant before sending it. A process crash, lost response, or
			// expired result cache must never cause a replay that revokes its family.
			unused, err := r.redis.SetNX(waitCtx, key+":consumed", "1", r.retention).Result()
			if err != nil {
				return auth.ResHubSession{}, auth.ErrResHubUnavailable
			}
			if !unused {
				return auth.ResHubSession{}, auth.ErrResHubSession
			}
			// Finish and publish the rotation even when its browser request is cancelled.
			rotationCtx, done := context.WithTimeout(context.Background(), 15*time.Second)
			defer done()
			result, err := r.provider.Refresh(rotationCtx, token)
			if err != nil {
				// An explicit provider rejection that did not consume the grant can
				// be retried later. Transport ambiguity is deliberately fail-closed.
				if errors.Is(err, auth.ErrResHubUnavailable) {
					// Discovery may have exhausted rotationCtx without consuming
					// the grant. Cleanup needs its own bounded lifetime.
					cleanupCtx, cleanupDone := context.WithTimeout(context.Background(), time.Second)
					cleanupErr := r.redis.Del(cleanupCtx, key+":consumed").Err()
					cleanupDone()
					if cleanupErr != nil {
						// Retain the marker when recovery cannot be confirmed;
						// never silently allow replay after a storage failure.
						return auth.ResHubSession{}, fmt.Errorf("%w: could not clear unused refresh grant", auth.ErrResHubUnavailable)
					}
				}
				return auth.ResHubSession{}, err
			}
			data, err := json.Marshal(result)
			if err != nil {
				return auth.ResHubSession{}, auth.ErrResHubUnavailable
			}
			if err := r.redis.Set(rotationCtx, key, data, 2*time.Minute).Err(); err != nil {
				return auth.ResHubSession{}, auth.ErrResHubUnavailable
			}
			return result, nil
		}
		timer := time.NewTimer(25 * time.Millisecond)
		select {
		case <-waitCtx.Done():
			timer.Stop()
			return auth.ResHubSession{}, auth.ErrResHubUnavailable
		case <-timer.C:
		}
	}
}
