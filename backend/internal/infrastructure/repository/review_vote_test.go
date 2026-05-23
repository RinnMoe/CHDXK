package repository_test

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/review"
	"jcourse/internal/infrastructure/repository"
)

func seedVote(t *testing.T, db *gorm.DB, reviewID, userID, voteType int) repository.ReviewVoteEntity {
	t.Helper()
	e := repository.ReviewVoteEntity{
		ReviewID:  reviewID,
		UserID:    userID,
		VoteType:  voteType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed vote: %v", err)
	}
	return e
}

func setupVoteTest(t *testing.T) (*gorm.DB, *repository.ReviewVoteRepository, int, int) {
	t.Helper()
	db := newTestDB(t)
	repo := repository.NewReviewVoteRepository(db)

	cleanTables(t, db, "review_votes", "reviews", "courses", "teachers", "users")

	teacher := seedTeacher(t, db)
	course := seedCourse(t, db, teacher.ID)
	user := seedUser(t, db)
	rv := seedReview(t, db, course.ID, user.ID)

	return db, repo, rv.ID, user.ID
}

func TestVoteRepository_Save_NewVote(t *testing.T) {
	db, repo, reviewID, userID := setupVoteTest(t)
	ctx := context.Background()

	v := &review.Vote{
		ReviewID:  reviewID,
		UserID:    userID,
		VoteType:  review.VoteLike,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repo.Save(ctx, v); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByReviewAndUser(ctx, reviewID, userID)
	if err != nil {
		t.Fatalf("FindByReviewAndUser: %v", err)
	}
	if got == nil {
		t.Fatal("expected vote, got nil")
	}
	if got.VoteType != review.VoteLike {
		t.Errorf("VoteType: got %d, want %d", got.VoteType, review.VoteLike)
	}

	var re repository.ReviewEntity
	db.Where("id = ?", reviewID).Take(&re)
	if re.LikeCount != 1 {
		t.Errorf("LikeCount: got %d, want 1", re.LikeCount)
	}
	if re.DislikeCount != 0 {
		t.Errorf("DislikeCount: got %d, want 0", re.DislikeCount)
	}
}

func TestVoteRepository_Save_ToggleVote(t *testing.T) {
	db, repo, reviewID, userID := setupVoteTest(t)
	ctx := context.Background()

	v := &review.Vote{
		ReviewID:  reviewID,
		UserID:    userID,
		VoteType:  review.VoteLike,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Save(ctx, v)

	v.VoteType = review.VoteDislike
	v.UpdatedAt = time.Now()
	if err := repo.Save(ctx, v); err != nil {
		t.Fatalf("Save toggle: %v", err)
	}

	got, _ := repo.FindByReviewAndUser(ctx, reviewID, userID)
	if got.VoteType != review.VoteDislike {
		t.Errorf("VoteType after toggle: got %d, want %d", got.VoteType, review.VoteDislike)
	}

	var re repository.ReviewEntity
	db.Where("id = ?", reviewID).Take(&re)
	if re.LikeCount != 0 {
		t.Errorf("LikeCount after toggle: got %d, want 0", re.LikeCount)
	}
	if re.DislikeCount != 1 {
		t.Errorf("DislikeCount after toggle: got %d, want 1", re.DislikeCount)
	}
}

func TestVoteRepository_Delete(t *testing.T) {
	db, repo, reviewID, userID := setupVoteTest(t)
	ctx := context.Background()

	v := &review.Vote{
		ReviewID:  reviewID,
		UserID:    userID,
		VoteType:  review.VoteLike,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Save(ctx, v)

	if err := repo.Delete(ctx, reviewID, userID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	got, err := repo.FindByReviewAndUser(ctx, reviewID, userID)
	if err != nil {
		t.Fatalf("FindByReviewAndUser after delete: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil after delete")
	}

	var re repository.ReviewEntity
	db.Where("id = ?", reviewID).Take(&re)
	if re.LikeCount != 0 {
		t.Errorf("LikeCount after delete: got %d, want 0", re.LikeCount)
	}
}

func TestVoteRepository_FindByReviewAndUser_NotFound(t *testing.T) {
	_, repo, reviewID, _ := setupVoteTest(t)
	ctx := context.Background()

	got, err := repo.FindByReviewAndUser(ctx, reviewID, 99999)
	if err != nil {
		t.Fatalf("FindByReviewAndUser: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil for non-existent vote")
	}
}

func TestVoteRepository_CountTodayByUser(t *testing.T) {
	_, repo, reviewID, userID := setupVoteTest(t)
	ctx := context.Background()

	count, err := repo.CountTodayByUser(ctx, userID)
	if err != nil {
		t.Fatalf("CountTodayByUser empty: %v", err)
	}
	if count != 0 {
		t.Errorf("count before voting: got %d, want 0", count)
	}

	repo.Save(ctx, &review.Vote{
		ReviewID: reviewID, UserID: userID, VoteType: review.VoteLike,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	count, err = repo.CountTodayByUser(ctx, userID)
	if err != nil {
		t.Fatalf("CountTodayByUser after vote: %v", err)
	}
	if count != 1 {
		t.Errorf("count after one vote: got %d, want 1", count)
	}

	// Toggle counts as another operation
	repo.Save(ctx, &review.Vote{
		ReviewID: reviewID, UserID: userID, VoteType: review.VoteDislike,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	count, err = repo.CountTodayByUser(ctx, userID)
	if err != nil {
		t.Fatalf("CountTodayByUser after toggle: %v", err)
	}
	// After save, there's still only 1 row (upsert), so count = 1
	if count != 1 {
		t.Errorf("count after toggle (same row upsert): got %d, want 1", count)
	}
}

func TestVoteRepository_MultipleUsers(t *testing.T) {
	db, repo, reviewID, userID := setupVoteTest(t)
	ctx := context.Background()

	// Seed a second user
	e2 := repository.UserEntity{
		Username:     "testuser2",
		Email:        "testuser2@example.com",
		Role:         "user",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		LastSeenAt:   time.Now(),
	}
	if err := db.Create(&e2).Error; err != nil {
		t.Fatalf("seed user2: %v", err)
	}

	repo.Save(ctx, &review.Vote{
		ReviewID: reviewID, UserID: userID, VoteType: review.VoteLike,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	repo.Save(ctx, &review.Vote{
		ReviewID: reviewID, UserID: e2.ID, VoteType: review.VoteLike,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	var re repository.ReviewEntity
	db.Where("id = ?", reviewID).Take(&re)
	if re.LikeCount != 2 {
		t.Errorf("LikeCount with two users: got %d, want 2", re.LikeCount)
	}

	repo.Delete(ctx, reviewID, userID)

	db.Where("id = ?", reviewID).Take(&re)
	if re.LikeCount != 1 {
		t.Errorf("LikeCount after one user deletes: got %d, want 1", re.LikeCount)
	}
}
