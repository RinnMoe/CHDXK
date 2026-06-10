package application_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"jcourse/internal/application"
	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	domainemail "jcourse/internal/domain/email"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/domain/task"
)

func newFakeCommandReviewRepo() *review.MockReviewRepository {
	return review.NewMockReviewRepository()
}

func newReviewCommandTestServiceWithRuntimeConfig(reviewRepo *review.MockReviewRepository, voteRepo *review.MockVoteRepository, runtimeConfig application.ReviewRuntimeConfig) *application.ReviewCommandService {
	courseRepo := course.NewMockCourseRepository()
	courseRepo.Courses[1] = &course.CourseView{ID: 1, Code: "CS101", LastSemester: "2025-2026-1"}
	courseRepo.OfferedCourses[1] = map[string]bool{"2025-2026-1": true}
	provider := application.NewDefaultSiteSettingsProvider()
	provider.ReviewRuntime = runtimeConfig
	return application.NewReviewCommandService(courseRepo, reviewRepo, voteRepo, provider)
}

type fakeReviewCommandEnqueuer struct {
	tasks []task.Task
}

func (f *fakeReviewCommandEnqueuer) Enqueue(ctx context.Context, t task.Task, opts ...task.EnqueueOption) error {
	f.tasks = append(f.tasks, t)
	return nil
}

func newReviewCommandTestService(reviewRepo *review.MockReviewRepository, voteRepo *review.MockVoteRepository) *application.ReviewCommandService {
	courseRepo := course.NewMockCourseRepository()
	courseRepo.Courses[1] = &course.CourseView{ID: 1, Code: "CS101", LastSemester: "2025-2026-1"}
	courseRepo.OfferedCourses[1] = map[string]bool{"2025-2026-1": true}
	return application.NewReviewCommandService(courseRepo, reviewRepo, voteRepo, nil)
}

type reviewCommandQueryRepo struct {
	*review.MockReviewRepository
	recent []review.ReviewView
}

func (r *reviewCommandQueryRepo) FindBy(ctx context.Context, filter review.ReviewFilter) ([]review.ReviewView, int64, error) {
	return r.recent, int64(len(r.recent)), nil
}

func (r *reviewCommandQueryRepo) GetByID(ctx context.Context, reviewID int) (*review.ReviewView, error) {
	return nil, nil
}

func (r *reviewCommandQueryRepo) GetCourseFilters(ctx context.Context, courseID int) (*review.ReviewFilters, error) {
	return &review.ReviewFilters{}, nil
}

func (r *reviewCommandQueryRepo) GetCourseTrend(ctx context.Context, courseID int) ([]review.ReviewTrendItem, error) {
	return nil, nil
}

func (r *reviewCommandQueryRepo) FindRevisions(ctx context.Context, reviewID int) ([]review.RevisionView, error) {
	return nil, nil
}

