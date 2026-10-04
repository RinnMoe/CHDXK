package repository_test

import (
	"context"
	"github.com/lib/pq"
	"jcourse/internal/infrastructure/repository"
	"testing"
)

func TestTeacherHeatUsesLatestOfferingAndDeduplicatesSections(t *testing.T) {
	db := newTestDB(t)
	old := seedTeacher(t, db)
	current := seedTeacherRaw(t, db, "CHD-HOT2", "当前教师", "计算机学院", "讲师")
	co := seedTeacherRaw(t, db, "CHD-HOT3", "合授教师", "计算机学院", "讲师")
	c := seedCourse(t, db, old.ID)
	for _, item := range []struct {
		semester string
		teachers pq.Int64Array
	}{
		{"2025-2026-2", pq.Int64Array{int64(old.ID)}},
		{"2026-2027-1", pq.Int64Array{int64(current.ID), int64(co.ID), int64(current.ID)}},
	} {
		if err := db.Create(&repository.OfferedCourseEntity{CourseID: c.ID, Semester: item.semester, TeacherIDs: item.teachers}).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Exec("INSERT INTO course_hot_scores (period, period_key, course_id, score) VALUES ('week', '2026-41', ?, 7), ('month', '2026-10', ?, 20)", c.ID, c.ID).Error; err != nil {
		t.Fatal(err)
	}
	repo := repository.NewGormCourseHotRepository(db)
	rows, err := repo.TopTeachers(context.Background(), "week", "2026-41", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected current co-teachers, got %+v", rows)
	}
	for _, r := range rows {
		if r.Score != 7 || r.CourseCount != 1 || r.Teacher.ID == old.ID {
			t.Fatalf("unexpected heat %+v", r)
		}
	}
	empty, err := repo.TopTeachers(context.Background(), "week", "2026-42", 50)
	if err != nil || len(empty) != 0 {
		t.Fatalf("period isolation: %+v, %v", empty, err)
	}
}
