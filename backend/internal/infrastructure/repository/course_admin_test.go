package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"jcourse/internal/domain/course"
	"jcourse/internal/infrastructure/repository"
)

func TestCourseDataRepositorySemesterReplaceAndMaintenance(t *testing.T) {
	dbName := fmt.Sprintf("jcourse_admin_%d", time.Now().UnixNano())
	db := newTestDBNamed(t, dbName)
	ctx := context.Background()

	const code = "CS101"
	const department = "计算机学院"
	const teacherName = "张三"
	const oldSemester = "2025-2026-2"
	const currentSemester = "2026-2027-1"
	teacher := repository.TeacherEntity{
		Code:         course.InternalTeacherCode(department, teacherName),
		Name:         teacherName,
		Department:   department,
		LastSemester: currentSemester,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(&teacher).Error; err != nil {
		t.Fatalf("seed teacher: %v", err)
	}
	master := seedCourseRaw(t, db, code, "程序设计", 3, department, teacher.ID, "中文", nil, nil)
	previous := seedOfferedCourseRaw(t, db, master.ID, oldSemester, "中文", nil, nil)
	if err := db.Create(&repository.CourseSectionEntity{
		OfferedCourseID: previous.ID,
		SourceLessonID:  "previous-term-lesson",
		CourseName:      "程序设计（旧学期）",
		Department:      department,
		Credit:          3,
		TeacherIDs:      []int64{int64(teacher.ID)},
	}).Error; err != nil {
		t.Fatalf("seed previous term section: %v", err)
	}
	current := seedOfferedCourseRaw(t, db, master.ID, currentSemester, "中文", nil, nil)
	if err := db.Create(&repository.CourseSectionEntity{
		OfferedCourseID: current.ID,
		SourceLessonID:  "old-current-lesson",
		CourseName:      "旧快照",
		Department:      department,
		Credit:          3,
		TeacherIDs:      []int64{int64(teacher.ID)},
	}).Error; err != nil {
		t.Fatalf("seed current term section: %v", err)
	}
	user := seedUser(t, db)
	seedReview(t, db, master.ID, user.ID)

	repo := repository.NewCourseRepository(db)
	replacement := course.ManagedOffering{
		Code:     code,
		Semester: currentSemester,
		Sections: []course.CourseSection{{
			SourceLessonID: "new-current-lesson",
			LessonNo:       "CS101.01",
			CourseName:     "程序设计（新学期）",
			Credit:         3,
			Department:     department,
			Language:       "中文",
			Teachers:       []string{teacherName},
		}},
	}
	if err := repo.ReplaceSemester(ctx, currentSemester, []course.ManagedOffering{replacement}); err != nil {
		t.Fatalf("replace current semester: %v", err)
	}

	currentItems, total, err := repo.ListManagedOfferings(ctx, course.ManagedOfferingFilter{Semester: currentSemester, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list replaced semester: %v", err)
	}
	if total != 1 || len(currentItems) != 1 || len(currentItems[0].Sections) != 1 || currentItems[0].Sections[0].SourceLessonID != "new-current-lesson" {
		t.Fatalf("current semester was not replaced: total=%d offerings=%#v", total, currentItems)
	}
	oldItems, oldTotal, err := repo.ListManagedOfferings(ctx, course.ManagedOfferingFilter{Semester: oldSemester, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list previous semester: %v", err)
	}
	if oldTotal != 1 || len(oldItems) != 1 || oldItems[0].Sections[0].SourceLessonID != "previous-term-lesson" {
		t.Fatalf("replacing one semester changed another: %#v", oldItems)
	}

	currentItems[0].Sections[0].CourseName = "人工维护后的课程名"
	updated, err := repo.SaveManagedOffering(ctx, currentItems[0].ID, currentItems[0])
	if err != nil {
		t.Fatalf("update managed offering: %v", err)
	}
	if updated.ID != currentItems[0].ID || updated.Sections[0].CourseName != "人工维护后的课程名" {
		t.Fatalf("managed offering update did not persist: %#v", updated)
	}
	if err := repo.DeleteManagedOffering(ctx, updated.ID); err != nil {
		t.Fatalf("delete managed offering: %v", err)
	}
	remaining, remainingCount, err := repo.ListManagedOfferings(ctx, course.ManagedOfferingFilter{Semester: currentSemester, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list deleted semester: %v", err)
	}
	if remainingCount != 0 || len(remaining) != 0 {
		t.Fatalf("deleted offering is still present: %#v", remaining)
	}
	var reviewCount int64
	if err := db.Model(&repository.ReviewEntity{}).Where("course_id = ?", master.ID).Count(&reviewCount).Error; err != nil {
		t.Fatalf("count retained reviews: %v", err)
	}
	if reviewCount != 1 {
		t.Fatalf("course reviews were deleted with an offering: got %d, want 1", reviewCount)
	}
	oldItems, oldTotal, err = repo.ListManagedOfferings(ctx, course.ManagedOfferingFilter{Semester: oldSemester, Page: 1, PageSize: 20})
	if err != nil || oldTotal != 1 || len(oldItems) != 1 {
		t.Fatalf("deleting current semester changed older offering: total=%d offerings=%#v err=%v", oldTotal, oldItems, err)
	}
}
