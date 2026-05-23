package repository_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/course"
	"jcourse/internal/infrastructure/repository"
)

func TestGormCourseHotRepository_AddScoreAndTop(t *testing.T) {
	db := newTestDB(t)
	cleanTables(t, db, "course_hot_scores")

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	repo := repository.NewGormCourseHotRepository(db)
	ctx := context.Background()
	at := time.Date(2026, time.May, 22, 12, 0, 0, 0, loc)

	if err := repo.AddScore(ctx, 1, 3, at); err != nil {
		t.Fatalf("add score: %v", err)
	}
	if err := repo.AddScore(ctx, 2, 5, at); err != nil {
		t.Fatalf("add score: %v", err)
	}
	if err := repo.AddScore(ctx, 1, 4, at); err != nil {
		t.Fatalf("add score again: %v", err)
	}
	if err := repo.AddScore(ctx, 3, 100, at.AddDate(0, 1, 0)); err != nil {
		t.Fatalf("add next month score: %v", err)
	}

	ranks, err := repo.Top(ctx, course.HotCoursePeriodWeek, at, 2)
	if err != nil {
		t.Fatalf("top week: %v", err)
	}
	assertHotCourseRanks(t, ranks, []course.HotCourseRank{
		{CourseID: 1, Score: 7},
		{CourseID: 2, Score: 5},
	})

	ranks, err = repo.Top(ctx, course.HotCoursePeriodMonth, at, 10)
	if err != nil {
		t.Fatalf("top month: %v", err)
	}
	assertHotCourseRanks(t, ranks, []course.HotCourseRank{
		{CourseID: 1, Score: 7},
		{CourseID: 2, Score: 5},
	})
}

func TestGormCourseHotRepository_TopEdgeCases(t *testing.T) {
	db := newTestDB(t)
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	repo := repository.NewGormCourseHotRepository(db)
	ctx := context.Background()
	at := time.Date(2026, time.May, 22, 12, 0, 0, 0, loc)

	ranks, err := repo.Top(ctx, course.HotCoursePeriod("daily"), at, 0)
	if err != nil {
		t.Fatalf("top with zero limit should match redis behavior: %v", err)
	}
	if len(ranks) != 0 {
		t.Fatalf("top with zero limit len = %d, want 0", len(ranks))
	}

	_, err = repo.Top(ctx, course.HotCoursePeriod("daily"), at, 10)
	if err != course.ErrInvalidHotCoursePeriod {
		t.Fatalf("top invalid period err = %v, want %v", err, course.ErrInvalidHotCoursePeriod)
	}
}

func assertHotCourseRanks(t *testing.T, got []course.HotCourseRank, want []course.HotCourseRank) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("rank len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("rank[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}
