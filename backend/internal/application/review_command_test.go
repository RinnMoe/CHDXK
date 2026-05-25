package application_test

import (
	"context"
	"encoding/json"
	"testing"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
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
	courseRepo := course.NewMockCourseRepository()
	courseRepo.Courses[1] = &course.Course{ID: 1, LastSemester: "2025-2026-1"}
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
			Vote: review.DefaultVoteConfig,
		},
		nil,
	)
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

func TestCourseHotCommandService_RecordActivityUsesConfiguredScore(t *testing.T) {
	hotRepo := &course.MockHotCourseRepository{}
	svc := application.NewCourseHotCommandService(hotRepo, course.HotScoreConfig{
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
