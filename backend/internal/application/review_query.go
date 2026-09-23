package application

import (
	"context"
	"sort"
	"strings"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

type ReviewListFilter struct {
	Q        string `form:"q"`
	Semester string `form:"semester"`
	Rating   int    `form:"rating"`
	Order    string `form:"order"`
	OrderBy  string `form:"order_by"`
	Ascend   bool   `form:"ascend"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type ReviewQueryService struct {
	repo             review.ReviewQuery
	courseRepo       course.CourseRepository
	voteRepo         review.VoteRepository
	notificationRepo course.CourseNotificationRepository
}

func NewReviewQueryService(repo review.ReviewQuery, voteRepo review.VoteRepository, notificationRepo course.CourseNotificationRepository, courseRepos ...course.CourseRepository) *ReviewQueryService {
	var courses course.CourseRepository
	if len(courseRepos) > 0 {
		courses = courseRepos[0]
	}
	return &ReviewQueryService{repo: repo, courseRepo: courses, voteRepo: voteRepo, notificationRepo: notificationRepo}
}

func (s *ReviewQueryService) GetCourseReviewFilters(ctx context.Context, courseID int) (*review.ReviewFilters, error) {
	if grouped, ok := s.repo.(review.CourseGroupReviewQuery); ok {
		ids, err := s.courseIDsByCode(ctx, courseID)
		if err != nil {
			return nil, err
		}
		return grouped.GetCourseFiltersByCourseIDs(ctx, ids)
	}
	return s.repo.GetCourseFilters(ctx, courseID)
}

func (s *ReviewQueryService) GetCourseReviewTrend(ctx context.Context, courseID int) ([]review.ReviewTrendItem, error) {
	if grouped, ok := s.repo.(review.CourseGroupReviewQuery); ok {
		ids, err := s.courseIDsByCode(ctx, courseID)
		if err != nil {
			return nil, err
		}
		return grouped.GetCourseTrendByCourseIDs(ctx, ids)
	}
	return s.repo.GetCourseTrend(ctx, courseID)
}

func (s *ReviewQueryService) courseIDsByCode(ctx context.Context, courseID int) ([]int, error) {
	if s.courseRepo == nil {
		return []int{courseID}, nil
	}
	current, err := s.courseRepo.Get(ctx, courseID)
	if err != nil || current == nil {
		return []int{courseID}, err
	}
	courses, _, err := s.courseRepo.FindBy(ctx, course.CourseFilter{Code: current.Code, Page: 1, PageSize: 10000})
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(courses)+1)
	ids = append(ids, courseID)
	for _, item := range courses {
		ids = append(ids, item.ID)
	}
	sort.Ints(ids)
	unique := ids[:0]
	for _, id := range ids {
		if len(unique) == 0 || unique[len(unique)-1] != id {
			unique = append(unique, id)
		}
	}
	return unique, nil
}

func (s *ReviewQueryService) buildDTOs(ctx context.Context, reviews []review.ReviewView, u *auth.User) ([]ReviewDTO, error) {
	myVotes := map[int]review.Vote{}
	if u != nil && len(reviews) > 0 {
		reviewIDs := make([]int, 0, len(reviews))
		for _, r := range reviews {
			reviewIDs = append(reviewIDs, r.ID)
		}
		votes, err := s.voteRepo.FindByReviewsAndUser(ctx, reviewIDs, u.ID)
		if err != nil {
			return nil, err
		}
		myVotes = votes
	}

	items := make([]ReviewDTO, len(reviews))
	for i := range reviews {
		g := review.NewViewGuardian(u, &reviews[i])
		items[i] = newReviewDTO(&reviews[i], g.CanViewPrivate())
		if vote, ok := myVotes[reviews[i].ID]; ok {
			items[i].Vote.MyVote = new(vote.VoteType)
		}
	}
	return items, nil
}

func (s *ReviewQueryService) GetReviewsByCourse(ctx context.Context, courseID int, u *auth.User, f ReviewListFilter) (*PaginatedResult[ReviewDTO], error) {
	orderBy := f.OrderBy
	if orderBy == "" {
		orderBy = f.Order
	}
	q := strings.TrimSpace(f.Q)
	ids, err := s.courseIDsByCode(ctx, courseID)
	if err != nil {
		return nil, err
	}
	reviewFilter := review.ReviewFilter{
		CourseIDs:  ids,
		Q:          q,
		Semester:   f.Semester,
		Rating:     f.Rating,
		OrderBy:    orderBy,
		Ascend:     f.Ascend,
		Page:       f.Page,
		PageSize:   f.PageSize,
		WithCourse: true,
	}

	reviews, total, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}
	items, err := s.buildDTOs(ctx, reviews, u)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[ReviewDTO]{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (s *ReviewQueryService) GetReviewsByUser(ctx context.Context, userID int, u *auth.User, f ReviewListFilter) (*PaginatedResult[ReviewDTO], error) {
	orderBy := f.OrderBy
	if orderBy == "" {
		orderBy = f.Order
	}
	q := strings.TrimSpace(f.Q)
	reviewFilter := review.ReviewFilter{
		UserID:     userID,
		WithCourse: true,
		Q:          q,
		Semester:   f.Semester,
		Rating:     f.Rating,
		OrderBy:    orderBy,
		Ascend:     f.Ascend,
		Page:       f.Page,
		PageSize:   f.PageSize,
	}

	reviews, total, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}
	items, err := s.buildDTOs(ctx, reviews, u)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[ReviewDTO]{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (s *ReviewQueryService) GetReviews(ctx context.Context, user *auth.User, f ReviewListFilter) (*PaginatedResult[ReviewDTO], error) {
	orderBy := f.OrderBy
	if orderBy == "" {
		orderBy = f.Order
	}
	q := strings.TrimSpace(f.Q)

	reviewFilter := review.ReviewFilter{
		Q:          q,
		Semester:   f.Semester,
		Rating:     f.Rating,
		OrderBy:    orderBy,
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
	items, err := s.buildDTOs(ctx, reviews, user)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[ReviewDTO]{
		Items:    items,
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

	orderBy := f.OrderBy
	if orderBy == "" {
		orderBy = f.Order
	}
	q := strings.TrimSpace(f.Q)
	reviewFilter := review.ReviewFilter{
		CourseIDs:  followed,
		Q:          q,
		Semester:   f.Semester,
		Rating:     f.Rating,
		OrderBy:    orderBy,
		Ascend:     f.Ascend,
		Page:       f.Page,
		PageSize:   f.PageSize,
		WithCourse: true,
	}

	reviews, total, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}
	items, err := s.buildDTOs(ctx, reviews, u)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[ReviewDTO]{
		Items:    items,
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
		if err != nil {
			return nil, err
		}
		if vote != nil {
			view.Vote.MyVote = new(vote.VoteType)
		}
	}
	return &view, nil
}

func (s *ReviewQueryService) GetReviewRevisions(ctx context.Context, reviewID int) ([]ReviewRevisionDTO, error) {
	reviewView, err := s.repo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	if reviewView == nil {
		return nil, review.ErrReviewNotFound
	}

	revisions, err := s.repo.FindRevisions(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	items := make([]ReviewRevisionDTO, len(revisions))
	for i := range revisions {
		items[i] = newReviewRevisionDTO(&revisions[i])
	}
	return items, nil
}
