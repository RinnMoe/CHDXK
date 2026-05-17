package policy

import (
	"context"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

type FrequencyPolicy struct {
	repo review.ReviewRepository
}

func NewFrequencyPolicy(repo review.ReviewRepository) *FrequencyPolicy {
	return &FrequencyPolicy{repo: repo}
}

func (p *FrequencyPolicy) CanCreate(ctx context.Context, u *auth.User, c *course.Course, r *review.Review) error {
	return nil
}
