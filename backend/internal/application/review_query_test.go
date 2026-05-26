package application_test

import (
	"context"
	"slices"
	"testing"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

type fakeReviewQuery struct {
	reviews []review.ReviewView
}

var errNotFound = errorString("not found")

type errorString string

func (e errorString) Error() string { return string(e) }

func newFakeReviewQuery() *fakeReviewQuery {
	return &fakeReviewQuery{reviews: make([]review.ReviewView, 0)}
}

func (q *fakeReviewQuery) FindBy(ctx context.Context, filter review.ReviewFilter) ([]review.ReviewView, int64, error) {
	var results []review.ReviewView
	for _, r := range q.reviews {
		if filter.CourseID != 0 && filter.CourseID != r.CourseID {
			continue
		}
		if filter.UserID != 0 && filter.UserID != r.UserID {
			continue
		}
		if len(filter.CourseIDs) > 0 {
			match := slices.Contains(filter.CourseIDs, r.CourseID)
			if !match {
				continue
			}
		}
		if len(filter.ExcludeCourseIDs) > 0 {
			exclude := slices.Contains(filter.ExcludeCourseIDs, r.CourseID)
			if exclude {
				continue
			}
		}
		results = append(results, r)
	}
	return results, int64(len(results)), nil
}

func (q *fakeReviewQuery) GetByID(ctx context.Context, reviewID int) (*review.ReviewView, error) {
	for _, r := range q.reviews {
		if r.ID == reviewID {
			rCopy := r
			return &rCopy, nil
		}
	}
	return nil, errNotFound
}

func (q *fakeReviewQuery) GetCourseFilters(ctx context.Context, courseID int) (*review.ReviewFilters, error) {
	return &review.ReviewFilters{}, nil
}

func (q *fakeReviewQuery) GetCourseTrend(ctx context.Context, courseID int) ([]review.ReviewTrendItem, error) {
	return []review.ReviewTrendItem{}, nil
}

func (q *fakeReviewQuery) FindRevisions(ctx context.Context, reviewID int) ([]review.RevisionView, error) {
	return nil, nil
}

func TestReviewQueryService_GetReviews_WithIgnoredCourses(t *testing.T) {
	reviewRepo := newFakeReviewQuery()
	reviewRepo.reviews = []review.ReviewView{
		{ID: 1, CourseID: 1, Rating: 5},
		{ID: 2, CourseID: 2, Rating: 4},
		{ID: 3, CourseID: 3, Rating: 3},
	}

	notifRepo := newFakeNotificationRepo()
	ctx := context.Background()
	notifRepo.SetLevel(ctx, 100, 2, course.NotificationLevelIgnored)

	voteRepo := &review.MockVoteRepository{}
	svc := application.NewReviewQueryService(reviewRepo, voteRepo, notifRepo)

	t.Run("excludes ignored courses for logged-in user", func(t *testing.T) {
		user := &auth.User{ID: 100}
		result, err := svc.GetReviews(ctx, user, application.ReviewListFilter{})
		if err != nil {
			t.Fatalf("GetReviews: %v", err)
		}
		if result.Total != 2 {
			t.Errorf("total: got %d, want 2 (course 2 is ignored)", result.Total)
		}
	})

	t.Run("returns all reviews for anonymous user", func(t *testing.T) {
		result, err := svc.GetReviews(ctx, nil, application.ReviewListFilter{})
		if err != nil {
			t.Fatalf("GetReviews: %v", err)
		}
		if result.Total != 3 {
			t.Errorf("total: got %d, want 3", result.Total)
		}
	})
}

func TestReviewQueryService_GetReviews_AttachesMyVotes(t *testing.T) {
	reviewRepo := newFakeReviewQuery()
	reviewRepo.reviews = []review.ReviewView{
		{ID: 1, CourseID: 1, Rating: 5},
		{ID: 2, CourseID: 2, Rating: 4},
	}

	voteRepo := &review.MockVoteRepository{Votes: map[int]review.Vote{
		2: {ReviewID: 2, UserID: 100, VoteType: review.VoteDislike},
	}}
	svc := application.NewReviewQueryService(reviewRepo, voteRepo, newFakeNotificationRepo())

	result, err := svc.GetReviews(context.Background(), &auth.User{ID: 100}, application.ReviewListFilter{})
	if err != nil {
		t.Fatalf("GetReviews: %v", err)
	}
	if voteRepo.BatchCalls != 1 {
		t.Fatalf("FindByReviewsAndUser calls = %d, want 1", voteRepo.BatchCalls)
	}
	if result.Items[0].Vote.MyVote != nil {
		t.Fatalf("first review MyVote = %v, want nil", *result.Items[0].Vote.MyVote)
	}
	if result.Items[1].Vote.MyVote == nil || *result.Items[1].Vote.MyVote != review.VoteDislike {
		t.Fatalf("second review MyVote = %v, want %d", result.Items[1].Vote.MyVote, review.VoteDislike)
	}
}

func TestReviewQueryService_GetReviews_AnonymousSkipsMyVotes(t *testing.T) {
	reviewRepo := newFakeReviewQuery()
	reviewRepo.reviews = []review.ReviewView{{ID: 1, CourseID: 1, Rating: 5}}
	voteRepo := &review.MockVoteRepository{}
	svc := application.NewReviewQueryService(reviewRepo, voteRepo, newFakeNotificationRepo())

	if _, err := svc.GetReviews(context.Background(), nil, application.ReviewListFilter{}); err != nil {
		t.Fatalf("GetReviews: %v", err)
	}
	if voteRepo.BatchCalls != 0 {
		t.Fatalf("FindByReviewsAndUser calls = %d, want 0", voteRepo.BatchCalls)
	}
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

	voteRepo := &review.MockVoteRepository{}
	svc := application.NewReviewQueryService(reviewRepo, voteRepo, notifRepo)

	t.Run("returns only followed course reviews", func(t *testing.T) {
		result, err := svc.GetFollowedReviews(ctx, 100, nil, application.ReviewListFilter{})
		if err != nil {
			t.Fatalf("GetFollowedReviews: %v", err)
		}
		if result.Total != 2 {
			t.Errorf("total: got %d, want 2 (courses 1 and 3 are followed)", result.Total)
		}
	})

	t.Run("returns empty when no followed courses", func(t *testing.T) {
		result, err := svc.GetFollowedReviews(ctx, 999, nil, application.ReviewListFilter{})
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
