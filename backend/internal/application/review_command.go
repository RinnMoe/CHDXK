package application

import (
	"context"
	"errors"
	"strconv"
	"time"

	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	domainemail "jcourse/internal/domain/email"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/domain/task"
	"jcourse/pkg/logx"
)

type ReviewCommandConfig struct {
	HotScores                     course.HotScoreConfig
	FrequencyViolationAdminEmails []string
}

type ReviewCommandService struct {
	courseRepo     course.CourseRepository
	reviewRepo     review.ReviewRepository
	reviewQuery    review.ReviewQuery
	voteRepo       review.VoteRepository
	settings       SiteSettingsProvider
	config         ReviewCommandConfig
	createPolicies []review.CreatePolicy
}

func NewReviewCommandService(
	courseRepo course.CourseRepository,
	reviewRepo review.ReviewRepository,
	voteRepo review.VoteRepository,
	settings SiteSettingsProvider,
	config ReviewCommandConfig,
	policies []review.CreatePolicy,
) *ReviewCommandService {
	if settings == nil {
		defaults := NewDefaultSiteSettingsProvider()
		settings = defaults
	}
	reviewQuery, _ := reviewRepo.(review.ReviewQuery)
	return &ReviewCommandService{
		courseRepo:     courseRepo,
		reviewRepo:     reviewRepo,
		reviewQuery:    reviewQuery,
		voteRepo:       voteRepo,
		settings:       settings,
		config:         config,
		createPolicies: policies,
	}
}

func (s *ReviewCommandService) CreateReview(ctx context.Context, u *auth.User, cmd *CreateReviewCommand) error {
	runtimeConfig, err := s.settings.ReviewRuntimeConfig(ctx)
	if err != nil {
		return err
	}
	reviewService := review.NewService(s.courseRepo, s.reviewRepo, s.createReviewPolicies(runtimeConfig.FrequencyPolicy))
	now := time.Now()
	rewards := buildCreateReviewRewards(runtimeConfig.Rewards, u.ID, cmd.CourseID, now)
	result, err := reviewService.CreateWithReward(ctx, u, review.CreateReview{
		CourseID: cmd.CourseID,
		Semester: cmd.Semester,
		UserID:   u.ID,
		Rating:   cmd.Rating,
		Content:  cmd.Content,
		Score:    cmd.Score,
		Now:      now,
	}, rewards)
	if err != nil {
		var violation *review.FrequencyViolation
		if errors.As(err, &violation) {
			s.enqueueFrequencyViolationTasks(ctx, violation)
		}
		return err
	}
	for _, rewardID := range result.RewardIDs {
		s.enqueueGrantReward(ctx, rewardID)
	}
	s.enqueueHotCourseActivity(ctx, u.ID, course.HotCourseActivityReviewCreate, cmd.CourseID)
	return nil
}

func (s *ReviewCommandService) createReviewPolicies(config policy.FrequencyPolicyConfig) []review.CreatePolicy {
	policies := make([]review.CreatePolicy, 0, len(s.createPolicies)+1)
	if s.reviewQuery != nil {
		policies = append(policies, policy.NewFrequencyPolicy(s.reviewQuery, config))
	}
	policies = append(policies, s.createPolicies...)
	return policies
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
	reviewService := review.NewService(s.courseRepo, s.reviewRepo, nil)
	err = reviewService.Update(ctx, u, review.UpdateReview{
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
	if u.IsAdmin() && u.ID != existing.UserID {
		audit.EnqueueLog(ctx, audit.Log{
			OccurredAt:  now,
			ActorUserID: u.ID,
			Action:      audit.ActionReviewUpdate,
			TargetType:  audit.TargetTypeReview,
			TargetID:    strconv.Itoa(existing.ID),
			Details: audit.Details{
				"course_id":        existing.CourseID,
				"review_author_id": existing.UserID,
			},
		})
	}
	return nil
}

func (s *ReviewCommandService) DeleteReview(ctx context.Context, u *auth.User, reviewID int) error {
	existing, err := s.reviewRepo.Get(ctx, reviewID)
	if err != nil {
		return err
	}
	if existing == nil {
		return review.ErrReviewNotFound
	}
	reviewService := review.NewService(s.courseRepo, s.reviewRepo, nil)
	if err := reviewService.Delete(ctx, u, reviewID); err != nil {
		return err
	}
	if u.IsAdmin() && u.ID != existing.UserID {
		audit.EnqueueLog(ctx, audit.Log{
			OccurredAt:  time.Now(),
			ActorUserID: u.ID,
			Action:      audit.ActionReviewDelete,
			TargetType:  audit.TargetTypeReview,
			TargetID:    strconv.Itoa(existing.ID),
			Details: audit.Details{
				"course_id":        existing.CourseID,
				"review_author_id": existing.UserID,
			},
		})
	}
	return nil
}

func (s *ReviewCommandService) UpdateModeratorRemark(ctx context.Context, u *auth.User, reviewID int, cmd *UpdateReviewModeratorRemarkCommand) error {
	existing, err := s.reviewRepo.Get(ctx, reviewID)
	if err != nil {
		return err
	}
	if existing == nil {
		return review.ErrReviewNotFound
	}
	reviewService := review.NewService(s.courseRepo, s.reviewRepo, nil)
	if err := reviewService.UpdateModeratorRemark(ctx, u, review.UpdateModeratorRemark{
		ReviewID:        reviewID,
		ModeratorRemark: cmd.ModeratorRemark,
	}); err != nil {
		return err
	}
	if u.IsAdmin() && u.ID != existing.UserID {
		audit.EnqueueLog(ctx, audit.Log{
			OccurredAt:  time.Now(),
			ActorUserID: u.ID,
			Action:      audit.ActionReviewModeratorRemarkEdit,
			TargetType:  audit.TargetTypeReview,
			TargetID:    strconv.Itoa(existing.ID),
			Details: audit.Details{
				"course_id":        existing.CourseID,
				"review_author_id": existing.UserID,
			},
		})
	}
	return nil
}

func (s *ReviewCommandService) VoteReview(ctx context.Context, userID int, reviewID int, voteType int) error {
	runtimeConfig, err := s.settings.ReviewRuntimeConfig(ctx)
	if err != nil {
		return err
	}
	voteService := review.NewVoteService(s.reviewRepo, s.voteRepo, runtimeConfig.Vote)
	now := time.Now()
	result, err := voteService.Vote(ctx, userID, reviewID, voteType, now)
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

func buildCreateReviewRewards(config point.RewardConfig, userID int, courseID int, now time.Time) []point.Reward {
	if !config.Enabled {
		return nil
	}
	reward := point.NewCourseFirstReviewReward(userID, courseID, config.CourseFirstReviewPoints, now)
	if reward == nil {
		return nil
	}
	return []point.Reward{*reward}
}

func (s *ReviewCommandService) enqueueGrantReward(ctx context.Context, rewardID int) {
	if err := task.Enqueue(ctx, point.NewGrantRewardTask(rewardID)); err != nil {
		logx.Warn(ctx, "enqueue grant reward", "reward_id", rewardID, "err", err)
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
