package repository_test

import (
	"context"
	"testing"

	"github.com/lib/pq"
	"jcourse/internal/domain/course"
	"jcourse/internal/infrastructure/repository"
)

func TestCourseFiltersUseSameCHDOfferingAndAllTeachers(t *testing.T) {
	db := newTestDB(t)
	main := seedTeacher(t, db)
	second := seedTeacherRaw(t, db, "CHD-T2", "第二教师", "计算机学院", "讲师")
	c := seedCourse(t, db, main.ID)
	for _, item := range []struct{ semester, campus string }{{"2025-2026-2", "雁塔校区"}, {"2026-2027-1", "渭水校区"}} {
		o := repository.OfferedCourseEntity{CourseID: c.ID, Semester: item.semester, TeacherIDs: pq.Int64Array{int64(main.ID), int64(second.ID)}}
		if err := db.Create(&o).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&repository.CourseSectionEntity{OfferedCourseID: o.ID, Campus: item.campus}).Error; err != nil {
			t.Fatal(err)
		}
	}
	repo := repository.NewCourseRepository(db)
	for _, tc := range []struct {
		filter course.CourseFilter
		count  int64
	}{
		{course.CourseFilter{Semester: "2026-2027-1", Campus: "渭水校区"}, 1},
		{course.CourseFilter{Semester: "2026-2027-1", Campus: "雁塔校区"}, 0},
		{course.CourseFilter{TeacherID: second.ID}, 1},
	} {
		_, total, err := repo.FindBy(context.Background(), tc.filter)
		if err != nil {
			t.Fatal(err)
		}
		if total != tc.count {
			t.Fatalf("filter %+v: count %d, want %d", tc.filter, total, tc.count)
		}
	}
}
