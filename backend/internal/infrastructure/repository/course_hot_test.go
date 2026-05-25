package repository

import (
	"testing"

	"jcourse/internal/domain/course"
)

func TestCourseHotRepositoryKeyFromPeriod(t *testing.T) {
	repo := NewCourseHotRepository(nil)
	period := course.HotCoursePeriod{Period: course.HotCoursePeriodWeek, PeriodKey: "2026-21"}

	if got, want := repo.key(period), "jcourse:course:hot:week:2026-21"; got != want {
		t.Fatalf("key = %s, want %s", got, want)
	}
}
