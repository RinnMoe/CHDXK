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

func (f *fakeReviewQuery) GetCourseTrend(ctx context.Context, courseID int) ([]review.ReviewTrendItem, error) {
	return []review.ReviewTrendItem{}, nil
}

func (f *fakeReviewQuery) FindRevisions(ctx context.Context, reviewID int) ([]review.RevisionView, error) {
	return nil, nil
}

func makeView(courseID int, content string) review.ReviewView {
	return makeViewWithCode(courseID, "", content)
}

func makeViewWithCode(courseID int, code string, content string) review.ReviewView {
	return review.ReviewView{
		CourseID: courseID,
		Content:  content,
		Course:   &course.CourseView{ID: courseID, Code: code},
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

	err := p.CanCreate(context.Background(), &auth.User{ID: 1}, &course.CourseView{ID: 2}, &review.Review{Content: "new"})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestFrequencyPolicy_SameCourseAll(t *testing.T) {
	q := &fakeReviewQuery{
		reviews: []review.ReviewView{
			makeViewWithCode(7, "CS101", "aaa"),
			makeViewWithCode(8, "CS101", "bbb"),
			makeViewWithCode(9, "CS101", "ccc"),
		},
	}
	p := policy.NewFrequencyPolicy(q, policy.FrequencyPolicyConfig{
		Window:          time.Hour,
		MaxReviews:      3,
		SimilarityRatio: 0.99,
		SuspendDuration: 2 * time.Hour,
	})

	targetCourse := &course.CourseView{ID: 10, Code: "CS101", Name: "Intro CS"}
	targetReview := &review.Review{UserID: 1, CourseID: 10, Content: "z"}
	before := time.Now()
	err := p.CanCreate(context.Background(), &auth.User{ID: 1}, targetCourse, targetReview)
	after := time.Now()
	if !errors.Is(err, policy.ErrSameCourseSpam) {
		t.Fatalf("expected ErrSameCourseSpam, got %v", err)
	}
	var violation *review.FrequencyViolation
	if !errors.As(err, &violation) {
		t.Fatalf("expected FrequencyViolation, got %T", err)
	}
	if violation.Review != targetReview || violation.Course != targetCourse || violation.SuspendDuration != 2*time.Hour {
		t.Fatalf("violation = %+v", violation)
	}
	if violation.BannedUntil.Before(before.Add(2*time.Hour)) || violation.BannedUntil.After(after.Add(2*time.Hour)) {
		t.Fatalf("BannedUntil = %s, want now + 2h", violation.BannedUntil)
	}
}

func TestFrequencyPolicy_SameCourseIDButDifferentCode(t *testing.T) {
	q := &fakeReviewQuery{
		reviews: []review.ReviewView{
			makeViewWithCode(7, "CS101", "aaa"),
			makeViewWithCode(7, "CS101", "bbb"),
			makeViewWithCode(7, "CS101", "ccc"),
		},
	}
	p := policy.NewFrequencyPolicy(q, policy.FrequencyPolicyConfig{
		Window: time.Hour, MaxReviews: 3, SimilarityRatio: 0.99,
	})

	err := p.CanCreate(context.Background(), &auth.User{ID: 1}, &course.CourseView{ID: 7, Code: "CS999"}, &review.Review{Content: "z"})
	if err != nil {
		t.Fatalf("expected nil (different target code), got %v", err)
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
		Window: time.Hour, MaxReviews: 4, SimilarityRatio: 0.7, SuspendDuration: 3 * time.Hour,
	})

	targetCourse := &course.CourseView{ID: 99, Code: "CS999"}
	targetReview := &review.Review{UserID: 1, CourseID: 99, Content: base}
	before := time.Now()
	err := p.CanCreate(context.Background(), &auth.User{ID: 1}, targetCourse, targetReview)
	after := time.Now()
	if !errors.Is(err, policy.ErrSimilarContentDetected) {
		t.Fatalf("expected ErrSimilarContentDetected, got %v", err)
	}
	var violation *review.FrequencyViolation
	if !errors.As(err, &violation) {
		t.Fatalf("expected FrequencyViolation, got %T", err)
	}
	if violation.Review != targetReview || violation.Course != targetCourse || violation.SuspendDuration != 3*time.Hour {
		t.Fatalf("violation = %+v", violation)
	}
	if violation.BannedUntil.Before(before.Add(3*time.Hour)) || violation.BannedUntil.After(after.Add(3*time.Hour)) {
		t.Fatalf("BannedUntil = %s, want now + 3h", violation.BannedUntil)
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

	err := p.CanCreate(context.Background(), &auth.User{ID: 1}, &course.CourseView{ID: 99}, &review.Review{Content: "completely original review"})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

type fakeModerator struct {
	sensitive bool
	err       error
	gotText   string
}

func (m *fakeModerator) IsSensitive(ctx context.Context, accountID string, content string) (bool, error) {
	m.gotText = content
	return m.sensitive, m.err
}

func TestSafetyPolicy_Nil(t *testing.T) {
	p := policy.NewSafetyPolicy(nil)
	if err := p.CanCreate(context.Background(), &auth.User{}, &course.CourseView{}, &review.Review{Content: "hi"}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestSafetyPolicy_Sensitive(t *testing.T) {
	m := &fakeModerator{sensitive: true}
	p := policy.NewSafetyPolicy(m)
	err := p.CanCreate(context.Background(), &auth.User{}, &course.CourseView{}, &review.Review{Content: "bad text"})
	if !errors.Is(err, policy.ErrContentSensitive) {
		t.Fatalf("expected ErrContentSensitive, got %v", err)
	}
	if m.gotText != "bad text" {
		t.Errorf("moderator got %q", m.gotText)
	}
}

func TestSafetyPolicy_Clean(t *testing.T) {
	p := policy.NewSafetyPolicy(&fakeModerator{sensitive: false})
	if err := p.CanCreate(context.Background(), &auth.User{}, &course.CourseView{}, &review.Review{Content: "good"}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestSafetyPolicy_ModeratorError(t *testing.T) {
	want := errors.New("network down")
	p := policy.NewSafetyPolicy(&fakeModerator{err: want})
	err := p.CanCreate(context.Background(), &auth.User{}, &course.CourseView{}, &review.Review{Content: "x"})
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}
