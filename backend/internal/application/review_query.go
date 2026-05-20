package application

import (
	"context"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/review"
)

type ReviewListFilter struct {
	Semester string `form:"semester"`
	Rating   int    `form:"rating"`
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type ReviewQueryService struct {
	repo     review.ReviewQuery
	voteRepo review.VoteRepository
}

func NewReviewQueryService(repo review.ReviewQuery, voteRepo review.VoteRepository) *ReviewQueryService {
	return &ReviewQueryService{repo: repo, voteRepo: voteRepo}
}

func (s *ReviewQueryService) GetReviewsByCourse(ctx context.Context, courseID int, f ReviewListFilter) (*PaginatedResult[ReviewDTO], error) {
	reviewFilter := review.ReviewFilter{
		CourseID: courseID,
		Semester: f.Semester,
		Rating:   f.Rating,
		OrderBy:  f.OrderBy,
		OrderDir: f.OrderDir,
		Page:     f.Page,
		PageSize: f.PageSize,
	}

	reviews, total, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}

	items := make([]ReviewDTO, len(reviews))
	for i, r := range reviews {
		items[i] = newReviewDTO(&r)
	}

	return &PaginatedResult[ReviewDTO]{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (s *ReviewQueryService) GetReviewsByUser(ctx context.Context, userID int, f ReviewListFilter) (*PaginatedResult[ReviewDTO], error) {
	reviewFilter := review.ReviewFilter{
		UserID:     userID,
		WithCourse: true,
		Semester:   f.Semester,
		Rating:     f.Rating,
		OrderBy:    f.OrderBy,
		OrderDir:   f.OrderDir,
		Page:       f.Page,
		PageSize:   f.PageSize,
	}

	reviews, total, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}

	items := make([]ReviewDTO, len(reviews))
	for i, r := range reviews {
		items[i] = newReviewDTO(&r)
	}

	return &PaginatedResult[ReviewDTO]{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (s *ReviewQueryService) GetLatestReviews(ctx context.Context, f ReviewListFilter) (*PaginatedResult[ReviewDTO], error) {
	reviewFilter := review.ReviewFilter{
		Semester:   f.Semester,
		Rating:     f.Rating,
		OrderBy:    f.OrderBy,
		OrderDir:   f.OrderDir,
		Page:       f.Page,
		PageSize:   f.PageSize,
		WithCourse: true,
	}

	reviews, total, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}

	items := make([]ReviewDTO, len(reviews))
	for i, r := range reviews {
		items[i] = newReviewDTO(&r)
	}

	return &PaginatedResult[ReviewDTO]{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (s *ReviewQueryService) GetReview(ctx context.Context, u *auth.User, reviewID int) (*ReviewDTO, error) {
	reviews, _, err := s.repo.FindBy(ctx, review.ReviewFilter{ReviewID: reviewID})
	if err != nil {
		return nil, err
	}
	view := newReviewDTO(&reviews[0])
	if u != nil {
		vote, err := s.voteRepo.FindByReviewAndUser(ctx, reviewID, u.ID)
		if err == nil && vote != nil {
			vt := vote.VoteType
			view.Vote.MyVote = &vt
		}
	}
	return &view, nil
}
