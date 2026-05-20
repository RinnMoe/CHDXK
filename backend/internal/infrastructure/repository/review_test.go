package repository_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/review"
	"jcourse/internal/infrastructure/repository"
)

func TestReviewRepository_Create(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewReviewRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "reviews", "review_revisions", "courses", "teachers", "users")

	teacher := seedTeacher(t, db)
	course := seedCourse(t, db, teacher.ID)
	user := seedUser(t, db)

	r := &review.Review{
		CourseID:  course.ID,
		Semester:  "2024-2025-1",
		UserID:    user.ID,
		Rating:    5,
		Content:   "非常好的课程！",
		Score:     "A",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if r.ID == 0 {
		t.Fatal("expected non-zero ID after Create")
	}

	got, err := repo.Get(ctx, r.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.CourseID != course.ID {
		t.Errorf("CourseID: got %d, want %d", got.CourseID, course.ID)
	}
	if got.Rating != 5 {
		t.Errorf("Rating: got %d, want 5", got.Rating)
	}
	if got.Content != "非常好的课程！" {
		t.Errorf("Content: got %q, want %q", got.Content, "非常好的课程！")
	}
	if got.Score != "A" {
		t.Errorf("Score: got %q, want A", got.Score)
	}

	var count int64
	db.Model(&repository.CourseEntity{}).Where("id = ?", course.ID).Select("rating_count").Scan(&count)
	if count != 1 {
		t.Errorf("course rating_count: got %d, want 1", count)
	}
}

func TestReviewRepository_Update(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewReviewRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "reviews", "review_revisions", "courses", "teachers", "users")

	teacher := seedTeacher(t, db)
	course := seedCourse(t, db, teacher.ID)
	user := seedUser(t, db)
	entity := seedReview(t, db, course.ID, user.ID)

	r, err := repo.Get(ctx, entity.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	originalContent := r.Content
	rv := r.MakeRevision()
	r.Content = "更新后的评价"
	r.Rating = 4
	r.Score = "B"
	r.UpdatedAt = time.Now()

	if err := repo.Update(ctx, r, rv); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.Get(ctx, entity.ID)
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}
	if got.Content != "更新后的评价" {
		t.Errorf("Content: got %q, want %q", got.Content, "更新后的评价")
	}
	if got.Rating != 4 {
		t.Errorf("Rating: got %d, want 4", got.Rating)
	}
	if got.Score != "B" {
		t.Errorf("Score: got %q, want B", got.Score)
	}

	revisions, err := repo.FindRevisions(ctx, entity.ID)
	if err != nil {
		t.Fatalf("FindRevisions: %v", err)
	}
	if len(revisions) != 1 {
		t.Fatalf("revisions count: got %d, want 1", len(revisions))
	}
	if revisions[0].Content != originalContent {
		t.Errorf("revision Content: got %q, want %q", revisions[0].Content, originalContent)
	}
}

func TestReviewRepository_Delete(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewReviewRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "reviews", "review_revisions", "courses", "teachers", "users")

	teacher := seedTeacher(t, db)
	course := seedCourse(t, db, teacher.ID)
	user := seedUser(t, db)
	entity := seedReview(t, db, course.ID, user.ID)

	if err := repo.Delete(ctx, &review.Review{ID: entity.ID, CourseID: course.ID}); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.Get(ctx, entity.ID)
	if err == nil {
		t.Fatal("expected error after Delete, got nil")
	}

	var count int64
	db.Model(&repository.CourseEntity{}).Where("id = ?", course.ID).Select("rating_count").Scan(&count)
	if count != 0 {
		t.Errorf("course rating_count after delete: got %d, want 0", count)
	}
}

func TestReviewRepository_Get_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewReviewRepository(db)
	ctx := context.Background()

	_, err := repo.Get(ctx, 999999)
	if err == nil {
		t.Fatal("expected error for missing review, got nil")
	}
}

