package repository_test

import (
	"context"
	"jcourse/internal/domain/course"
	"jcourse/internal/infrastructure/repository"
	"testing"
)

func TestEnrollmentImportMatchesSemesterSectionAndPreservesOtherUsers(t *testing.T) {
	db := newTestDB(t)
	teacher := seedTeacher(t, db)
	first := seedCourseRaw(t, db, "CS1", "程序设计", 3, "计算机学院", teacher.ID, "中文", nil, nil)
	second := seedCourseRaw(t, db, "CS2", "程序设计", 3, "计算机学院", teacher.ID, "中文", nil, nil)
	semester := "2026-2027-1"
	offering := seedOfferedCourseRaw(t, db, first.ID, semester, "中文", nil, nil)
	otherOffering := seedOfferedCourseRaw(t, db, second.ID, semester, "中文", nil, nil)
	oldOffering := seedOfferedCourseRaw(t, db, second.ID, "2025-2026-2", "中文", nil, nil)
	for _, section := range []repository.CourseSectionEntity{
		{OfferedCourseID: offering.ID, LessonNo: "CS1.01", SourceLessonID: "123", CourseName: "程序设计"},
		{OfferedCourseID: oldOffering.ID, LessonNo: "CS1.01", SourceLessonID: "456", CourseName: "程序设计"},
		{OfferedCourseID: offering.ID, LessonNo: "AMBIG.01", SourceLessonID: "700", CourseName: "同名"},
		{OfferedCourseID: otherOffering.ID, LessonNo: "AMBIG.01", SourceLessonID: "701", CourseName: "同名"},
	} {
		if err := db.Create(&section).Error; err != nil {
			t.Fatal(err)
		}
	}
	user := seedUser(t, db)
	otherUser := seedUserRaw(t, db, "other-import-user", "other-import@example.com")
	repo := repository.NewCourseEnrollmentRepository(db)
	ctx := context.Background()
	result, err := repo.ImportSemester(ctx, user.ID, semester, []course.EnrollmentSourceCourse{
		{LessonNo: "CS1.01", SourceLessonID: "unrelated-id", Name: "程序设计"},
		{LessonNo: "CS1.01", Name: "程序设计"},
		{LessonNo: "AMBIG.01", Name: "同名"},
		{LessonNo: "MISSING.01", Name: "尚未收录"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Added != 1 || result.Matched != 1 || len(result.Unmatched) != 2 {
		t.Fatalf("wrong result %+v", result)
	}
	result, err = repo.ImportSemester(ctx, user.ID, semester, []course.EnrollmentSourceCourse{{SourceLessonID: "123", Name: "程序设计"}})
	if err != nil || result.Added != 0 || result.Matched != 1 {
		t.Fatalf("idempotence %+v %v", result, err)
	}
	rows, err := repo.FindUserEnrollments(ctx, user.ID)
	if err != nil || len(rows) != 1 || rows[0].Course.ID != first.ID {
		t.Fatalf("wrong semester matched %+v %v", rows, err)
	}
	others, err := repo.FindUserEnrollments(ctx, otherUser.ID)
	if err != nil || len(others) != 0 {
		t.Fatalf("another user's records changed %+v %v", others, err)
	}
	empty, err := repo.ImportSemester(ctx, user.ID, semester, []course.EnrollmentSourceCourse{})
	if err != nil || empty.Added != 0 {
		t.Fatal("empty timetable rejected")
	}
	rows, _ = repo.FindUserEnrollments(ctx, user.ID)
	if len(rows) != 1 {
		t.Fatal("empty import erased previous records")
	}
}
