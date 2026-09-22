package repository

import (
	"testing"

	"jcourse/internal/domain/course"
)

func TestRedisKey(t *testing.T) {
	got := redisKey("course", 42, "detail")
	want := "jcourse:course:42:detail"
	if got != want {
		t.Fatalf("redisKey = %s, want %s", got, want)
	}
}

func TestCourseHotRepositoryKey(t *testing.T) {
	repo := NewCourseHotRepository(nil)

	week := course.HotCoursePeriod{Period: course.HotCoursePeriodWeek, PeriodKey: "2026-21"}
	month := course.HotCoursePeriod{Period: course.HotCoursePeriodMonth, PeriodKey: "2026-05"}

	if got, want := repo.key(week), "jcourse:course:hot:week:2026-21"; got != want {
		t.Fatalf("week key = %s, want %s", got, want)
	}
	if got, want := repo.key(month), "jcourse:course:hot:month:2026-05"; got != want {
		t.Fatalf("month key = %s, want %s", got, want)
	}
}