func TestReviewCommandService_CreateReviewEnqueuesHotCourseActivity(t *testing.T) {
	reviewRepo := newFakeCommandReviewRepo()
	enqueuer := &fakeReviewCommandEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	svc := newReviewCommandTestService(reviewRepo, &review.MockVoteRepository{})

	err := svc.CreateReview(context.Background(), &auth.User{ID: 10}, &application.CreateReviewCommand{
		CourseID: 1,
		Semester: "2025-2026-1",
		Rating:   5,
		Content:  "good course",
	})
	if err != nil {
		t.Fatalf("CreateReview: %v", err)
	}
	if len(enqueuer.tasks) != 1 {
		t.Fatalf("tasks = %d, want 1", len(enqueuer.tasks))
	}
	if got := enqueuer.tasks[0].Type(); got != course.TaskTypeRecordHotCourseActivity {
		t.Fatalf("task type = %q, want %q", got, course.TaskTypeRecordHotCourseActivity)
	}
	var payload course.RecordHotCourseActivityPayload
	if err := json.Unmarshal(enqueuer.tasks[0].Payload(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	want := course.RecordHotCourseActivityPayload{UserID: 10, Activity: course.HotCourseActivityReviewCreate, CourseID: 1}
	if payload != want {
		t.Fatalf("payload = %+v, want %+v", payload, want)
	}
}

func TestReviewCommandService_CreateReviewEnqueuesGrantRewardWhenEnabled(t *testing.T) {
	reviewRepo := newFakeCommandReviewRepo()
	enqueuer := &fakeReviewCommandEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	svc := newReviewCommandTestServiceWithRuntimeConfig(reviewRepo, &review.MockVoteRepository{}, application.ReviewRuntimeConfig{
		Vote:            review.DefaultVoteConfig,
		FrequencyPolicy: policy.DefaultFrequencyPolicyConfig,
		Rewards: point.RewardConfig{
			CourseFirstReviewEnabled: true,
			CourseFirstReviewPoints:  10,
		},
	})

	err := svc.CreateReview(context.Background(), &auth.User{ID: 10}, &application.CreateReviewCommand{
		CourseID: 1,
		Semester: "2025-2026-1",
		Rating:   5,
		Content:  "good course",
	})
	if err != nil {
		t.Fatalf("CreateReview: %v", err)
	}
	if len(enqueuer.tasks) != 2 {
		t.Fatalf("tasks = %d, want 2", len(enqueuer.tasks))
	}
	if got := enqueuer.tasks[0].Type(); got != point.TaskTypeGrantReward {
		t.Fatalf("first task type = %q, want %q", got, point.TaskTypeGrantReward)
	}
	var rewardPayload point.GrantRewardPayload
	if err := json.Unmarshal(enqueuer.tasks[0].Payload(), &rewardPayload); err != nil {
		t.Fatalf("unmarshal reward payload: %v", err)
	}
	if rewardPayload.RewardID == 0 {
		t.Fatal("reward payload ID was not set")
	}
	if got := enqueuer.tasks[1].Type(); got != course.TaskTypeRecordHotCourseActivity {
		t.Fatalf("second task type = %q, want %q", got, course.TaskTypeRecordHotCourseActivity)
	}
}

func TestReviewCommandService_CreateReviewEnqueuesReviewCreateRewardWhenEnabled(t *testing.T) {
	reviewRepo := newFakeCommandReviewRepo()
	reviewRepo.OnCreateWithReward = func(ctx context.Context, rv *review.Review, rewards []point.Reward) (review.CreateResult, error) {
		if len(rewards) != 1 {
			t.Fatalf("rewards = %d, want 1", len(rewards))
		}
		got := rewards[0]
		if got.Reason != point.RewardReasonReviewCreate || got.Amount != 1 || got.SourceType != point.RewardSourceTypeReview || got.Description != "发布点评奖励" {
			t.Fatalf("reward = %+v", got)
		}
		if err := reviewRepo.Create(ctx, rv); err != nil {
			return review.CreateResult{}, err
		}
		return review.CreateResult{RewardIDs: []int{1001}}, nil
	}
	enqueuer := &fakeReviewCommandEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	svc := newReviewCommandTestServiceWithRuntimeConfig(reviewRepo, &review.MockVoteRepository{}, application.ReviewRuntimeConfig{
		Vote:            review.DefaultVoteConfig,
		FrequencyPolicy: policy.DefaultFrequencyPolicyConfig,
		Rewards: point.RewardConfig{
			ReviewCreateEnabled: true,
			ReviewCreatePoints:  1,
		},
	})

	err := svc.CreateReview(context.Background(), &auth.User{ID: 10}, &application.CreateReviewCommand{
		CourseID: 1,
		Semester: "2025-2026-1",
		Rating:   5,
		Content:  "good course",
	})
	if err != nil {
		t.Fatalf("CreateReview: %v", err)
	}
	if len(enqueuer.tasks) != 2 {
		t.Fatalf("tasks = %d, want reward and hot course", len(enqueuer.tasks))
	}
	if got := enqueuer.tasks[0].Type(); got != point.TaskTypeGrantReward {
		t.Fatalf("first task type = %q, want %q", got, point.TaskTypeGrantReward)
	}
	if got := enqueuer.tasks[1].Type(); got != course.TaskTypeRecordHotCourseActivity {
		t.Fatalf("second task type = %q, want %q", got, course.TaskTypeRecordHotCourseActivity)
	}
}

func TestReviewCommandService_CreateReviewBuildsBothRewardsWhenEnabled(t *testing.T) {
	reviewRepo := newFakeCommandReviewRepo()
	reviewRepo.OnCreateWithReward = func(ctx context.Context, rv *review.Review, rewards []point.Reward) (review.CreateResult, error) {
		if len(rewards) != 2 {
			t.Fatalf("rewards = %d, want 2", len(rewards))
		}
		if rewards[0].Reason != point.RewardReasonCourseFirstReview {
			t.Fatalf("first reward reason = %q, want %q", rewards[0].Reason, point.RewardReasonCourseFirstReview)
		}
		if rewards[1].Reason != point.RewardReasonReviewCreate {
			t.Fatalf("second reward reason = %q, want %q", rewards[1].Reason, point.RewardReasonReviewCreate)
		}
		if err := reviewRepo.Create(ctx, rv); err != nil {
			return review.CreateResult{}, err
		}
		return review.CreateResult{RewardIDs: []int{1001, 1002}}, nil
	}
	enqueuer := &fakeReviewCommandEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	svc := newReviewCommandTestServiceWithRuntimeConfig(reviewRepo, &review.MockVoteRepository{}, application.ReviewRuntimeConfig{
		Vote:            review.DefaultVoteConfig,
		FrequencyPolicy: policy.DefaultFrequencyPolicyConfig,
		Rewards: point.RewardConfig{
			CourseFirstReviewEnabled: true,
			CourseFirstReviewPoints:  10,
			ReviewCreateEnabled:      true,
			ReviewCreatePoints:       1,
		},
	})

	err := svc.CreateReview(context.Background(), &auth.User{ID: 10}, &application.CreateReviewCommand{
		CourseID: 1,
		Semester: "2025-2026-1",
		Rating:   5,
		Content:  "good course",
	})
	if err != nil {
		t.Fatalf("CreateReview: %v", err)
	}
	if len(enqueuer.tasks) != 3 {
		t.Fatalf("tasks = %d, want 3", len(enqueuer.tasks))
	}
}

func TestReviewCommandService_CreateReviewDoesNotEnqueueRewardWhenConflict(t *testing.T) {
	reviewRepo := newFakeCommandReviewRepo()
	reviewRepo.OnCreateWithReward = func(ctx context.Context, rv *review.Review, rewards []point.Reward) (review.CreateResult, error) {
		if err := reviewRepo.Create(ctx, rv); err != nil {
			return review.CreateResult{}, err
		}
		return review.CreateResult{}, nil
	}
	enqueuer := &fakeReviewCommandEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	svc := newReviewCommandTestServiceWithRuntimeConfig(reviewRepo, &review.MockVoteRepository{}, application.ReviewRuntimeConfig{
		Vote:            review.DefaultVoteConfig,
		FrequencyPolicy: policy.DefaultFrequencyPolicyConfig,
		Rewards: point.RewardConfig{
			CourseFirstReviewEnabled: true,
			CourseFirstReviewPoints:  10,
		},
	})

	err := svc.CreateReview(context.Background(), &auth.User{ID: 10}, &application.CreateReviewCommand{
		CourseID: 1,
		Semester: "2025-2026-1",
		Rating:   5,
		Content:  "good course",
	})
	if err != nil {
		t.Fatalf("CreateReview: %v", err)
	}
	if len(enqueuer.tasks) != 1 {
		t.Fatalf("tasks = %d, want hot course only", len(enqueuer.tasks))
	}
	if got := enqueuer.tasks[0].Type(); got != course.TaskTypeRecordHotCourseActivity {
		t.Fatalf("task type = %q, want %q", got, course.TaskTypeRecordHotCourseActivity)
	}
}

func TestReviewCommandService_CreateReviewEnqueuesFrequencyViolationTasks(t *testing.T) {
	reviewRepo := &reviewCommandQueryRepo{
		MockReviewRepository: newFakeCommandReviewRepo(),
		recent: []review.ReviewView{
			{CourseID: 1, Content: "previous 1", Course: &course.CourseView{ID: 1, Code: "CS101"}},
			{CourseID: 1, Content: "previous 2", Course: &course.CourseView{ID: 1, Code: "CS101"}},
			{CourseID: 1, Content: "previous 3", Course: &course.CourseView{ID: 1, Code: "CS101"}},
		},
	}
	enqueuer := &fakeReviewCommandEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	duration := 2 * time.Hour
	courseRepo := course.NewMockCourseRepository()
	courseRepo.Courses[1] = &course.CourseView{ID: 1, Code: "CS101", Name: "Intro CS", LastSemester: "2025-2026-1"}
	courseRepo.OfferedCourses[1] = map[string]bool{"2025-2026-1": true}
	provider := application.NewDefaultSiteSettingsProvider()
	provider.ReviewRuntime.FrequencyPolicy = policy.FrequencyPolicyConfig{
		Window:          time.Hour,
		MaxReviews:      3,
		SimilarityRatio: 0.9,
		SuspendDuration: duration,
	}
	provider.ReviewRuntime.FrequencyViolationAdminEmails = []string{"admin@example.edu"}
	svc := application.NewReviewCommandService(courseRepo, reviewRepo, &review.MockVoteRepository{}, provider)

	err := svc.CreateReview(context.Background(), &auth.User{ID: 10}, &application.CreateReviewCommand{
		CourseID: 1,
		Semester: "2025-2026-1",
		Rating:   5,
		Content:  "spam content",
	})
	if !errors.Is(err, policy.ErrSameCourseSpam) {
		t.Fatalf("CreateReview error = %v, want %v", err, policy.ErrSameCourseSpam)
	}
	if len(enqueuer.tasks) != 2 {
		t.Fatalf("tasks = %d, want 2", len(enqueuer.tasks))
	}
	assertSuspendTask(t, enqueuer.tasks[0], 10, duration)
	assertEmailTask(t, enqueuer.tasks[1], "admin@example.edu", "10", "CS101", "Intro CS", "spam content", duration.String())
}

func TestReviewCommandService_UpdateReviewEnqueuesHotCourseActivity(t *testing.T) {
	reviewRepo := newFakeCommandReviewRepo()
	reviewRepo.Reviews[1] = &review.Review{ID: 1, CourseID: 1, UserID: 10, Semester: "2025-2026-1", Rating: 4, Content: "old"}
	enqueuer := &fakeReviewCommandEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	svc := newReviewCommandTestService(reviewRepo, &review.MockVoteRepository{})

	err := svc.UpdateReview(context.Background(), &auth.User{ID: 10}, &application.UpdateReviewCommand{
		ReviewID: 1,
		Semester: "2025-2026-1",
		Rating:   5,
		Content:  "updated",
	})
	if err != nil {
		t.Fatalf("UpdateReview: %v", err)
	}
	if len(enqueuer.tasks) != 1 {
		t.Fatalf("tasks = %d, want 1", len(enqueuer.tasks))
	}
	var payload course.RecordHotCourseActivityPayload
	if err := json.Unmarshal(enqueuer.tasks[0].Payload(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	want := course.RecordHotCourseActivityPayload{UserID: 10, Activity: course.HotCourseActivityReviewUpdate, CourseID: 1}
	if payload != want {
		t.Fatalf("payload = %+v, want %+v", payload, want)
	}
}

func TestReviewCommandService_DeleteReviewEnqueuesRewardRevocationForAdminDelete(t *testing.T) {
	reviewRepo := newFakeCommandReviewRepo()
	reviewRepo.Reviews[1] = &review.Review{ID: 1, CourseID: 2, UserID: 10, Semester: "2025-2026-1", Rating: 4, Content: "old"}
	enqueuer := &fakeReviewCommandEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	svc := newReviewCommandTestService(reviewRepo, &review.MockVoteRepository{})

	err := svc.DeleteReview(context.Background(), &auth.User{ID: 99, Role: auth.RoleAdmin}, 1)
	if err != nil {
		t.Fatalf("DeleteReview: %v", err)
	}
	if len(enqueuer.tasks) != 2 {
		t.Fatalf("tasks = %d, want revoke rewards and audit log", len(enqueuer.tasks))
	}
	if got := enqueuer.tasks[0].Type(); got != point.TaskTypeRevokeReviewRewardsByID {
		t.Fatalf("first task type = %q, want %q", got, point.TaskTypeRevokeReviewRewardsByID)
	}
	var revokePayload point.RevokeReviewRewardsByIDPayload
	if err := json.Unmarshal(enqueuer.tasks[0].Payload(), &revokePayload); err != nil {
		t.Fatalf("unmarshal revoke payload: %v", err)
	}
	wantRevoke := point.RevokeReviewRewardsByIDPayload{ReviewID: 1, CourseID: 2, AuthorUserID: 10}
	if revokePayload != wantRevoke {
		t.Fatalf("revoke payload = %+v, want %+v", revokePayload, wantRevoke)
	}

	if got := enqueuer.tasks[1].Type(); got != audit.TaskTypeRecordLog {
		t.Fatalf("second task type = %q, want %q", got, audit.TaskTypeRecordLog)
	}
	var auditPayload audit.RecordLogPayload
	if err := json.Unmarshal(enqueuer.tasks[1].Payload(), &auditPayload); err != nil {
		t.Fatalf("unmarshal audit payload: %v", err)
	}
	if auditPayload.Action != audit.ActionReviewDelete || auditPayload.ActorUserID != 99 || auditPayload.TargetID != "1" {
		t.Fatalf("audit payload = %+v", auditPayload)
	}
}

func TestReviewCommandService_DeleteReviewRevokesRewardsWithoutAuditForAdminDeletingOwnReview(t *testing.T) {
	reviewRepo := newFakeCommandReviewRepo()
	reviewRepo.Reviews[1] = &review.Review{ID: 1, CourseID: 2, UserID: 99, Semester: "2025-2026-1", Rating: 4, Content: "old"}
	enqueuer := &fakeReviewCommandEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	svc := newReviewCommandTestService(reviewRepo, &review.MockVoteRepository{})

	err := svc.DeleteReview(context.Background(), &auth.User{ID: 99, Role: auth.RoleAdmin}, 1)
	if err != nil {
		t.Fatalf("DeleteReview: %v", err)
	}
	if len(enqueuer.tasks) != 1 {
		t.Fatalf("tasks = %d, want revoke rewards only", len(enqueuer.tasks))
	}
	if got := enqueuer.tasks[0].Type(); got != point.TaskTypeRevokeReviewRewardsByID {
		t.Fatalf("task type = %q, want %q", got, point.TaskTypeRevokeReviewRewardsByID)
	}
	var payload point.RevokeReviewRewardsByIDPayload
	if err := json.Unmarshal(enqueuer.tasks[0].Payload(), &payload); err != nil {
		t.Fatalf("unmarshal revoke payload: %v", err)
	}
	want := point.RevokeReviewRewardsByIDPayload{ReviewID: 1, CourseID: 2, AuthorUserID: 99}
	if payload != want {
		t.Fatalf("payload = %+v, want %+v", payload, want)
	}
}

func TestReviewCommandService_UpdateModeratorRemarkRequiresAdmin(t *testing.T) {
	reviewRepo := newFakeCommandReviewRepo()
	reviewRepo.Reviews[1] = &review.Review{ID: 1, CourseID: 1, UserID: 10, Semester: "2025-2026-1", Rating: 4, Content: "old"}
	svc := newReviewCommandTestService(reviewRepo, &review.MockVoteRepository{})

	err := svc.UpdateModeratorRemark(context.Background(), &auth.User{ID: 10, Role: auth.RoleUser}, 1, &application.UpdateReviewModeratorRemarkCommand{
		ModeratorRemark: "需要补充依据",
	})
	if err != review.ErrUserCannotModerate {
		t.Fatalf("non-admin UpdateModeratorRemark error = %v, want %v", err, review.ErrUserCannotModerate)
	}

	err = svc.UpdateModeratorRemark(context.Background(), &auth.User{ID: 99, Role: auth.RoleAdmin}, 1, &application.UpdateReviewModeratorRemarkCommand{
		ModeratorRemark: "已核实",
	})
	if err != nil {
		t.Fatalf("admin UpdateModeratorRemark: %v", err)
	}
	if got := reviewRepo.Reviews[1].ModeratorRemark; got != "已核实" {
		t.Fatalf("ModeratorRemark = %q, want 已核实", got)
	}
}

func TestReviewCommandService_VoteReviewRecordsOnlyChangedVote(t *testing.T) {
	reviewRepo := newFakeCommandReviewRepo()
	reviewRepo.Reviews[1] = &review.Review{ID: 1, CourseID: 1, UserID: 20, Semester: "2025-2026-1", Rating: 4, Content: "ok"}
	voteRepo := &review.MockVoteRepository{Existing: &review.Vote{ReviewID: 1, UserID: 10, VoteType: review.VoteLike}}
	enqueuer := &fakeReviewCommandEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	svc := newReviewCommandTestService(reviewRepo, voteRepo)

	if err := svc.VoteReview(context.Background(), 10, 1, review.VoteLike); err != nil {
		t.Fatalf("VoteReview duplicate: %v", err)
	}
	if len(enqueuer.tasks) != 0 {
		t.Fatalf("duplicate vote tasks = %d, want none", len(enqueuer.tasks))
	}

	if err := svc.VoteReview(context.Background(), 10, 1, review.VoteDislike); err != nil {
		t.Fatalf("VoteReview changed: %v", err)
	}
	if len(enqueuer.tasks) != 1 {
		t.Fatalf("changed vote tasks = %d, want 1", len(enqueuer.tasks))
	}
	var payload course.RecordHotCourseActivityPayload
	if err := json.Unmarshal(enqueuer.tasks[0].Payload(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	want := course.RecordHotCourseActivityPayload{UserID: 10, Activity: course.HotCourseActivityReviewVote, CourseID: 1}
	if payload != want {
		t.Fatalf("payload = %+v, want %+v", payload, want)
	}
}

func TestCourseHotService_RecordActivityUsesConfiguredScore(t *testing.T) {
	hotRepo := &course.MockHotCourseRepository{}
	svc := course.NewCourseHotService(hotRepo)
	scores := course.HotScoreConfig{
		ReviewCreateScore: 5,
		ReviewUpdateScore: 2,
		ReviewVoteScore:   1,
	}

	err := svc.RecordActivityWithScores(context.Background(), course.RecordHotCourseActivityPayload{
		UserID:   10,
		Activity: course.HotCourseActivityReviewUpdate,
		CourseID: 1,
	}, scores)
	if err != nil {
		t.Fatalf("RecordActivity: %v", err)
	}
	if len(hotRepo.Calls) != 1 || hotRepo.Calls[0].CourseID != 1 || hotRepo.Calls[0].Score != 2 {
		t.Fatalf("hot calls = %+v, want course=1 score=2", hotRepo.Calls)
	}
}

func assertSuspendTask(t *testing.T, taskItem task.Task, userID int, duration time.Duration) {
	t.Helper()
	if got := taskItem.Type(); got != auth.TaskTypeSuspendUser {
		t.Fatalf("task type = %q, want %q", got, auth.TaskTypeSuspendUser)
	}
	var payload auth.SuspendUserPayload
	if err := json.Unmarshal(taskItem.Payload(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.UserID != userID || payload.Duration != duration {
		t.Fatalf("payload = %+v, want userID=%d duration=%s", payload, userID, duration)
	}
}

func assertEmailTask(t *testing.T, taskItem task.Task, to string, userID string, courseCode string, courseName string, reviewContent string, duration string) {
	t.Helper()
	if got := taskItem.Type(); got != domainemail.TaskTypeSendEmail {
		t.Fatalf("task type = %q, want %q", got, domainemail.TaskTypeSendEmail)
	}
	var payload domainemail.SendEmailPayload
	if err := json.Unmarshal(taskItem.Payload(), &payload); err != nil {
		t.Fatalf("unmarshal email payload: %v", err)
	}
	if payload.EmailType != review.SpamSuspensionEmailType {
		t.Fatalf("email type = %q, want %q", payload.EmailType, review.SpamSuspensionEmailType)
	}
	if payload.Email.To != to || payload.Email.Subject == "" || !strings.Contains(payload.Email.Body, userID) || !strings.Contains(payload.Email.Body, courseCode) || !strings.Contains(payload.Email.Body, courseName) || !strings.Contains(payload.Email.Body, reviewContent) || !strings.Contains(payload.Email.Body, duration) {
		t.Fatalf("email payload = %+v", payload)
	}
}
