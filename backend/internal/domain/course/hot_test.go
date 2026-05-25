package course

import (
	"testing"
	"time"
)

func mustShanghaiLocation(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	return loc
}

func TestHotCoursePeriodKeyWeek(t *testing.T) {
	loc := mustShanghaiLocation(t)
	tests := []struct {
		name string
		at   time.Time
		want string
	}{
		{
			name: "monday starts iso week one",
			at:   time.Date(2024, time.January, 1, 12, 0, 0, 0, loc),
			want: "2024-01",
		},
		{
			name: "midweek uses same iso week",
			at:   time.Date(2024, time.January, 3, 12, 0, 0, 0, loc),
			want: "2024-01",
		},
		{
			name: "early january can belong to current iso year",
			at:   time.Date(2025, time.January, 1, 12, 0, 0, 0, loc),
			want: "2025-01",
		},
		{
			name: "late december can belong to next iso year",
			at:   time.Date(2024, time.December, 31, 12, 0, 0, 0, loc),
			want: "2025-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HotCoursePeriodKey(HotCoursePeriodWeek, tt.at, loc); got != tt.want {
				t.Fatalf("HotCoursePeriodKey(week) = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestHotCoursePeriodKeyMonth(t *testing.T) {
	loc := mustShanghaiLocation(t)
	at := time.Date(2026, time.May, 22, 12, 0, 0, 0, loc)
	if got := HotCoursePeriodKey(HotCoursePeriodMonth, at, loc); got != "2026-05" {
		t.Fatalf("HotCoursePeriodKey(month) = %s, want 2026-05", got)
	}
}

func TestHotCoursePeriodRange(t *testing.T) {
	loc := mustShanghaiLocation(t)
	at := time.Date(2026, time.May, 24, 12, 0, 0, 0, loc)

	start, end := HotCoursePeriodRange(HotCoursePeriodWeek, at, loc)
	assertTimeEqual(t, start, time.Date(2026, time.May, 18, 0, 0, 0, 0, loc))
	assertTimeEqual(t, end, time.Date(2026, time.May, 25, 0, 0, 0, 0, loc))

	start, end = HotCoursePeriodRange(HotCoursePeriodMonth, at, loc)
	assertTimeEqual(t, start, time.Date(2026, time.May, 1, 0, 0, 0, 0, loc))
	assertTimeEqual(t, end, time.Date(2026, time.June, 1, 0, 0, 0, 0, loc))
}

func assertTimeEqual(t *testing.T, got time.Time, want time.Time) {
	t.Helper()
	if !got.Equal(want) {
		t.Fatalf("time = %s, want %s", got, want)
	}
}
