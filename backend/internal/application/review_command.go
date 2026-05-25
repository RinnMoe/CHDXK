package application

import (
	"context"
	"log"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/task"
)

type ReviewCommandConfig struct {
	HotScores course.HotScoreConfig
	Vote      review.VoteConfig
}

type ReviewCommandService struct {
	reviewService *review.Service
	voteService   *review.VoteService
	reviewRepo    review.ReviewRepository
}

func NewReviewCommandService(
	courseRepo course.CourseRepository,
	reviewRepo review.ReviewRepository,
	voteRepo review.VoteRepository,
	config ReviewCommandConfig,
	policies []review.CreatePolicy,
) *ReviewCommandService {
	return &ReviewCommandService{
		reviewService: review.NewService(courseRepo, reviewRepo, policies),
		voteService:   review.NewVoteService(reviewRepo, voteRepo, config.Vote),
		reviewRepo:    reviewRepo,
	}
}

func (s *ReviewCommandService) CreateReview(ctx context.Context, u *auth.User, cmd *CreateReviewCommand) error {
	now := time.Now()
	err := s.reviewService.Create(ctx, u, review.CreateReview{
		CourseID: cmd.CourseID,
		Semester: cmd.Semester,
		UserID:   u.ID,
		Rating:   cmd.Rating,
		Content:  cmd.Content,
		Score:    cmd.Score,
		Now:      now,
	})
	if err != nil {
		return err
	}
	s.enqueueHotCourseActivity(ctx, u.ID, course.HotCourseActivityReviewCreate, cmd.CourseID)
	return nil
}

func (s *ReviewCommandService) UpdateReview(ctx context.Context, u *auth.User, cmd *UpdateReviewCommand) error {
	existing, err := s.reviewRepo.Get(ctx, cmd.ReviewID)
	if err != nil {
		return err
	}
	if existing == nil {
		return review.ErrReviewNotFound
	}

	now := time.Now()
	err = s.reviewService.Update(ctx, u, review.UpdateReview{
		ReviewID: cmd.ReviewID,
		Semester: cmd.Semester,
		Rating:   cmd.Rating,
		Content:  cmd.Content,
		Score:    cmd.Score,
		Now:      now,
	})
	if err != nil {
		return err
	}
	s.enqueueHotCourseActivity(ctx, u.ID, course.HotCourseActivityReviewUpdate, existing.CourseID)
	return nil
}

func (s *ReviewCommandService) DeleteReview(ctx context.Context, u *auth.User, reviewID int) error {
	return s.reviewService.Delete(ctx, u, reviewID)
}

func (s *ReviewCommandService) UpdateModeratorRemark(ctx context.Context, u *auth.User, reviewID int, cmd *UpdateReviewModeratorRemarkCommand) error {
	return s.reviewService.UpdateModeratorRemark(ctx, u, review.UpdateModeratorRemark{
		ReviewID:        reviewID,
		ModeratorRemark: cmd.ModeratorRemark,
	})
}

func (s *ReviewCommandService) VoteReview(ctx context.Context, userID int, reviewID int, voteType int) error {
	now := time.Now()
	result, err := s.voteService.Vote(ctx, userID, reviewID, voteType, now)
	if err != nil {
		return err
	}
	if result != nil && result.Changed {
		s.enqueueHotCourseActivity(ctx, userID, course.HotCourseActivityReviewVote, result.CourseID)
	}
	return nil
}

func (s *ReviewCommandService) enqueueHotCourseActivity(ctx context.Context, userID int, activity course.HotCourseActivity, courseID int) {
	if err := task.Enqueue(ctx, course.NewRecordHotCourseActivityTask(userID, activity, courseID)); err != nil {
		log.Printf("enqueue hot course activity: %v", err)
	}
}
