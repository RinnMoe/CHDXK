package repository_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/announcement"
	"jcourse/internal/infrastructure/repository"
)

func TestAnnouncementRepository_FindActive(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewAnnouncementRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "announcements")

	now := time.Now().UTC().Truncate(time.Microsecond)

	// Active announcement with higher priority
	a1 := repository.AnnouncementEntity{
		Title:     "紧急通知",
		Body:      "系统维护",
		Priority:  10,
		ShowStart: now.Add(-1 * time.Hour),
		ShowEnd:   now.Add(1 * time.Hour),
		CreatedAt: now.Add(-2 * time.Hour),
	}
	// Active announcement with lower priority
	a2 := repository.AnnouncementEntity{
		Title:     "一般通知",
		Body:      "功能更新",
		Priority:  5,
		ShowStart: now.Add(-2 * time.Hour),
		ShowEnd:   now.Add(2 * time.Hour),
		CreatedAt: now.Add(-1 * time.Hour),
	}
	// Expired announcement
	a3 := repository.AnnouncementEntity{
		Title:     "过期通知",
		Body:      "已结束",
		Priority:  20,
		ShowStart: now.Add(-3 * time.Hour),
		ShowEnd:   now.Add(-1 * time.Hour),
		CreatedAt: now.Add(-4 * time.Hour),
	}
	// Future announcement
	a4 := repository.AnnouncementEntity{
		Title:     "未来通知",
		Body:      "尚未开始",
		Priority:  15,
		ShowStart: now.Add(1 * time.Hour),
		ShowEnd:   now.Add(3 * time.Hour),
		CreatedAt: now,
	}

	for _, e := range []repository.AnnouncementEntity{a1, a2, a3, a4} {
		if err := db.Create(&e).Error; err != nil {
			t.Fatalf("seed announcement: %v", err)
		}
	}

	t.Run("returns only active announcements sorted by priority desc", func(t *testing.T) {
		results, err := repo.FindActive(ctx)
		if err != nil {
			t.Fatalf("FindActive: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("results count: got %d, want 2", len(results))
		}
		if results[0].Title != "紧急通知" {
			t.Errorf("first result: got %q, want 紧急通知", results[0].Title)
		}
		if results[0].Priority != 10 {
			t.Errorf("first priority: got %d, want 10", results[0].Priority)
		}
		if results[1].Title != "一般通知" {
			t.Errorf("second result: got %q, want 一般通知", results[1].Title)
		}
		if results[1].Priority != 5 {
			t.Errorf("second priority: got %d, want 5", results[1].Priority)
		}
	})

	t.Run("same priority sorted by created_at desc", func(t *testing.T) {
		cleanTables(t, db, "announcements")

		for _, e := range []repository.AnnouncementEntity{
			{Title: "较早", Body: "b", Priority: 5, ShowStart: now.Add(-1 * time.Hour), ShowEnd: now.Add(1 * time.Hour), CreatedAt: now.Add(-2 * time.Hour)},
			{Title: "较晚", Body: "b", Priority: 5, ShowStart: now.Add(-1 * time.Hour), ShowEnd: now.Add(1 * time.Hour), CreatedAt: now},
		} {
			if err := db.Create(&e).Error; err != nil {
				t.Fatalf("seed: %v", err)
			}
		}

		results, err := repo.FindActive(ctx)
		if err != nil {
			t.Fatalf("FindActive: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("results count: got %d, want 2", len(results))
		}
		if results[0].Title != "较晚" {
			t.Errorf("first: got %q, want 较晚", results[0].Title)
		}
		if results[1].Title != "较早" {
			t.Errorf("second: got %q, want 较早", results[1].Title)
		}
	})

	t.Run("empty when no active announcements", func(t *testing.T) {
		cleanTables(t, db, "announcements")

		for _, e := range []repository.AnnouncementEntity{
			{Title: "过期", Body: "b", Priority: 1, ShowStart: now.Add(-2 * time.Hour), ShowEnd: now.Add(-1 * time.Hour), CreatedAt: now},
			{Title: "未来", Body: "b", Priority: 1, ShowStart: now.Add(1 * time.Hour), ShowEnd: now.Add(2 * time.Hour), CreatedAt: now},
		} {
			if err := db.Create(&e).Error; err != nil {
				t.Fatalf("seed: %v", err)
			}
		}

		results, err := repo.FindActive(ctx)
		if err != nil {
			t.Fatalf("FindActive: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("results: got %d, want 0", len(results))
		}
	})

	t.Run("maps all fields correctly", func(t *testing.T) {
		cleanTables(t, db, "announcements")

		e := repository.AnnouncementEntity{
			Title:     "标题",
			Body:      "内容",
			Priority:  7,
			ShowStart: now.Add(-1 * time.Hour),
			ShowEnd:   now.Add(1 * time.Hour),
			LinkURL:   "https://example.com/detail",
			LinkTitle: "查看详情",
			CreatedAt: now,
		}
		if err := db.Create(&e).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}

		results, err := repo.FindActive(ctx)
		if err != nil {
			t.Fatalf("FindActive: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("results: got %d, want 1", len(results))
		}
		r := results[0]
		if r.Title != "标题" || r.Body != "内容" || r.Priority != 7 || r.LinkURL != "https://example.com/detail" || r.LinkTitle != "查看详情" {
			t.Errorf("fields mismatch: got %+v", r)
		}
	})
}

// Compile-time interface check
var _ announcement.AnnouncementQuery = (*repository.AnnouncementRepository)(nil)
