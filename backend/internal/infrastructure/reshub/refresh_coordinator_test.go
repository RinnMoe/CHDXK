package reshub

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"jcourse/internal/domain/auth"
)

type coordinatorRefreshFunc func(context.Context, string) (auth.ResHubSession, error)

func (f coordinatorRefreshFunc) Refresh(ctx context.Context, token string) (auth.ResHubSession, error) {
	return f(ctx, token)
}

// Use an explicitly selected isolated Redis instance. Never flush a database:
// each test owns only its random credential's keys.
func coordinatorRedis(t *testing.T) (*redis.Client, string) {
	t.Helper()
	addr := os.Getenv("JCOURSE_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("JCOURSE_TEST_REDIS_ADDR is not set")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = client.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("connect to test Redis: %v", err)
	}
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		t.Fatal("generate test credential")
	}
	token := hex.EncodeToString(random[:])
	digest := sha256.Sum256([]byte(token))
	key := "jcourse:oidc-refresh:" + hex.EncodeToString(digest[:])
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// Include additional coordination markers belonging to this exact token.
		keys, err := client.Keys(ctx, key+"*").Result()
		if err == nil && len(keys) > 0 {
			_ = client.Del(ctx, keys...).Err()
		}
	})
	return client, token
}

func coordinatorResult() auth.ResHubSession {
	return auth.ResHubSession{AccessToken: "test-access", RefreshToken: "test-next", ExpiresAt: time.Unix(2000000000, 0).UTC()}
}

func TestRefreshCoordinatorConcurrentProcessesShareRotation(t *testing.T) {
	client, token := coordinatorRedis(t)
	second := redis.NewClient(client.Options())
	t.Cleanup(func() { _ = second.Close() })
	var calls atomic.Int32
	provider := coordinatorRefreshFunc(func(ctx context.Context, _ string) (auth.ResHubSession, error) {
		calls.Add(1)
		select {
		case <-time.After(80 * time.Millisecond):
			return coordinatorResult(), nil
		case <-ctx.Done():
			return auth.ResHubSession{}, ctx.Err()
		}
	})
	coordinators := []*RefreshCoordinator{NewRefreshCoordinator(client, provider), NewRefreshCoordinator(second, provider)}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			got, err := coordinators[i%2].Refresh(context.Background(), token)
			if err != nil || got != coordinatorResult() {
				t.Error("concurrent request did not receive the shared rotation")
			}
		}(i)
	}
	close(start)
	wg.Wait()
	if calls.Load() != 1 {
		t.Errorf("provider calls = %d, want exactly 1", calls.Load())
	}
}

func TestRefreshCoordinatorRequestCancellationStillPublishes(t *testing.T) {
	client, token := coordinatorRedis(t)
	started, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	provider := coordinatorRefreshFunc(func(ctx context.Context, _ string) (auth.ResHubSession, error) {
		calls.Add(1)
		close(started)
		select {
		case <-release:
			return coordinatorResult(), nil
		case <-ctx.Done():
			return auth.ResHubSession{}, ctx.Err()
		}
	})
	coordinator := NewRefreshCoordinator(client, provider)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := coordinator.Refresh(ctx, token); done <- err }()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("rotation did not start")
	}
	cancel()
	close(release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal("cancelled owner failed to publish rotation")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("rotation did not finish")
	}
	got, err := coordinator.Refresh(context.Background(), token)
	if err != nil || got != coordinatorResult() || calls.Load() != 1 {
		t.Fatal("subsequent request did not recover the completed rotation")
	}
}

func TestRefreshCoordinatorProviderFailureNeverReplaysGrant(t *testing.T) {
	for _, providerError := range []error{auth.ErrResHubRefreshUncertain, auth.ErrResHubSession, context.DeadlineExceeded, errors.New("unclassified transport failure")} {
		t.Run(providerError.Error(), func(t *testing.T) {
			client, token := coordinatorRedis(t)
			var calls atomic.Int32
			provider := coordinatorRefreshFunc(func(context.Context, string) (auth.ResHubSession, error) {
				calls.Add(1)
				return auth.ResHubSession{}, providerError
			})
			first := NewRefreshCoordinator(client, provider)
			if _, err := first.Refresh(context.Background(), token); err == nil {
				t.Fatal("provider failure unexpectedly succeeded")
			}
			second := NewRefreshCoordinator(client, provider)
			if _, err := second.Refresh(context.Background(), token); err == nil {
				t.Fatal("spent grant unexpectedly succeeded")
			}
			if calls.Load() != 1 {
				t.Errorf("provider calls = %d: a possibly consumed grant must never be retried", calls.Load())
			}
		})
	}
}

