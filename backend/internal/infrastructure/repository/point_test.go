package repository_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/point"
	"jcourse/internal/infrastructure/repository"

	"gorm.io/gorm"
)

func TestPointRepository_SumByUser(t *testing.T) {
	db := newTestDB(t)
	cleanTables(t, db, "user_point_records", "users")
	repo := repository.NewPointRepository(db)
	ctx := context.Background()
	user := seedUser(t, db)

	total, err := repo.SumByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("SumByUser empty: %v", err)
	}
	if total != 0 {
		t.Fatalf("empty total = %d, want 0", total)
	}

	seedPointRecord(t, db, user.ID, "review_created", 10, "发布课程评价", time.Now())
	seedPointRecord(t, db, user.ID, "review_deleted", -3, "删除课程评价", time.Now())
	other := seedUserRaw(t, db, "other", "other@example.com")
	seedPointRecord(t, db, other.ID, "review_created", 100, "发布课程评价", time.Now())

	total, err = repo.SumByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("SumByUser: %v", err)
	}
	if total != 7 {
		t.Fatalf("total = %d, want 7", total)
	}
}

func TestPointRepository_FindRecordsByUser(t *testing.T) {
	db := newTestDB(t)
	cleanTables(t, db, "user_point_records", "users")
	repo := repository.NewPointRepository(db)
	ctx := context.Background()
	user := seedUser(t, db)
	other := seedUserRaw(t, db, "other", "other@example.com")

	older := time.Now().Add(-time.Hour).Truncate(time.Second)
	newer := time.Now().Truncate(time.Second)
	seedPointRecord(t, db, user.ID, "old", 1, "较早记录", older)
	seedPointRecord(t, db, user.ID, "newer_first", 2, "同一时间较早插入", newer)
	latest := seedPointRecord(t, db, user.ID, "newer_second", 3, "同一时间较晚插入", newer)
	seedPointRecord(t, db, other.ID, "other", 100, "其他用户记录", time.Now())

	records, total, err := repo.FindRecordsByUser(ctx, point.RecordFilter{UserID: user.ID, Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("FindRecordsByUser: %v", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2", len(records))
	}
	if records[0].ID != latest.ID || records[0].Reason != "newer_second" {
		t.Fatalf("first record = %+v, want latest same-time id", records[0])
	}
	if records[1].Reason != "newer_first" {
		t.Fatalf("second reason = %q, want newer_first", records[1].Reason)
	}

	records, total, err = repo.FindRecordsByUser(ctx, point.RecordFilter{UserID: user.ID, Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("FindRecordsByUser page 2: %v", err)
	}
	if total != 3 || len(records) != 1 || records[0].Reason != "old" {
		t.Fatalf("page 2 records=%+v total=%d, want old only with total 3", records, total)
	}
}

func seedUserRaw(t *testing.T, db *gorm.DB, username, email string) repository.UserEntity {
	t.Helper()
	e := repository.UserEntity{
		Username:   username,
		Email:      email,
		Role:       "user",
		Password:   "hashed_password",
		CreatedAt:  time.Now(),
		LastSeenAt: time.Now(),
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed user raw: %v", err)
	}
	return e
}

func seedPointRecord(t *testing.T, db *gorm.DB, userID int, reason string, amount int, description string, createdAt time.Time) repository.UserPointRecordEntity {
	t.Helper()
	e := repository.UserPointRecordEntity{
		UserID:      userID,
		Reason:      reason,
		Amount:      amount,
		Description: description,
		CreatedAt:   createdAt,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed point record: %v", err)
	}
	return e
}
