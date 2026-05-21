package repository_test

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/review"
	"jcourse/internal/domain/stat"
	"jcourse/internal/infrastructure/repository"
)

func TestSiteDailyStatRepository_Collect(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewSiteDailyStatRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "site_daily_stats", "review_votes", "reviews", "courses", "teachers", "users")

	loc := mustRepoStatsLocation(t)
	periodStart := time.Date(2026, 5, 20, 0, 0, 0, 0, loc)
	periodEnd := periodStart.AddDate(0, 0, 1)
	before := periodStart.Add(-time.Hour)
	during := periodStart.Add(10 * time.Hour)
	after := periodEnd.Add(time.Hour)

	teacher := seedTeacher(t, db)
	course1 := seedCourseRaw(t, db, "CS101", "数据结构", 3, "计算机学院", teacher.ID, "zh", []string{"核心课"}, []string{"2026"})
	course2 := seedCourseRaw(t, db, "CS102", "算法", 3, "计算机学院", teacher.ID, "zh", []string{"核心课"}, []string{"2026"})
	course3 := seedCourseRaw(t, db, "CS103", "数据库", 3, "计算机学院", teacher.ID, "zh", []string{"核心课"}, []string{"2026"})

	author1 := seedStatUser(t, db, "author1", "author1@example.com", during, during)
	author2 := seedStatUser(t, db, "author2", "author2@example.com", before, during)
	activeOnly := seedStatUser(t, db, "active", "active@example.com", before, during)
	oldUser := seedStatUser(t, db, "old", "old@example.com", before, before)

	r1 := seedStatReview(t, db, course1.ID, author1.ID, during)
	r2 := seedStatReview(t, db, course2.ID, author1.ID, during.Add(time.Hour))
	r3 := seedStatReview(t, db, course1.ID, author2.ID, during.Add(2*time.Hour))
	r4 := seedStatReview(t, db, course2.ID, oldUser.ID, before)
	r5 := seedStatReview(t, db, course3.ID, author2.ID, after)

	seedStatVote(t, db, r1.ID, activeOnly.ID, review.VoteLike, during)
	seedStatVote(t, db, r2.ID, author2.ID, review.VoteLike, during.Add(time.Hour))
	seedStatVote(t, db, r3.ID, author1.ID, review.VoteDislike, during.Add(2*time.Hour))
	seedStatVote(t, db, r4.ID, author1.ID, review.VoteLike, before)
	seedStatVote(t, db, r5.ID, oldUser.ID, review.VoteDislike, after)

	metrics, err := repo.Collect(ctx, periodStart, periodEnd)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}

	want := stat.Metrics{
		stat.MetricActiveUserCount:     3,
		stat.MetricNewUserCount:        1,
		stat.MetricNewReviewCount:      3,
		stat.MetricReviewAuthorCount:   2,
		stat.MetricReviewedCourseTotal: 2,
		stat.MetricNewLikeCount:        2,
		stat.MetricNewDislikeCount:     1,
	}
	for key, wantValue := range want {
		if metrics[key] != wantValue {
			t.Fatalf("metric %s = %d, want %d; all metrics = %+v", key, metrics[key], wantValue, metrics)
		}
	}
}

func TestSiteDailyStatRepository_UpsertAndFindByDateRange(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewSiteDailyStatRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "site_daily_stats")

	loc := mustRepoStatsLocation(t)
	date18 := time.Date(2026, 5, 18, 0, 0, 0, 0, loc)
	date19 := time.Date(2026, 5, 19, 0, 0, 0, 0, loc)
	date20 := time.Date(2026, 5, 20, 0, 0, 0, 0, loc)

	upsertDailyStat(t, ctx, repo, date18, 18)
	upsertDailyStat(t, ctx, repo, date19, 19)
	upsertDailyStat(t, ctx, repo, date20, 20)
	upsertDailyStat(t, ctx, repo, date19, 190)

	got, err := repo.GetByDate(ctx, date19)
	if err != nil {
		t.Fatalf("GetByDate: %v", err)
	}
	if got.ActiveUserCount != 190 || got.NewReviewCount != 191 {
		t.Fatalf("upserted stat = %+v", got)
	}

	items, total, err := repo.FindByDateRange(ctx, stat.DailyStatFilter{
		StartDate: date18,
		EndDate:   date20,
		Page:      1,
		PageSize:  2,
	})
	if err != nil {
		t.Fatalf("FindByDateRange: %v", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if len(items) != 2 {
		t.Fatalf("items count = %d, want 2", len(items))
	}
	if items[0].StatDate.Format("2006-01-02") != "2026-05-20" || items[1].StatDate.Format("2006-01-02") != "2026-05-19" {
		t.Fatalf("dates = %s, %s; want 2026-05-20, 2026-05-19", items[0].StatDate, items[1].StatDate)
	}
	if items[1].ActiveUserCount != 190 || items[1].NewReviewCount != 191 {
		t.Fatalf("flattened item = %+v", items[1])
	}
}

func seedStatUser(t *testing.T, db *gorm.DB, username, email string, createdAt, lastSeenAt time.Time) repository.UserEntity {
	t.Helper()
	e := repository.UserEntity{
		Username:   username,
		Email:      email,
		Role:       "user",
		Password:   "hashed_password",
		CreatedAt:  createdAt,
		LastSeenAt: lastSeenAt,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed stat user: %v", err)
	}
	return e
}

func seedStatReview(t *testing.T, db *gorm.DB, courseID, userID int, at time.Time) repository.ReviewEntity {
	t.Helper()
	e := repository.ReviewEntity{
		CourseID:  courseID,
		Semester:  "2025-2026-2",
		UserID:    userID,
		Rating:    5,
		Content:   "不错",
		Score:     "A",
		CreatedAt: at,
		UpdatedAt: at,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed stat review: %v", err)
	}
	return e
}

func seedStatVote(t *testing.T, db *gorm.DB, reviewID, userID, voteType int, at time.Time) repository.ReviewVoteEntity {
	t.Helper()
	e := repository.ReviewVoteEntity{
		ReviewID:  reviewID,
		UserID:    userID,
		VoteType:  voteType,
		CreatedAt: at,
		UpdatedAt: at,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed stat vote: %v", err)
	}
	return e
}

func upsertDailyStat(t *testing.T, ctx context.Context, repo *repository.SiteDailyStatRepository, date time.Time, base int64) {
	t.Helper()
	err := repo.Upsert(ctx, &stat.DailyStat{
		StatDate: date,
		Metrics: stat.Metrics{
			stat.MetricActiveUserCount:     base,
			stat.MetricNewUserCount:        base + 1,
			stat.MetricNewReviewCount:      base + 1,
			stat.MetricReviewAuthorCount:   base + 2,
			stat.MetricReviewedCourseTotal: base + 3,
			stat.MetricNewLikeCount:        base + 4,
			stat.MetricNewDislikeCount:     base + 5,
		},
		GeneratedAt: date.Add(time.Hour),
		UpdatedAt:   date.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
}

func mustRepoStatsLocation(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	return loc
}
