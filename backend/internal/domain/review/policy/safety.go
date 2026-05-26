package policy

import (
	"context"
	"errors"
	"strconv"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

var ErrContentSensitive = errors.New("review content is sensitive")

// ContentModerator is an abstraction over an external content moderation service.
// Implementations may call third-party sensitive-content detection APIs.
type ContentModerator interface {
	IsSensitive(ctx context.Context, accountID string, content string) (bool, error)
}

type SafetyPolicy struct {
	moderator ContentModerator
}

func NewSafetyPolicy(moderator ContentModerator) *SafetyPolicy {
	return &SafetyPolicy{moderator: moderator}
}

func (p *SafetyPolicy) CanCreate(ctx context.Context, u *auth.User, c *course.CourseView, r *review.Review) error {
	if p.moderator == nil {
		return nil
	}
	if r.Content == "" {
		return nil
	}
	accountID := ""
	if u != nil {
		accountID = strconv.Itoa(u.ID)
	}
	sensitive, err := p.moderator.IsSensitive(ctx, accountID, r.Content)
	if err != nil {
		return err
	}
	if sensitive {
		return ErrContentSensitive
	}
	return nil
}
