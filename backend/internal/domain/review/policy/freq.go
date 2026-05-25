package policy

import (
	"context"
	"errors"
	"time"

	"github.com/agnivade/levenshtein"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

var (
	ErrSimilarContentDetected = errors.New("too many similar reviews detected")
	ErrSameCourseSpam         = errors.New("too many reviews for the same course")
)

type FrequencyPolicyConfig struct {
	Window          time.Duration
	MaxReviews      int
	SimilarityRatio float64
	SuspendDuration time.Duration
}

var DefaultFrequencyPolicyConfig = FrequencyPolicyConfig{
	Window:          5 * time.Minute,
	MaxReviews:      3,
	SimilarityRatio: 0.85,
	SuspendDuration: 90 * 24 * time.Hour,
}

type FrequencyPolicy struct {
	query  review.ReviewQuery
	config FrequencyPolicyConfig
}

func NewFrequencyPolicy(query review.ReviewQuery, config FrequencyPolicyConfig) *FrequencyPolicy {
	defaults := DefaultFrequencyPolicyConfig
	if config.Window <= 0 {
		config.Window = defaults.Window
	}
	if config.MaxReviews <= 0 {
		config.MaxReviews = defaults.MaxReviews
	}
	if config.SimilarityRatio <= 0 {
		config.SimilarityRatio = defaults.SimilarityRatio
	}
	if config.SuspendDuration <= 0 {
		config.SuspendDuration = defaults.SuspendDuration
	}
	return &FrequencyPolicy{query: query, config: config}
}

func (p *FrequencyPolicy) CanCreate(ctx context.Context, u *auth.User, c *course.Course, r *review.Review) error {
	if p.config.MaxReviews <= 0 {
		return nil
	}

	recent, _, err := p.query.FindBy(ctx, review.ReviewFilter{
		UserID:       u.ID,
		CreatedAfter: time.Now().Add(-p.config.Window),
		OrderBy:      "created_at",
		PageSize:     p.config.MaxReviews,
		WithCourse:   true,
	})
	if err != nil {
		return err
	}

	if len(recent) < p.config.MaxReviews {
		return nil
	}

	similarCount := 0
	sameCourseCodeAll := true
	for _, rev := range recent {
		if !reviewCourseMatches(rev, c) {
			sameCourseCodeAll = false
		}
		if similarity(r.Content, rev.Content) >= p.config.SimilarityRatio {
			similarCount++
		}
	}

	if sameCourseCodeAll {
		return p.newViolation(r, c, ErrSameCourseSpam)
	}

	if similarCount*2 > len(recent) {
		return p.newViolation(r, c, ErrSimilarContentDetected)
	}

	return nil
}

func (p *FrequencyPolicy) newViolation(r *review.Review, c *course.Course, reason error) error {
	return &review.FrequencyViolation{
		Reason:          reason,
		Review:          r,
		Course:          c,
		SuspendDuration: p.config.SuspendDuration,
	}
}

func reviewCourseMatches(r review.ReviewView, c *course.Course) bool {
	if r.Course != nil && c.Code != "" {
		return r.Course.Code == c.Code
	}
	return r.CourseID == c.ID
}

func similarity(a, b string) float64 {
	if a == "" && b == "" {
		return 1.0
	}
	maxLen := max(len(a), len(b))
	if maxLen == 0 {
		return 1.0
	}
	dist := levenshtein.ComputeDistance(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}
