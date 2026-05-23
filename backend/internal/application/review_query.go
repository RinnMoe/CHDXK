package application

import (
	"context"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

type ReviewListFilter struct {
	Q        string `form:"q"`
	Semester string `form:"semester"`
	Rating   int    `form:"rating"`
	OrderBy  string `form:"order_by"`
	Ascend   bool   `form:"ascend"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type ReviewQueryService struct {
	repo             review.ReviewQuery
	voteRepo         review.VoteRepository
	notificationRepo course.CourseNotificationRepository
}

func NewReviewQueryService(repo review.ReviewQuery, voteRepo review.VoteRepository, notificationRepo course.CourseNotificationRepository) *ReviewQueryService {
	return &ReviewQueryService{repo: repo, voteRepo: voteRepo, notificationRepo: notificationRepo}
}

func (s *ReviewQueryService) GetCourseReviewFilters(ctx context.Context, courseID int) (*review.ReviewFilters, error) {
	return s.repo.GetCourseFilters(ctx, courseID)
}

func (s *ReviewQueryService) GetCourseReviewTrend(ctx context.Context, courseID int) ([]review.ReviewTrendItem, error) {
	return s.repo.GetCourseTrend(ctx, courseID)
}

func (s *ReviewQueryService) buildDTOs(reviews []review.ReviewView, u *auth.User) []ReviewDTO {
	items := make([]ReviewDTO, len(reviews))
	for i, r := range reviews {
		g := review.NewViewGuardian(u, &reviews[i])
		items[i] = newReviewDTO(&r, g.CanViewPrivate())
	}
	return items
}

func (s *ReviewQueryService) GetReviewsByCourse(ctx context.Context, courseID int, u *auth.User, f ReviewListFilter) (*PaginatedResult[ReviewDTO], error) {
	reviewFilter := review.ReviewFilter{
		CourseID: courseID,
		Q:        f.Q,
		Semester: f.Semester,
		Rating:   f.Rating,
		OrderBy:  f.OrderBy,
		Ascend:   f.Ascend,
		Page:     f.Page,
		PageSize: f.PageSize,
	}

	reviews, total, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[ReviewDTO]{
		Items:    s.buildDTOs(reviews, u),
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (s *ReviewQueryService) GetReviewsByUser(ctx context.Context, userID int, u *auth.User, f ReviewListFilter) (*PaginatedResult[ReviewDTO], error) {
	reviewFilter := review.ReviewFilter{
		UserID:     userID,
		WithCourse: true,
		Q:          f.Q,
		Semester:   f.Semester,
		Rating:     f.Rating,
		OrderBy:    f.OrderBy,
		Ascend:     f.Ascend,
		Page:       f.Page,
		PageSize:   f.PageSize,
	}

	reviews, total, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[ReviewDTO]{
		Items:    s.buildDTOs(reviews, u),
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (s *ReviewQueryService) GetLatestReviews(ctx context.Context, user *auth.User, f ReviewListFilter) (*PaginatedResult[ReviewDTO], error) {
	reviewFilter := review.ReviewFilter{
		Q:          f.Q,
		Semester:   f.Semester,
		Rating:     f.Rating,
		OrderBy:    f.OrderBy,
		Ascend:     f.Ascend,
		Page:       f.Page,
		PageSize:   f.PageSize,
		WithCourse: true,
	}

	if user != nil {
		ignored, err := s.notificationRepo.GetCoursesByLevel(ctx, user.ID, course.NotificationLevelIgnored)
		if err != nil {
			return nil, err
		}
		if len(ignored) > 0 {
			reviewFilter.ExcludeCourseIDs = ignored
		}
	}

	reviews, total, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[ReviewDTO]{
		Items:    s.buildDTOs(reviews, user),
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (s *ReviewQueryService) GetFollowedReviews(ctx context.Context, userID int, u *auth.User, f ReviewListFilter) (*PaginatedResult[ReviewDTO], error) {
	followed, err := s.notificationRepo.GetCoursesByLevel(ctx, userID, course.NotificationLevelFollow)
	if err != nil {
		return nil, err
	}
	if len(followed) == 0 {
		return &PaginatedResult[ReviewDTO]{
			Items:    []ReviewDTO{},
			Total:    0,
			Page:     f.Page,
			PageSize: f.PageSize,
		}, nil
	}

	reviewFilter := review.ReviewFilter{
		CourseIDs:  followed,
		Q:          f.Q,
		Semester:   f.Semester,
		Rating:     f.Rating,
		OrderBy:    f.OrderBy,
		Ascend:     f.Ascend,
		Page:       f.Page,
		PageSize:   f.PageSize,
		WithCourse: true,
	}

	reviews, total, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[ReviewDTO]{
		Items:    s.buildDTOs(reviews, u),
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (s *ReviewQueryService) GetReviewByID(ctx context.Context, u *auth.User, reviewID int) (*ReviewDTO, error) {
	reviewView, err := s.repo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	if reviewView == nil {
		return nil, review.ErrReviewNotFound
	}
	g := review.NewViewGuardian(u, reviewView)
	view := newReviewDTO(reviewView, g.CanViewPrivate())
	if u != nil {
		vote, err := s.voteRepo.FindByReviewAndUser(ctx, reviewID, u.ID)
		if err == nil && vote != nil {
			vt := vote.VoteType
			view.Vote.MyVote = &vt
		}
	}
	return &view, nil
}
