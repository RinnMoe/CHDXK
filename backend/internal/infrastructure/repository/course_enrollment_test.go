package repository_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/course"
	"jcourse/internal/infrastructure/repository"
)

func TestCourseEnrollmentRepository_CreateFindDelete(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseEnrollmentRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "course_enrollments", "courses", "teachers", "users")
	teacher := seedTeacher(t, db)
	courseEntity := seedCourse(t, db, teacher.ID)
	user := seedUser(t, db)

	enrollment := &course.CourseEnrollment{
		UserID:    user.ID,
		CourseID:  courseEntity.ID,
		Semester:  "2024-2025-1",
		CreatedAt: time.Now(),
	}
	if err := repo.Create(ctx, enrollment); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if enrollment.ID == 0 {
		t.Fatal("Create should populate id")
	}

	duplicate := &course.CourseEnrollment{
		UserID:    user.ID,
		CourseID:  courseEntity.ID,
		Semester:  "2024-2025-1",
		CreatedAt: time.Now(),
	}
	if err := repo.Create(ctx, duplicate); err != nil {
		t.Fatalf("Create duplicate: %v", err)
	}

	otherSemester := &course.CourseEnrollment{
		UserID:    user.ID,
		CourseID:  courseEntity.ID,
		Semester:  "2024-2025-2",
		CreatedAt: time.Now(),
	}
	if err := repo.Create(ctx, otherSemester); err != nil {
		t.Fatalf("Create other semester: %v", err)
	}

	rows, err := repo.FindUserEnrollments(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindEnrollments: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("enrollments: got len=%d, want 2", len(rows))
	}
	if rows[0].Course.ID != courseEntity.ID || rows[0].Course.MainTeacher == nil {
		t.Fatalf("course view not loaded: %+v", rows[0].Course)
	}

	if err := repo.Delete(ctx, enrollment.ID, user.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	rows, err = repo.FindUserEnrollments(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindEnrollments after delete: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("count after delete: got %d, want 1", len(rows))
	}

	otherUser := seedUserRaw(t, db, "enrollment-other", "enrollment-other@example.com")
	if err := repo.Delete(ctx, otherSemester.ID, otherUser.ID); err != nil {
		t.Fatalf("Delete as other user: %v", err)
	}
	rows, err = repo.FindUserEnrollments(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindEnrollments after wrong user delete: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("wrong user delete should not remove row, count=%d", len(rows))
	}
}
