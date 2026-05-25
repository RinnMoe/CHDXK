package auth

import (
	"context"

	"jcourse/internal/domain/task"
)

func EnsureUserActive(ctx context.Context, u *User) error {
	if u.SuspensionExpired() {
		_ = task.Enqueue(ctx, NewClearExpiredSuspensionTask(u.ID))
		return nil
	}
	if u.IsSuspended() {
		return ErrUserSuspended
	}
	return nil
}
