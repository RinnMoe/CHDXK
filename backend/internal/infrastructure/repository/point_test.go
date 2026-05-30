package repository_test

import (
	"context"
	"database/sql"
	"errors"
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

	seedPointRecord(t, db, user.ID, point.RecordReason("review_created"), 10, "发布课程评价", time.Now())
	seedPointRecord(t, db, user.ID, point.RecordReason("review_deleted"), -3, "删除课程评价", time.Now())
	other := seedUserRaw(t, db, "other", "other@example.com")
	seedPointRecord(t, db, other.ID, point.RecordReason("review_created"), 100, "发布课程评价", time.Now())

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
	seedPointRecord(t, db, user.ID, point.RecordReason("old"), 1, "较早记录", older)
	seedPointRecord(t, db, user.ID, point.RecordReason("newer_first"), 2, "同一时间较早插入", newer)
	latest := seedPointRecord(t, db, user.ID, point.RecordReason("newer_second"), 3, "同一时间较晚插入", newer)
	seedPointRecord(t, db, other.ID, point.RecordReason("other"), 100, "其他用户记录", time.Now())

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

func TestPointRepository_CreateTransfer(t *testing.T) {
	db := newTestDB(t)
	cleanTables(t, db, "point_transfers", "user_point_records", "users")
	repo := repository.NewPointRepository(db)
	ctx := context.Background()
	sender := seedUser(t, db)
	recipient := seedUserRaw(t, db, "recipient", "recipient@example.com")
	now := time.Now().Truncate(time.Second)
	seedPointRecord(t, db, sender.ID, point.RecordReason("seed"), 200, "初始积分", now.Add(-time.Minute))

	transfer := point.Transfer{
		SenderUserID:    sender.ID,
		RecipientUserID: recipient.ID,
		Amount:          100,
		Fee:             2,
		FeePayer:        point.FeePayerSender,
		SenderDelta:     -102,
		RecipientDelta:  100,
		CreatedAt:       now,
	}
	senderRecord := point.Record{UserID: sender.ID, Reason: point.RecordReasonTransferOut, Amount: -102, Description: "积分转出", CreatedAt: now}
	recipientRecord := point.Record{UserID: recipient.ID, Reason: point.RecordReasonTransferIn, Amount: 100, Description: "积分转入", CreatedAt: now}

	if err := repo.CreateTransfer(ctx, &transfer, senderRecord, recipientRecord); err != nil {
		t.Fatalf("CreateTransfer: %v", err)
	}
	if transfer.ID == 0 {
		t.Fatal("transfer ID was not set")
	}

	var transferCount int64
	if err := db.Model(&repository.PointTransferEntity{}).Count(&transferCount).Error; err != nil {
		t.Fatalf("count transfers: %v", err)
	}
	if transferCount != 1 {
		t.Fatalf("transfer count = %d, want 1", transferCount)
	}

	senderTotal, err := repo.SumByUser(ctx, sender.ID)
	if err != nil {
		t.Fatalf("sender SumByUser: %v", err)
	}
	recipientTotal, err := repo.SumByUser(ctx, recipient.ID)
	if err != nil {
		t.Fatalf("recipient SumByUser: %v", err)
	}
	if senderTotal != 98 || recipientTotal != 100 {
		t.Fatalf("totals sender=%d recipient=%d, want 98 and 100", senderTotal, recipientTotal)
	}
}

func TestPointRepository_CreateTransferRejectsInsufficientBalance(t *testing.T) {
	db := newTestDB(t)
	cleanTables(t, db, "point_transfers", "user_point_records", "users")
	repo := repository.NewPointRepository(db)
	ctx := context.Background()
	sender := seedUser(t, db)
	recipient := seedUserRaw(t, db, "recipient", "recipient@example.com")
	seedPointRecord(t, db, sender.ID, point.RecordReason("seed"), 50, "初始积分", time.Now())

	transfer := point.Transfer{
		SenderUserID:    sender.ID,
		RecipientUserID: recipient.ID,
		Amount:          100,
		Fee:             2,
		FeePayer:        point.FeePayerSender,
		SenderDelta:     -102,
		RecipientDelta:  100,
		CreatedAt:       time.Now(),
	}
	err := repo.CreateTransfer(ctx, &transfer,
		point.Record{UserID: sender.ID, Reason: point.RecordReasonTransferOut, Amount: -102, Description: "积分转出", CreatedAt: time.Now()},
		point.Record{UserID: recipient.ID, Reason: point.RecordReasonTransferIn, Amount: 100, Description: "积分转入", CreatedAt: time.Now()},
	)
	if !errors.Is(err, point.ErrInsufficientBalance) {
		t.Fatalf("CreateTransfer error = %v, want ErrInsufficientBalance", err)
	}

	var transferCount int64
	if err := db.Model(&repository.PointTransferEntity{}).Count(&transferCount).Error; err != nil {
		t.Fatalf("count transfers: %v", err)
	}
	if transferCount != 0 {
		t.Fatalf("transfer count = %d, want 0", transferCount)
	}
}

func TestPointRepository_GrantRewardIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	cleanTables(t, db, "point_rewards", "user_point_records", "users")
	repo := repository.NewPointRepository(db)
	ctx := context.Background()
	user := seedUser(t, db)
	now := time.Now().Truncate(time.Second)
	reward := repository.PointRewardEntity{
		UserID:      user.ID,
		Reason:      string(point.RewardReasonCourseFirstReview),
		Amount:      10,
		SourceType:  point.RewardSourceTypeCourse,
		SourceKey:   "1",
		Description: "课程首评奖励",
		Status:      string(point.RewardStatusPending),
		CreatedAt:   now,
	}
	if err := db.Create(&reward).Error; err != nil {
		t.Fatalf("seed reward: %v", err)
	}

	if err := repo.GrantReward(ctx, reward.ID, now.Add(time.Minute)); err != nil {
		t.Fatalf("GrantReward first: %v", err)
	}
	if err := repo.GrantReward(ctx, reward.ID, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("GrantReward second: %v", err)
	}

	var recordCount int64
	if err := db.Model(&repository.UserPointRecordEntity{}).Where("user_id = ?", user.ID).Count(&recordCount).Error; err != nil {
		t.Fatalf("count records: %v", err)
	}
	if recordCount != 1 {
		t.Fatalf("point record count = %d, want 1", recordCount)
	}

	total, err := repo.SumByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("SumByUser: %v", err)
	}
	if total != 10 {
		t.Fatalf("total = %d, want 10", total)
	}

	var got repository.PointRewardEntity
	if err := db.Take(&got, reward.ID).Error; err != nil {
		t.Fatalf("load reward: %v", err)
	}
	if got.Status != string(point.RewardStatusGranted) || got.GrantedAt == nil {
		t.Fatalf("reward status=%q granted_at=%v, want granted", got.Status, got.GrantedAt)
	}
}

func seedUserRaw(t *testing.T, db *gorm.DB, username, email string) repository.UserEntity {
	t.Helper()
	e := repository.UserEntity{
		Username:     username,
		Email:        sql.NullString{String: email, Valid: email != ""},
		Role:         "user",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		LastSeenAt:   time.Now(),
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed user raw: %v", err)
	}
	return e
}

func seedPointRecord(t *testing.T, db *gorm.DB, userID int, reason point.RecordReason, amount int, description string, createdAt time.Time) repository.UserPointRecordEntity {
	t.Helper()
	e := repository.UserPointRecordEntity{
		UserID:      userID,
		Reason:      string(reason),
		Amount:      amount,
		Description: description,
		CreatedAt:   createdAt,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed point record: %v", err)
	}
	return e
}
