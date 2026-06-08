package course

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/oauth2"

	"jcourse/internal/domain/jaccount"
)

func newFakeEnrollmentCourseRepo() *MockCourseRepository {
	repo := NewMockCourseRepository()
	repo.OnOfferedCourseExists = func(_ context.Context, _ int, semester string) (bool, error) {
		return repo.OfferedSemesters[semester], nil
	}
	return repo
}

func TestEnrollmentService_StartSyncReturnsAuthURL(t *testing.T) {
	courseRepo := newFakeEnrollmentCourseRepo()
	courseRepo.OfferedSemesters["2025-2026-1"] = true
	client := &fakeJAccountClient{authURL: "https://jaccount.example/auth"}
	svc := NewEnrollmentService(courseRepo, &MockCourseEnrollmentRepository{}, client)

	semester, authURL, err := svc.StartSync(context.Background(), " 2025-2026-1 ", "sync-state")
	if err != nil {
		t.Fatalf("StartSync: %v", err)
	}
	if semester != "2025-2026-1" || authURL != client.authURL || client.state != "sync-state" {
		t.Fatalf("semester/authURL/state = %q/%q/%q", semester, authURL, client.state)
	}
}

func TestEnrollmentService_StartSyncRejectsInvalidSemester(t *testing.T) {
	svc := NewEnrollmentService(newFakeEnrollmentCourseRepo(), &MockCourseEnrollmentRepository{}, &fakeJAccountClient{})

	_, _, err := svc.StartSync(context.Background(), "2024-2025-2", "sync-state")
	if !errors.Is(err, ErrInvalidSyncSemester) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidSyncSemester)
	}
}

func TestEnrollmentService_SyncLessonsSkipsIncompleteLessons(t *testing.T) {
	enrollmentRepo := &MockCourseEnrollmentRepository{SyncCount: 2}
	svc := NewEnrollmentService(newFakeEnrollmentCourseRepo(), enrollmentRepo, nil)

	result, err := svc.SyncLessons(context.Background(), 10, " 2025-2026-1 ", []jaccount.LessonCourse{
		{Code: " CS101 ", TeacherName: " Alice "},
		{Code: "", TeacherName: "Bob"},
		{Code: "MA101", TeacherName: " "},
		{Code: " PH101 ", TeacherName: " Carol "},
	})
	if err != nil {
		t.Fatalf("SyncLessons: %v", err)
	}
	if result.Matched != 2 || result.Total != 2 {
		t.Fatalf("result = %+v", result)
	}
	if enrollmentRepo.SyncedUserID != 10 || enrollmentRepo.SyncedSemester != "2025-2026-1" {
		t.Fatalf("synced user/semester = %d/%q", enrollmentRepo.SyncedUserID, enrollmentRepo.SyncedSemester)
	}
	want := []CourseCodeTeacher{{Code: "CS101", TeacherName: "Alice"}, {Code: "PH101", TeacherName: "Carol"}}
	if len(enrollmentRepo.SyncedPairs) != len(want) {
		t.Fatalf("synced pairs = %+v", enrollmentRepo.SyncedPairs)
	}
	for i := range want {
		if enrollmentRepo.SyncedPairs[i] != want[i] {
			t.Fatalf("synced pairs = %+v, want %+v", enrollmentRepo.SyncedPairs, want)
		}
	}
}

type fakeJAccountClient struct {
	authURL string
	state   string
}

func (c *fakeJAccountClient) AuthCodeURL(state string) (string, error) {
	c.state = state
	return c.authURL, nil
}

func (c *fakeJAccountClient) Exchange(context.Context, string) (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

func (c *fakeJAccountClient) Lessons(context.Context, *oauth2.Token, string) ([]jaccount.LessonCourse, error) {
	return nil, nil
}
