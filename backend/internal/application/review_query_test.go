package application_test

import (
	"context"
	"testing"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

type fakeReviewQuery struct {
	reviews []review.ReviewView
}

func newFakeReviewQuery() *fakeReviewQuery {
	return &fakeReviewQuery{reviews: make([]review.ReviewView, 0)}
}

func (q *fakeReviewQuery) FindBy(ctx context.Context, filter review.ReviewFilter) ([]review.ReviewView, int64, error) {
	var results []review.ReviewView
	for _, r := range q.reviews {
		if filter.CourseID != 0 && filter.CourseID != r.CourseID {
			continue
		}
		if len(filter.CourseIDs) > 0 {
			match := false
			for _, cid := range filter.CourseIDs {
				if cid == r.CourseID {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		if len(filter.ExcludeCourseIDs) > 0 {
			exclude := false
			for _, cid := range filter.ExcludeCourseIDs {
				if cid == r.CourseID {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}
		results = append(results, r)
	}
	return results, int64(len(results)), nil
}

func (q *fakeReviewQuery) FindRevisions(ctx context.Context, reviewID int) ([]review.RevisionView, error) {
	return nil, nil
}

func TestReviewQueryService_GetLatestReviews_WithIgnoredCourses(t *testing.T) {
	reviewRepo := newFakeReviewQuery()
	reviewRepo.reviews = []review.ReviewView{
		{ID: 1, CourseID: 1, Rating: 5},
		{ID: 2, CourseID: 2, Rating: 4},
		{ID: 3, CourseID: 3, Rating: 3},
	}

	notifRepo := newFakeNotificationRepo()
	ctx := context.Background()
	notifRepo.SetLevel(ctx, 100, 2, course.NotificationLevelIgnored)

	voteRepo := &fakeVoteRepo{}
	svc := application.NewReviewQueryService(reviewRepo, voteRepo, notifRepo)

	t.Run("excludes ignored courses for logged-in user", func(t *testing.T) {
		user := &auth.User{ID: 100}
		result, err := svc.GetLatestReviews(ctx, user, application.ReviewListFilter{})
		if err != nil {
			t.Fatalf("GetLatestReviews: %v", err)
		}
		if result.Total != 2 {
			t.Errorf("total: got %d, want 2 (course 2 is ignored)", result.Total)
		}
	})

	t.Run("returns all reviews for anonymous user", func(t *testing.T) {
		result, err := svc.GetLatestReviews(ctx, nil, application.ReviewListFilter{})
		if err != nil {
			t.Fatalf("GetLatestReviews: %v", err)
		}
		if result.Total != 3 {
			t.Errorf("total: got %d, want 3", result.Total)
		}
	})
}

func TestReviewQueryService_GetFollowedReviews(t *testing.T) {
	reviewRepo := newFakeReviewQuery()
	reviewRepo.reviews = []review.ReviewView{
		{ID: 1, CourseID: 1, Rating: 5},
		{ID: 2, CourseID: 2, Rating: 4},
		{ID: 3, CourseID: 3, Rating: 3},
	}

	notifRepo := newFakeNotificationRepo()
	ctx := context.Background()
	notifRepo.SetLevel(ctx, 100, 1, course.NotificationLevelFollow)
	notifRepo.SetLevel(ctx, 100, 3, course.NotificationLevelFollow)

	voteRepo := &fakeVoteRepo{}
	svc := application.NewReviewQueryService(reviewRepo, voteRepo, notifRepo)

	t.Run("returns only followed course reviews", func(t *testing.T) {
		result, err := svc.GetFollowedReviews(ctx, 100, application.ReviewListFilter{})
		if err != nil {
			t.Fatalf("GetFollowedReviews: %v", err)
		}
		if result.Total != 2 {
			t.Errorf("total: got %d, want 2 (courses 1 and 3 are followed)", result.Total)
		}
	})

	t.Run("returns empty when no followed courses", func(t *testing.T) {
		result, err := svc.GetFollowedReviews(ctx, 999, application.ReviewListFilter{})
		if err != nil {
			t.Fatalf("GetFollowedReviews: %v", err)
		}
		if result.Total != 0 {
			t.Errorf("total: got %d, want 0", result.Total)
		}
		if len(result.Items) != 0 {
			t.Errorf("Items length: got %d, want 0", len(result.Items))
		}
	})
}

type fakeVoteRepo struct{}

func (v *fakeVoteRepo) FindByReviewAndUser(ctx context.Context, reviewID, userID int) (*review.Vote, error) {
	return nil, nil
}

func (v *fakeVoteRepo) CountTodayByUser(ctx context.Context, userID int) (int64, error) {
	return 0, nil
}

func (v *fakeVoteRepo) Save(ctx context.Context, vote *review.Vote) error {
	return nil
}

func (v *fakeVoteRepo) Delete(ctx context.Context, reviewID, userID int) error {
	return nil
}
