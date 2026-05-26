package application_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	domainemail "jcourse/internal/domain/email"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/domain/task"
)

func newFakeCommandReviewRepo() *review.MockReviewRepository {
	return review.NewMockReviewRepository()
}

type fakeReviewCommandEnqueuer struct {
	tasks []task.Task
}

func (f *fakeReviewCommandEnqueuer) Enqueue(ctx context.Context, t task.Task, opts ...task.EnqueueOption) error {
	f.tasks = append(f.tasks, t)
	return nil
}

func newReviewCommandTestService(reviewRepo *review.MockReviewRepository, voteRepo *review.MockVoteRepository) *application.ReviewCommandService {
	return newReviewCommandTestServiceWithPolicies(reviewRepo, voteRepo, nil)
}

func newReviewCommandTestServiceWithPolicies(reviewRepo *review.MockReviewRepository, voteRepo *review.MockVoteRepository, policies []review.CreatePolicy) *application.ReviewCommandService {
	courseRepo := course.NewMockCourseRepository()
	courseRepo.Courses[1] = &course.CourseView{ID: 1, Code: "CS101", LastSemester: "2025-2026-1"}
	courseRepo.OfferedCourses[1] = map[string]bool{"2025-2026-1": true}
	return application.NewReviewCommandService(
		courseRepo,
		reviewRepo,
		voteRepo,
		application.ReviewCommandConfig{
			HotScores: course.HotScoreConfig{
				ReviewCreateScore: 5,
				ReviewUpdateScore: 2,
				ReviewVoteScore:   1,
			},
			Vote:                          review.DefaultVoteConfig,
			FrequencyViolationAdminEmails: []string{"admin@example.edu"},
		},
		policies,
	)
}

type rejectCreatePolicy struct {
	err error
}

func (p rejectCreatePolicy) CanCreate(ctx context.Context, u *auth.User, c *course.CourseView, r *review.Review) error {
	return p.err
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

func TestReviewCommandService_CreateReviewEnqueuesFrequencyViolationTasks(t *testing.T) {
	reviewRepo := newFakeCommandReviewRepo()
	enqueuer := &fakeReviewCommandEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	duration := 2 * time.Hour
	violation := &review.FrequencyViolation{
		Reason:          policy.ErrSameCourseSpam,
		Review:          &review.Review{UserID: 10, CourseID: 1, Content: "spam content"},
		Course:          &course.CourseView{ID: 1, Code: "CS101", Name: "Intro CS"},
		SuspendDuration: duration,
	}
	svc := newReviewCommandTestServiceWithPolicies(reviewRepo, &review.MockVoteRepository{}, []review.CreatePolicy{
		rejectCreatePolicy{err: violation},
	})

	err := svc.CreateReview(context.Background(), &auth.User{ID: 10}, &application.CreateReviewCommand{
		CourseID: 1,
		Semester: "2025-2026-1",
		Rating:   5,
		Content:  "spam",
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
	svc := course.NewCourseHotService(hotRepo, course.HotScoreConfig{
		ReviewCreateScore: 5,
		ReviewUpdateScore: 2,
		ReviewVoteScore:   1,
	})

	err := svc.RecordActivity(context.Background(), course.RecordHotCourseActivityPayload{
		UserID:   10,
		Activity: course.HotCourseActivityReviewUpdate,
		CourseID: 1,
	})
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
	if payload.Email.To != to || payload.Email.Subject == "" || !strings.Contains(payload.Email.Body, userID) || !strings.Contains(payload.Email.Body, courseCode) || !strings.Contains(payload.Email.Body, courseName) || !strings.Contains(payload.Email.Body, reviewContent) || !strings.Contains(payload.Email.Body, duration) {
		t.Fatalf("email payload = %+v", payload)
	}
}
