package policy

import (
	"context"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

type SafetyPolicy struct {
}

func (p *SafetyPolicy) CanCreate(ctx context.Context, u *auth.User, c *course.Course, r *review.Review) error {
	return nil
}
