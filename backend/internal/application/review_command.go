package application

import (
	"context"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

type ReviewCommandService struct {
	reviewService *review.Service
	voteService   *review.VoteService
}

func NewReviewCommandService(
	courseRepo course.CourseRepository,
	reviewRepo review.ReviewRepository,
	voteRepo review.VoteRepository,
	policies []review.CreatePolicy,
) *ReviewCommandService {
	return &ReviewCommandService{
		reviewService: review.NewService(courseRepo, reviewRepo, policies),
		voteService:   review.NewVoteService(reviewRepo, voteRepo),
	}
}

func (s *ReviewCommandService) CreateReview(ctx context.Context, u *auth.User, cmd *CreateReviewCommand) error {
	return s.reviewService.Create(ctx, u, review.CreateReview{
		CourseID: cmd.CourseID,
		Semester: cmd.Semester,
		UserID:   u.ID,
		Rating:   cmd.Rating,
		Content:  cmd.Content,
		Score:    cmd.Score,
		Now:      time.Now(),
	})
}

func (s *ReviewCommandService) UpdateReview(ctx context.Context, u *auth.User, cmd *UpdateReviewCommand) error {
	return s.reviewService.Update(ctx, u, review.UpdateReview{
		ReviewID: cmd.ReviewID,
		Semester: cmd.Semester,
		Rating:   cmd.Rating,
		Content:  cmd.Content,
		Score:    cmd.Score,
		Now:      time.Now(),
	})
}

func (s *ReviewCommandService) DeleteReview(ctx context.Context, u *auth.User, reviewID int) error {
	return s.reviewService.Delete(ctx, u, reviewID)
}

func (s *ReviewCommandService) VoteReview(ctx context.Context, userID int, reviewID int, voteType int) error {
	return s.voteService.Vote(ctx, userID, reviewID, voteType, time.Now())
}
