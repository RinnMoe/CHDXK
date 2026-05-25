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

type fakeCommandCourseRepo struct {
	courses map[int]*course.Course
}

func (r *fakeCommandCourseRepo) Get(ctx context.Context, courseID int) (*course.Course, error) {
	c, ok := r.courses[courseID]
	if !ok {
		return nil, nil
	}
	copy := *c
	return &copy, nil
}

func (r *fakeCommandCourseRepo) OfferedCourseExists(ctx context.Context, courseID int, semester string) (bool, error) {
	_, ok := r.courses[courseID]
	return ok, nil
}

func (r *fakeCommandCourseRepo) OfferedSemesterExists(ctx context.Context, semester string) (bool, error) {
	return false, nil
}

type fakeCommandReviewRepo struct {
	nextID  int
	reviews map[int]*review.Review
}

func newFakeCommandReviewRepo() *fakeCommandReviewRepo {
	return &fakeCommandReviewRepo{nextID: 1, reviews: map[int]*review.Review{}}
}

func (r *fakeCommandReviewRepo) Create(ctx context.Context, rv *review.Review) error {
	copy := *rv
	copy.ID = r.nextID
	r.nextID++
	r.reviews[copy.ID] = &copy
	rv.ID = copy.ID
	return nil
}

func (r *fakeCommandReviewRepo) Update(ctx context.Context, rv *review.Review, revision review.Revision) error {
	copy := *rv
	r.reviews[copy.ID] = &copy
	return nil
}

func (r *fakeCommandReviewRepo) UpdateModeratorRemark(ctx context.Context, reviewID int, moderatorRemark string) error {
	rv, ok := r.reviews[reviewID]
	if !ok {
		return nil
	}
	copy := *rv
	copy.ModeratorRemark = moderatorRemark
	r.reviews[reviewID] = &copy
	return nil
}

func (r *fakeCommandReviewRepo) Delete(ctx context.Context, rv *review.Review) error {
	delete(r.reviews, rv.ID)
	return nil
}

func (r *fakeCommandReviewRepo) Get(ctx context.Context, reviewID int) (*review.Review, error) {
	rv, ok := r.reviews[reviewID]
	if !ok {
		return nil, nil
	}
	copy := *rv
	return &copy, nil
}

type fakeCommandVoteRepo struct {
	existing   *review.Vote
	todayCount int64
}

func (r *fakeCommandVoteRepo) FindByReviewAndUser(ctx context.Context, reviewID, userID int) (*review.Vote, error) {
	if r.existing == nil {
		return nil, nil
	}
	copy := *r.existing
	return &copy, nil
}

func (r *fakeCommandVoteRepo) FindByReviewsAndUser(ctx context.Context, reviewIDs []int, userID int) (map[int]review.Vote, error) {
	return map[int]review.Vote{}, nil
}

func (r *fakeCommandVoteRepo) CountTodayByUser(ctx context.Context, userID int) (int64, error) {
	return r.todayCount, nil
}

func (r *fakeCommandVoteRepo) Save(ctx context.Context, vote *review.Vote) error {
	copy := *vote
	r.existing = &copy
	return nil
}

func (r *fakeCommandVoteRepo) Delete(ctx context.Context, reviewID, userID int) error {
	r.existing = nil
	return nil
}

type fakeHotScoreRepo struct {
	calls []course.HotCourseRank
}

func (r *fakeHotScoreRepo) AddScore(ctx context.Context, courseID int, score int64, periods ...course.HotCoursePeriod) error {
	r.calls = append(r.calls, course.HotCourseRank{CourseID: courseID, Score: score})
	return nil
}

func (r *fakeHotScoreRepo) Top(ctx context.Context, period course.HotCoursePeriod, limit int64) ([]course.HotCourseRank, error) {
	return nil, nil
}

type fakeReviewCommandEnqueuer struct {
	tasks []task.Task
}

func (f *fakeReviewCommandEnqueuer) Enqueue(ctx context.Context, t task.Task, opts ...task.EnqueueOption) error {
	f.tasks = append(f.tasks, t)
	return nil
}

func newReviewCommandTestService(reviewRepo *fakeCommandReviewRepo, voteRepo *fakeCommandVoteRepo) *application.ReviewCommandService {
	courseRepo := &fakeCommandCourseRepo{courses: map[int]*course.Course{
		1: &course.Course{ID: 1},
	}}
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
	svc := newReviewCommandTestService(reviewRepo, &fakeCommandVoteRepo{})

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
	reviewRepo.reviews[1] = &review.Review{ID: 1, CourseID: 1, UserID: 10, Semester: "2025-2026-1", Rating: 4, Content: "old"}
	enqueuer := &fakeReviewCommandEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	svc := newReviewCommandTestService(reviewRepo, &fakeCommandVoteRepo{})

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
	reviewRepo.reviews[1] = &review.Review{ID: 1, CourseID: 1, UserID: 10, Semester: "2025-2026-1", Rating: 4, Content: "old"}
	svc := newReviewCommandTestService(reviewRepo, &fakeCommandVoteRepo{})

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
	if got := reviewRepo.reviews[1].ModeratorRemark; got != "已核实" {
		t.Fatalf("ModeratorRemark = %q, want 已核实", got)
	}
}

func TestReviewCommandService_VoteReviewRecordsOnlyChangedVote(t *testing.T) {
	reviewRepo := newFakeCommandReviewRepo()
	reviewRepo.reviews[1] = &review.Review{ID: 1, CourseID: 1, UserID: 20, Semester: "2025-2026-1", Rating: 4, Content: "ok"}
	voteRepo := &fakeCommandVoteRepo{existing: &review.Vote{ReviewID: 1, UserID: 10, VoteType: review.VoteLike}}
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
	hotRepo := &fakeHotScoreRepo{}
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
	if len(hotRepo.calls) != 1 || hotRepo.calls[0].CourseID != 1 || hotRepo.calls[0].Score != 2 {
		t.Fatalf("hot calls = %+v, want course=1 score=2", hotRepo.calls)
	}
}
