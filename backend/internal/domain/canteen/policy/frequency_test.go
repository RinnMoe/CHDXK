package policy

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
)

type fakeReviewQuery struct {
	count        int
	contents     []string
	countErr     error
	contentErr   error
	lastSince    time.Time
	lastUserID   int
	contentCalls int
}

func (f *fakeReviewQuery) ListReviews(context.Context, canteen.ReviewListFilter) ([]canteen.ReviewView, int64, error) {
	return nil, 0, nil
}

func (f *fakeReviewQuery) GetReview(context.Context, int) (*canteen.ReviewView, error) {
	return nil, nil
}

func (f *fakeReviewQuery) ListUserReviews(context.Context, int, int, int) ([]canteen.ReviewView, int64, error) {
	return nil, 0, nil
}

func (f *fakeReviewQuery) GetUserSummary(context.Context, int, int) (*canteen.UserRatingSummary, error) {
	return nil, nil
}

func (f *fakeReviewQuery) CountSince(_ context.Context, userID int, since time.Time) (int, error) {
	f.lastUserID = userID
	f.lastSince = since
	return f.count, f.countErr
}

func (f *fakeReviewQuery) ListRecentContent(context.Context, int, time.Time, int) ([]string, error) {
	f.contentCalls++
	return f.contents, f.contentErr
}

func TestFrequencyPolicyBurst(t *testing.T) {
	query := &fakeReviewQuery{count: 5}
	policy := NewFrequencyPolicy(query, Config{Window: time.Hour, MaxReviews: 5, SimilarityRatio: 0.85})
	err := policy.Check(context.Background(), &auth.User{ID: 2}, "还不错")
	if err != canteen.ErrTooFrequent {
		t.Fatalf("Check() = %v, want ErrTooFrequent", err)
	}
	if query.lastUserID != 2 || query.lastSince.IsZero() {
		t.Fatalf("query not called with user/window: %+v", query)
	}
}

func TestFrequencyPolicySimilarContent(t *testing.T) {
	query := &fakeReviewQuery{count: 1, contents: []string{"今天的面条很好吃，分量也足"}}
	policy := NewFrequencyPolicy(query, Config{Window: time.Hour, MaxReviews: 5, SimilarityRatio: 0.8})
	err := policy.Check(context.Background(), &auth.User{ID: 2}, "今天的面条很好吃，分量也足")
	if err != canteen.ErrSimilarReview {
		t.Fatalf("Check() = %v, want ErrSimilarReview", err)
	}
}

func TestFrequencyPolicyAllowsDistinctContent(t *testing.T) {
	query := &fakeReviewQuery{count: 1, contents: []string{"今天的面条很好吃"}}
	policy := NewFrequencyPolicy(query, Config{Window: time.Hour, MaxReviews: 5, SimilarityRatio: 0.8})
	if err := policy.Check(context.Background(), &auth.User{ID: 2}, "换了一家窗口，米饭有点硬"); err != nil {
		t.Fatalf("Check() = %v", err)
	}
}

func TestConfigNormalized(t *testing.T) {
	got := Config{}.Normalized()
	if got.Window != DefaultConfig.Window || got.MaxReviews != DefaultConfig.MaxReviews || got.SimilarityRatio != DefaultConfig.SimilarityRatio {
		t.Fatalf("Normalized() = %+v", got)
	}
}
