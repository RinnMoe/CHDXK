package policy_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
)

type fakeReviewQuery struct {
	reviews    []review.ReviewView
	lastFilter review.ReviewFilter
}

func (f *fakeReviewQuery) FindBy(ctx context.Context, filter review.ReviewFilter) ([]review.ReviewView, int64, error) {
	f.lastFilter = filter
	return f.reviews, int64(len(f.reviews)), nil
}

func (f *fakeReviewQuery) GetByID(ctx context.Context, reviewID int) (*review.ReviewView, error) {
	return nil, nil
}

func (f *fakeReviewQuery) GetCourseFilters(ctx context.Context, courseID int) (*review.ReviewFilters, error) {
	return &review.ReviewFilters{}, nil
}

func (f *fakeReviewQuery) FindRevisions(ctx context.Context, reviewID int) ([]review.RevisionView, error) {
	return nil, nil
}

func makeView(courseID int, content string) review.ReviewView {
	return review.ReviewView{
		CourseID: courseID,
		Content:  content,
	}
}

func TestFrequencyPolicy_FewerThanMax(t *testing.T) {
	q := &fakeReviewQuery{
		reviews: []review.ReviewView{
			makeView(1, "a"),
			makeView(1, "b"),
		},
	}
	p := policy.NewFrequencyPolicy(q, policy.FrequencyPolicyConfig{
		Window: time.Hour, MaxReviews: 5, SimilarityRatio: 0.7,
	})

	err := p.CanCreate(context.Background(), &auth.User{ID: 1}, &course.Course{ID: 2}, &review.Review{Content: "new"})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestFrequencyPolicy_SameCourseAll(t *testing.T) {
	q := &fakeReviewQuery{
		reviews: []review.ReviewView{
			makeView(7, "aaa"),
			makeView(7, "bbb"),
			makeView(7, "ccc"),
		},
	}
	p := policy.NewFrequencyPolicy(q, policy.FrequencyPolicyConfig{
		Window: time.Hour, MaxReviews: 3, SimilarityRatio: 0.99,
	})

	err := p.CanCreate(context.Background(), &auth.User{ID: 1}, &course.Course{ID: 7}, &review.Review{Content: "z"})
	if !errors.Is(err, policy.ErrSameCourseSpam) {
		t.Fatalf("expected ErrSameCourseSpam, got %v", err)
	}
}

func TestFrequencyPolicy_SameCourseButDifferentTarget(t *testing.T) {
	q := &fakeReviewQuery{
		reviews: []review.ReviewView{
			makeView(7, "aaa"),
			makeView(7, "bbb"),
			makeView(7, "ccc"),
		},
	}
	p := policy.NewFrequencyPolicy(q, policy.FrequencyPolicyConfig{
		Window: time.Hour, MaxReviews: 3, SimilarityRatio: 0.99,
	})

	// target course is different from the existing reviews' courseID
	err := p.CanCreate(context.Background(), &auth.User{ID: 1}, &course.Course{ID: 999}, &review.Review{Content: "z"})
	if err != nil {
		t.Fatalf("expected nil (different target course), got %v", err)
	}
}

func TestFrequencyPolicy_SimilarContent(t *testing.T) {
	base := strings.Repeat("hello world course is great", 3)
	q := &fakeReviewQuery{
		reviews: []review.ReviewView{
			makeView(1, base),
			makeView(2, base),
			makeView(3, base),
			makeView(4, "totally different content here unrelated"),
		},
	}
	p := policy.NewFrequencyPolicy(q, policy.FrequencyPolicyConfig{
		Window: time.Hour, MaxReviews: 4, SimilarityRatio: 0.7,
	})

	err := p.CanCreate(context.Background(), &auth.User{ID: 1}, &course.Course{ID: 99}, &review.Review{Content: base})
	if !errors.Is(err, policy.ErrSimilarContentDetected) {
		t.Fatalf("expected ErrSimilarContentDetected, got %v", err)
	}
}

func TestFrequencyPolicy_NoSpam(t *testing.T) {
	q := &fakeReviewQuery{
		reviews: []review.ReviewView{
			makeView(1, "alpha"),
			makeView(2, "beta xyz"),
			makeView(3, "gamma 1234"),
			makeView(4, "delta zzz"),
		},
	}
	p := policy.NewFrequencyPolicy(q, policy.FrequencyPolicyConfig{
		Window: time.Hour, MaxReviews: 4, SimilarityRatio: 0.9,
	})

	err := p.CanCreate(context.Background(), &auth.User{ID: 1}, &course.Course{ID: 99}, &review.Review{Content: "completely original review"})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

type fakeModerator struct {
	sensitive bool
	err       error
	gotText   string
}

func (m *fakeModerator) IsSensitive(ctx context.Context, content string) (bool, error) {
	m.gotText = content
	return m.sensitive, m.err
}

func TestSafetyPolicy_Nil(t *testing.T) {
	p := policy.NewSafetyPolicy(nil)
	if err := p.CanCreate(context.Background(), &auth.User{}, &course.Course{}, &review.Review{Content: "hi"}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestSafetyPolicy_Sensitive(t *testing.T) {
	m := &fakeModerator{sensitive: true}
	p := policy.NewSafetyPolicy(m)
	err := p.CanCreate(context.Background(), &auth.User{}, &course.Course{}, &review.Review{Content: "bad text"})
	if !errors.Is(err, policy.ErrContentSensitive) {
		t.Fatalf("expected ErrContentSensitive, got %v", err)
	}
	if m.gotText != "bad text" {
		t.Errorf("moderator got %q", m.gotText)
	}
}

func TestSafetyPolicy_Clean(t *testing.T) {
	p := policy.NewSafetyPolicy(&fakeModerator{sensitive: false})
	if err := p.CanCreate(context.Background(), &auth.User{}, &course.Course{}, &review.Review{Content: "good"}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestSafetyPolicy_ModeratorError(t *testing.T) {
	want := errors.New("network down")
	p := policy.NewSafetyPolicy(&fakeModerator{err: want})
	err := p.CanCreate(context.Background(), &auth.User{}, &course.Course{}, &review.Review{Content: "x"})
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}
