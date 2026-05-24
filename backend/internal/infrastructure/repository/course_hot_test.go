package repository

import (
	"testing"
	"time"

	"jcourse/internal/domain/course"
)

func mustShanghaiLocation(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	return loc
}

func TestWeekKeyPart(t *testing.T) {
	loc := mustShanghaiLocation(t)
	tests := []struct {
		name string
		at   time.Time
		want string
	}{
		{
			name: "monday starts week one",
			at:   time.Date(2024, time.January, 1, 12, 0, 0, 0, loc),
			want: "2024-01",
		},
		{
			name: "midweek uses same monday",
			at:   time.Date(2024, time.January, 3, 12, 0, 0, 0, loc),
			want: "2024-01",
		},
		{
			name: "cross year uses week start year",
			at:   time.Date(2025, time.January, 1, 12, 0, 0, 0, loc),
			want: "2024-53",
		},
		{
			name: "late december week belongs to monday year",
			at:   time.Date(2026, time.January, 1, 12, 0, 0, 0, loc),
			want: "2025-52",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HotCoursePeriodKey(course.HotCoursePeriodWeek, tt.at, loc); got != tt.want {
				t.Fatalf("HotCoursePeriodKey(week) = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestMonthKeyPart(t *testing.T) {
	loc := mustShanghaiLocation(t)
	at := time.Date(2026, time.May, 22, 12, 0, 0, 0, loc)
	if got := HotCoursePeriodKey(course.HotCoursePeriodMonth, at, loc); got != "2026-05" {
		t.Fatalf("HotCoursePeriodKey(month) = %s, want 2026-05", got)
	}
}

func TestHotCoursePeriodRange(t *testing.T) {
	loc := mustShanghaiLocation(t)
	at := time.Date(2026, time.May, 24, 12, 0, 0, 0, loc)

	start, end := HotCoursePeriodRange(course.HotCoursePeriodWeek, at, loc)
	assertTimeEqual(t, start, time.Date(2026, time.May, 18, 0, 0, 0, 0, loc))
	assertTimeEqual(t, end, time.Date(2026, time.May, 25, 0, 0, 0, 0, loc))

	start, end = HotCoursePeriodRange(course.HotCoursePeriodMonth, at, loc)
	assertTimeEqual(t, start, time.Date(2026, time.May, 1, 0, 0, 0, 0, loc))
	assertTimeEqual(t, end, time.Date(2026, time.June, 1, 0, 0, 0, 0, loc))
}

func assertTimeEqual(t *testing.T, got time.Time, want time.Time) {
	t.Helper()
	if !got.Equal(want) {
		t.Fatalf("time = %s, want %s", got, want)
	}
}
