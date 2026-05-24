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
}

var DefaultFrequencyPolicyConfig = FrequencyPolicyConfig{
	Window:          time.Hour,
	MaxReviews:      10,
	SimilarityRatio: 0.7,
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
	})
	if err != nil {
		return err
	}

	if len(recent) < p.config.MaxReviews {
		return nil
	}

	similarCount := 0
	sameCourseAll := true
	firstCourseID := recent[0].CourseID
	for _, rev := range recent {
		if rev.CourseID != firstCourseID {
			sameCourseAll = false
		}
		if similarity(r.Content, rev.Content) >= p.config.SimilarityRatio {
			similarCount++
		}
	}

	if sameCourseAll && firstCourseID == c.ID {
		return ErrSameCourseSpam
	}

	if similarCount*2 > len(recent) {
		return ErrSimilarContentDetected
	}

	return nil
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
