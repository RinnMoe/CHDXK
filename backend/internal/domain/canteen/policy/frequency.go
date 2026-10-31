package policy

import (
	"context"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
)

// FrequencyPolicy rejects bursts of similar canteen reviews. It mirrors the
// course-domain rule without coupling the two packages.
type FrequencyPolicy struct {
	query  canteen.ReviewQuery
	config Config
}

func NewFrequencyPolicy(query canteen.ReviewQuery, config Config) *FrequencyPolicy {
	return &FrequencyPolicy{query: query, config: config.Normalized()}
}

func (p *FrequencyPolicy) Check(ctx context.Context, user *auth.User, content string) error {
	if p == nil || p.query == nil || user == nil || p.config.MaxReviews <= 0 {
		return nil
	}
	since := time.Now().Add(-p.config.Window)
	count, err := p.query.CountSince(ctx, user.ID, since)
	if err != nil {
		return err
	}
	if count >= p.config.MaxReviews {
		return canteen.ErrTooFrequent
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	recent, err := p.query.ListRecentContent(ctx, user.ID, since, p.config.MaxReviews)
	if err != nil {
		return err
	}
	for _, previous := range recent {
		if contentSimilarity(previous, content) >= p.config.SimilarityRatio {
			return canteen.ErrSimilarReview
		}
	}
	return nil
}

func contentSimilarity(a, b string) float64 {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" || b == "" {
		return 0
	}
	longest := len([]rune(a))
	if other := len([]rune(b)); other > longest {
		longest = other
	}
	if longest == 0 {
		return 1
	}
	distance := levenshtein.ComputeDistance(a, b)
	return 1 - float64(distance)/float64(longest)
}
