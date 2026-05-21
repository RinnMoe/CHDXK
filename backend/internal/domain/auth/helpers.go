package auth

import (
	"context"
	"crypto/rand"

	"jcourse/internal/domain/task"
)

func numericCode(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	code := make([]byte, length)
	for i, b := range buf {
		code[i] = byte('0' + int(b)%10)
	}
	return string(code), nil
}

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