func TestRefreshCoordinatorExplicitSafeFailureCanRetry(t *testing.T) {
	client, token := coordinatorRedis(t)
	var calls atomic.Int32
	provider := coordinatorRefreshFunc(func(context.Context, string) (auth.ResHubSession, error) {
		if calls.Add(1) == 1 {
			return auth.ResHubSession{}, auth.ErrResHubUnavailable
		}
		return coordinatorResult(), nil
	})
	coordinator := NewRefreshCoordinator(client, provider)
	if _, err := coordinator.Refresh(context.Background(), token); !errors.Is(err, auth.ErrResHubUnavailable) {
		t.Fatal("expected safe provider failure")
	}
	got, err := coordinator.Refresh(context.Background(), token)
	if err != nil || got != coordinatorResult() || calls.Load() != 2 {
		t.Fatal("safe provider failure did not allow recovery")
	}
}

func TestRefreshCoordinatorExpiredResultNeverReplaysGrant(t *testing.T) {
	client, token := coordinatorRedis(t)
	var calls atomic.Int32
	provider := coordinatorRefreshFunc(func(context.Context, string) (auth.ResHubSession, error) {
		calls.Add(1)
		return coordinatorResult(), nil
	})
	coordinator := NewRefreshCoordinator(client, provider, time.Hour)
	if _, err := coordinator.Refresh(context.Background(), token); err != nil {
		t.Fatal("initial rotation failed")
	}
	digest := sha256.Sum256([]byte(token))
	key := "jcourse:oidc-refresh:" + hex.EncodeToString(digest[:])
	if err := client.Del(context.Background(), key).Err(); err != nil {
		t.Fatal("remove expired test result")
	}
	if _, err := NewRefreshCoordinator(client, provider, time.Hour).Refresh(context.Background(), token); !errors.Is(err, auth.ErrResHubSession) {
		t.Fatal("old grant without a cached result must require sign-in")
	}
	if calls.Load() != 1 {
		t.Fatal("expired result cache allowed a consumed grant to be replayed")
	}
}

func TestRefreshCoordinatorPublishFailureNeverReplaysGrant(t *testing.T) {
	client, token := coordinatorRedis(t)
	ownerClient := redis.NewClient(client.Options())
	t.Cleanup(func() { _ = ownerClient.Close() })
	var calls atomic.Int32
	provider := coordinatorRefreshFunc(func(context.Context, string) (auth.ResHubSession, error) {
		calls.Add(1)
		// Model Redis becoming unavailable after the provider consumed the grant.
		_ = ownerClient.Close()
		return coordinatorResult(), nil
	})
	if _, err := NewRefreshCoordinator(ownerClient, provider).Refresh(context.Background(), token); err == nil {
		t.Fatal("publish failure unexpectedly succeeded")
	}
	digest := sha256.Sum256([]byte(token))
	key := "jcourse:oidc-refresh:" + hex.EncodeToString(digest[:])
	// The failed owner could not unlock. Remove only this test's known lock to
	// model expiry without waiting for its 30-second lease.
	if err := client.Del(context.Background(), key+":lock").Err(); err != nil {
		t.Fatal("expire failed owner's test lock")
	}
	if _, err := NewRefreshCoordinator(client, provider).Refresh(context.Background(), token); !errors.Is(err, auth.ErrResHubSession) {
		t.Fatal("lost publication must require sign-in")
	}
	if calls.Load() != 1 {
		t.Fatal("lost publication allowed grant replay")
	}
}

func TestRefreshCoordinatorCorruptResultDoesNotReplayGrant(t *testing.T) {
	client, token := coordinatorRedis(t)
	digest := sha256.Sum256([]byte(token))
	key := "jcourse:oidc-refresh:" + hex.EncodeToString(digest[:])
	if err := client.Set(context.Background(), key, "invalid-json", time.Minute).Err(); err != nil {
		t.Fatal("write malformed test result")
	}
	var calls atomic.Int32
	provider := coordinatorRefreshFunc(func(context.Context, string) (auth.ResHubSession, error) {
		calls.Add(1)
		return coordinatorResult(), nil
	})
	if _, err := NewRefreshCoordinator(client, provider).Refresh(context.Background(), token); err == nil || calls.Load() != 0 {
		t.Fatal("corrupt cached result must fail without grant replay")
	}
}

func TestRefreshCoordinatorCancelledWaiterDoesNotRotate(t *testing.T) {
	client, token := coordinatorRedis(t)
	digest := sha256.Sum256([]byte(token))
	lock := "jcourse:oidc-refresh:" + hex.EncodeToString(digest[:]) + ":lock"
	if err := client.Set(context.Background(), lock, "other-test-process", time.Minute).Err(); err != nil {
		t.Fatal("set test lock")
	}
	var calls atomic.Int32
	provider := coordinatorRefreshFunc(func(context.Context, string) (auth.ResHubSession, error) {
		calls.Add(1)
		return coordinatorResult(), nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	_, err := NewRefreshCoordinator(client, provider).Refresh(ctx, token)
	if !errors.Is(err, auth.ErrResHubUnavailable) || calls.Load() != 0 {
		t.Fatal("cancelled waiting request rotated or returned an unexpected error")
	}
	if got, err := client.Get(context.Background(), lock).Result(); err != nil || got != "other-test-process" {
		t.Fatal("waiter removed another process's lock")
	}
}
