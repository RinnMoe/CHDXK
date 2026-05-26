package application

import (
	"context"
	"errors"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	domainemail "jcourse/internal/domain/email"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/task"
	"jcourse/pkg/logx"
)

type ReviewCommandConfig struct {
	HotScores                     course.HotScoreConfig
	Vote                          review.VoteConfig
	FrequencyViolationAdminEmails []string
}

type ReviewCommandService struct {
	reviewService *review.Service
	voteService   *review.VoteService
	reviewRepo    review.ReviewRepository
	config        ReviewCommandConfig
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
		config:        config,
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
		var violation *review.FrequencyViolation
		if errors.As(err, &violation) {
			s.enqueueFrequencyViolationTasks(ctx, violation)
		}
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
		logx.Warn(ctx, "enqueue hot course activity", "user_id", userID, "activity", activity, "course_id", courseID, "err", err)
	}
}

func (s *ReviewCommandService) enqueueFrequencyViolationTasks(ctx context.Context, violation *review.FrequencyViolation) {
	if violation == nil || violation.Review == nil {
		return
	}
	userID := violation.Review.UserID
	if err := task.Enqueue(ctx, auth.NewSuspendUserTask(userID, violation.SuspendDuration)); err != nil {
		logx.Warn(ctx, "enqueue frequency violation suspension", "user_id", userID, "err", err)
	}
	mails, err := violation.NewSpamSuspensionEmails(s.config.FrequencyViolationAdminEmails)
	if err != nil {
		logx.Warn(ctx, "render frequency violation email", "user_id", userID, "err", err)
		return
	}
	for _, mail := range mails {
		if err := task.Enqueue(ctx, domainemail.NewSendEmailTask(review.SpamSuspensionEmailType, mail)); err != nil {
			logx.Warn(ctx, "enqueue frequency violation email", "user_id", userID, "to", mail.To, "err", err)
		}
	}
}
