package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"jcourse/internal/domain/course"
	"jcourse/internal/infrastructure/repository"

	"gorm.io/gorm"
)

func TestSemesterImportBatchesPreserveIdentityAndRollback(t *testing.T) {
	db := newTestDBNamed(t, fmt.Sprintf("jcourse_batches_%d", time.Now().UnixNano()))
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()
	const semester = "2025-2026-2"
	if err := db.Callback().Create().After("gorm:create").Register("test:bounded_inserts", func(tx *gorm.DB) {
		switch tx.Statement.Table {
		case "teachers", "courses", "course_sections":
			if tx.RowsAffected > 50 {
				t.Errorf("unbounded %s insert: %d rows", tx.Statement.Table, tx.RowsAffected)
			}
		}
	}); err != nil {
		t.Fatal(err)
	}
	offerings := make([]course.ManagedOffering, 125)
	for i := range offerings {
		offerings[i] = course.ManagedOffering{Code: fmt.Sprintf("BATCH%03d", i), Semester: semester,
			Sections: []course.CourseSection{{SourceLessonID: fmt.Sprintf("lesson-%d", i),
				CourseName: "批量课程", Department: "测试学院", Credit: 2,
				Teachers: []string{fmt.Sprintf("测试教师%d", i)}}}}
	}
	// Exercise section batching as well as teacher and course batching.
	for i := 1; i < 125; i++ {
		section := offerings[0].Sections[0]
		section.SourceLessonID = fmt.Sprintf("extra-%d", i)
		offerings[0].Sections = append(offerings[0].Sections, section)
	}
	for attempt := 0; attempt < 2; attempt++ {
		if err := repo.ReplaceSemester(ctx, semester, offerings); err != nil {
			t.Fatal(err)
		}
	}
	for _, check := range []struct {
		model any
		want  int64
	}{
		{&repository.TeacherEntity{}, 125}, {&repository.CourseEntity{}, 125},
		{&repository.OfferedCourseEntity{}, 125},
	} {
		var count int64
		if err := db.Model(check.model).Count(&count).Error; err != nil || count != check.want {
			t.Fatalf("%T: count=%d want=%d err=%v", check.model, count, check.want, err)
		}
	}
	// Fail after earlier batches have already written, including the semester delete.
	if err := db.Callback().Create().After("gorm:create").Register("test:fail_sections", func(tx *gorm.DB) {
		if tx.Statement.Table == "course_sections" {
			tx.AddError(fmt.Errorf("injected section failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceSemester(ctx, semester, offerings[:1]); err == nil {
		t.Fatal("expected import failure")
	}
	var count int64
	if err := db.Model(&repository.CourseSectionEntity{}).
		Joins("JOIN offered_courses oc ON oc.id = course_sections.offered_course_id").
		Where("oc.semester = ?", semester).Count(&count).Error; err != nil || count != 249 {
		t.Fatalf("failed import did not restore the previous semester: count=%d err=%v", count, err)
	}
}

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

func TestManagedOfferingColumnFiltersAndPagination(t *testing.T) {
	db := newTestDBNamed(t, fmt.Sprintf("jcourse_filters_%d", time.Now().UnixNano()))
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()
	semester := "2026-2027-1"
	section := func(name, no, department, teacher string) course.CourseSection {
		return course.CourseSection{CourseName: name, LessonNo: no, Department: department, Teachers: []string{teacher}, Credit: 2}
	}
	err := repo.ReplaceSemester(ctx, semester, []course.ManagedOffering{
		{Code: "A_01", Semester: semester, Sections: []course.CourseSection{
			section("Alpha 100%", "A.01", "First", "Alice"), section("Alpha 100%", "A.02", "Second", "Bob"),
		}},
		{Code: "AB01", Semester: semester, Sections: []course.CourseSection{section("Beta", "B.01", "First", "Alice")}},
	})
	if err != nil {
		t.Fatal(err)
	}
	one, two, zero := 1, 2, 0
	for _, tc := range []struct {
		name   string
		filter course.ManagedOfferingFilter
		want   int64
	}{
		{"literal underscore", course.ManagedOfferingFilter{Code: "_"}, 1},
		{"literal percent", course.ManagedOfferingFilter{Name: "%"}, 1},
		{"case insensitive", course.ManagedOfferingFilter{Code: "ab"}, 1},
		{"second section primary teacher", course.ManagedOfferingFilter{Teacher: "Bob"}, 1},
		{"combined columns", course.ManagedOfferingFilter{Code: "A_", Name: "Alpha", Teacher: "Bob", Department: "Second", LessonNo: "A.02"}, 1},
		{"same section required", course.ManagedOfferingFilter{Teacher: "Bob", Department: "First"}, 0},
		{"exact count", course.ManagedOfferingFilter{MinSections: &two, MaxSections: &two}, 1},
		{"upper bound", course.ManagedOfferingFilter{MaxSections: &one}, 1},
		{"zero upper bound", course.ManagedOfferingFilter{MaxSections: &zero}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.filter.Semester = semester
			items, total, err := repo.ListManagedOfferings(ctx, tc.filter)
			if err != nil || total != tc.want || int64(len(items)) != tc.want {
				t.Fatalf("total=%d items=%d err=%v", total, len(items), err)
			}
		})
	}
	for _, sort := range []string{"code", "name", "teacher", "department", "sections"} {
		items, total, err := repo.ListManagedOfferings(ctx, course.ManagedOfferingFilter{Semester: semester, Sort: sort, Desc: true, Page: 2, PageSize: 1})
		if err != nil || total != 2 || len(items) != 1 {
			t.Fatalf("sort %s pagination: total=%d items=%d err=%v", sort, total, len(items), err)
		}
		if sort == "sections" && len(items[0].Sections) != 1 {
			t.Fatal("section count sorting was ignored")
		}
	}
}