func TestReviewRepository_FindBy(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewReviewRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "reviews", "review_revisions", "courses", "teachers", "users")

	teacher := seedTeacher(t, db)
	course := seedCourse(t, db, teacher.ID)
	user := seedUser(t, db)

	r1 := repository.ReviewEntity{
		CourseID: course.ID, Semester: "2024-2025-1", UserID: user.ID,
		Rating: 5, Content: "很好", Score: "A",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	r2 := repository.ReviewEntity{
		CourseID: course.ID, Semester: "2024-2025-2", UserID: user.ID,
		Rating: 3, Content: "一般", Score: "C",
		CreatedAt: time.Now().Add(time.Hour), UpdatedAt: time.Now().Add(time.Hour),
	}
	for _, r := range []repository.ReviewEntity{r1, r2} {
		if err := db.Create(&r).Error; err != nil {
			t.Fatalf("seed review: %v", err)
		}
	}

	t.Run("find by course", func(t *testing.T) {
		results, err := repo.FindBy(ctx, review.ReviewFilter{CourseID: course.ID})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("count: got %d, want 2", len(results))
		}
	})

	t.Run("find by semester", func(t *testing.T) {
		results, err := repo.FindBy(ctx, review.ReviewFilter{Semester: "2024-2025-1"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("count: got %d, want 1", len(results))
		}
		if results[0].Rating != 5 {
			t.Errorf("Rating: got %d, want 5", results[0].Rating)
		}
	})

	t.Run("find by rating", func(t *testing.T) {
		results, err := repo.FindBy(ctx, review.ReviewFilter{Rating: 5})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("count: got %d, want 1", len(results))
		}
	})

	t.Run("find by user", func(t *testing.T) {
		results, err := repo.FindBy(ctx, review.ReviewFilter{UserID: user.ID})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("count: got %d, want 2", len(results))
		}
	})

	t.Run("find by course and semester", func(t *testing.T) {
		results, err := repo.FindBy(ctx, review.ReviewFilter{CourseID: course.ID, Semester: "2024-2025-1"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("count: got %d, want 1", len(results))
		}
	})
}

func TestReviewRepository_CourseStatsAggregation(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewReviewRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "reviews", "review_revisions", "courses", "teachers", "users")

	teacher := seedTeacher(t, db)
	course := seedCourse(t, db, teacher.ID)
	user := seedUser(t, db)

	ratings := []int{5, 4, 3}
	for i, rating := range ratings {
		r := &review.Review{
			CourseID:  course.ID,
			Semester:  "2024-2025-1",
			UserID:    user.ID,
			Rating:    rating,
			Content:   "评价内容",
			Score:     "B",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
			UpdatedAt: time.Now().Add(time.Duration(i) * time.Minute),
		}
		if err := repo.Create(ctx, r); err != nil {
			t.Fatalf("Create review: %v", err)
		}
	}

	var c repository.CourseEntity
	if err := db.Where("id = ?", course.ID).Take(&c).Error; err != nil {
		t.Fatalf("fetch course: %v", err)
	}
	if c.RatingCount != 3 {
		t.Errorf("ReviewCount: got %d, want 3", c.RatingCount)
	}
	avg := float64(5+4+3) / 3
	if c.RatingAvg != avg {
		t.Errorf("AvgRating: got %v, want %v", c.RatingAvg, avg)
	}
}

func TestReviewRepository_FindRevisions(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewReviewRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "reviews", "review_revisions", "courses", "teachers", "users")

	teacher := seedTeacher(t, db)
	course := seedCourse(t, db, teacher.ID)
	user := seedUser(t, db)
	entity := seedReview(t, db, course.ID, user.ID)

	r, _ := repo.Get(ctx, entity.ID)
	rv1 := r.MakeRevision()
	r.Content = "第一次更新"
	r.Rating = 4
	r.UpdatedAt = time.Now()
	repo.Update(ctx, r, rv1)

	rv2 := r.MakeRevision()
	r.Content = "第二次更新"
	r.Rating = 3
	r.UpdatedAt = time.Now()
	repo.Update(ctx, r, rv2)

	revs, err := repo.FindRevisions(ctx, entity.ID)
	if err != nil {
		t.Fatalf("FindRevisions: %v", err)
	}
	if len(revs) != 2 {
		t.Fatalf("revisions count: got %d, want 2", len(revs))
	}
}
