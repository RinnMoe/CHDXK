package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/lib/pq"

	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
	"jcourse/internal/infrastructure/repository"
)

func TestCourseRepository_Get(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "courses", "teachers")
	teacher := seedTeacher(t, db)
	courseEntity := seedCourse(t, db, teacher.ID)

	c, err := repo.Get(ctx, courseEntity.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if c.Code != courseEntity.Code {
		t.Errorf("Code: got %q, want %q", c.Code, courseEntity.Code)
	}
	if c.Name != courseEntity.Name {
		t.Errorf("Name: got %q, want %q", c.Name, courseEntity.Name)
	}
	if c.Credit != courseEntity.Credit {
		t.Errorf("Credit: got %v, want %v", c.Credit, courseEntity.Credit)
	}
	if c.MainTeacherID != teacher.ID {
		t.Errorf("MainTeacherID: got %d, want %d", c.MainTeacherID, teacher.ID)
	}
}

func TestCourseRepository_Get_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()

	got, err := repo.Get(ctx, 999999)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != nil {
		t.Fatalf("Get got %+v, want nil", got)
	}
}

func TestCourseRepository_FindBy(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "courses", "teachers", "reviews")

	t1 := seedTeacher(t, db)
	t2 := repository.TeacherEntity{Code: "T002", Name: "李老师", Department: "数学学院", Title: "副教授", Pinyin: "lilaoshi", PinyinAbbr: "lls"}
	if err := db.Create(&t2).Error; err != nil {
		t.Fatalf("seed teacher t2: %v", err)
	}

	c1 := seedCourseRaw(t, db, "CS101", "数据结构", 3.0, "计算机学院", t1.ID, "zh", []string{"核心课"}, []string{"2021"})
	seedCourseRaw(t, db, "CS102", "算法设计", 3.0, "计算机学院", t1.ID, "en", []string{"选修课"}, []string{"2022"})
	seedCourseRaw(t, db, "MA101", "高等数学", 4.0, "数学学院", t2.ID, "zh", []string{"核心课"}, []string{"2021", "2022"})
	if err := db.Model(&repository.CourseEntity{}).Where("id = ?", c1.ID).Update("rating_count", 1).Error; err != nil {
		t.Fatalf("mark course reviewed: %v", err)
	}

	t.Run("list all", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, course.CourseFilter{})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 3 {
			t.Errorf("total: got %d, want 3", total)
		}
		if len(results) != 3 {
			t.Errorf("count: got %d, want 3", len(results))
		}
	})

	t.Run("filter by department", func(t *testing.T) {
		_, total, err := repo.FindBy(ctx, course.CourseFilter{Department: "计算机学院"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 2 {
			t.Errorf("total: got %d, want 2", total)
		}
	})

	t.Run("filter by code", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, course.CourseFilter{Code: "cs101"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 1 {
			t.Errorf("total: got %d, want 1", total)
		}
		if results[0].Name != "数据结构" {
			t.Errorf("Name: got %q, want 数据结构", results[0].Name)
		}
	})

	t.Run("filter by language", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, course.CourseFilter{Language: "en"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 1 {
			t.Errorf("total: got %d, want 1", total)
		}
		if results[0].Code != "CS102" {
			t.Errorf("Code: got %q, want CS102", results[0].Code)
		}
	})

	t.Run("filter by teacher_id", func(t *testing.T) {
		_, total, err := repo.FindBy(ctx, course.CourseFilter{TeacherID: t1.ID})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 2 {
			t.Errorf("total: got %d, want 2", total)
		}
	})

	t.Run("filter by categories", func(t *testing.T) {
		_, total, err := repo.FindBy(ctx, course.CourseFilter{Categories: []string{"核心课"}})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 2 {
			t.Errorf("total: got %d, want 2", total)
		}
	})

	t.Run("filter by target_years", func(t *testing.T) {
		_, total, err := repo.FindBy(ctx, course.CourseFilter{TargetYears: []string{"2022"}})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 2 {
			t.Errorf("total: got %d, want 2", total)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, course.CourseFilter{Page: 1, PageSize: 2})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 3 {
			t.Errorf("total: got %d, want 3", total)
		}
		if len(results) != 2 {
			t.Errorf("page 1 count: got %d, want 2", len(results))
		}

		results2, _, err := repo.FindBy(ctx, course.CourseFilter{Page: 2, PageSize: 2})
		if err != nil {
			t.Fatalf("FindBy page 2: %v", err)
		}
		if len(results2) != 1 {
			t.Errorf("page 2 count: got %d, want 1", len(results2))
		}
	})

	t.Run("default sort honors ascend", func(t *testing.T) {
		results, _, err := repo.FindBy(ctx, course.CourseFilter{Ascend: true})
		if err != nil {
			t.Fatalf("FindBy default ascend: %v", err)
		}
		if len(results) != 3 {
			t.Fatalf("count: got %d, want 3", len(results))
		}
		if results[0].ID != c1.ID {
			t.Errorf("first id: got %d, want %d", results[0].ID, c1.ID)
		}
	})

	t.Run("exclude_id", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, course.CourseFilter{ExcludeID: c1.ID})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 2 {
			t.Errorf("total: got %d, want 2", total)
		}
		for _, r := range results {
			if r.ID == c1.ID {
				t.Errorf("ExcludeID should have filtered out c1 (id=%d)", c1.ID)
			}
		}
	})

	t.Run("filter by course_ids", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, course.CourseFilter{CourseIDs: []int{c1.ID, c1.ID + 1}})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 2 {
			t.Errorf("total: got %d, want 2", total)
		}
		if len(results) != 2 {
			t.Errorf("count: got %d, want 2", len(results))
		}
	})

	t.Run("filter by has_review", func(t *testing.T) {
		hasReview := true
		results, total, err := repo.FindBy(ctx, course.CourseFilter{HasReview: &hasReview})
		if err != nil {
			t.Fatalf("FindBy has_review=true: %v", err)
		}
		if total != 1 || len(results) != 1 {
			t.Fatalf("reviewed count: got total=%d len=%d, want 1", total, len(results))
		}
		if results[0].ID != c1.ID {
			t.Errorf("reviewed course id: got %d, want %d", results[0].ID, c1.ID)
		}

		hasReview = false
		results, total, err = repo.FindBy(ctx, course.CourseFilter{HasReview: &hasReview})
		if err != nil {
			t.Fatalf("FindBy has_review=false: %v", err)
		}
		if total != 2 || len(results) != 2 {
			t.Fatalf("unreviewed count: got total=%d len=%d, want 2", total, len(results))
		}
		for _, r := range results {
			if r.ID == c1.ID {
				t.Errorf("HasReview=false should have filtered out reviewed course id=%d", c1.ID)
			}
		}
	})

	t.Run("sort by rating_count", func(t *testing.T) {
		results, _, err := repo.FindBy(ctx, course.CourseFilter{OrderBy: "rating_count"})
		if err != nil {
			t.Fatalf("FindBy rating_count desc: %v", err)
		}
		if len(results) == 0 || results[0].ID != c1.ID {
			t.Fatalf("desc first id: got %v, want %d", results, c1.ID)
		}

		results, _, err = repo.FindBy(ctx, course.CourseFilter{OrderBy: "rating_count", Ascend: true})
		if err != nil {
			t.Fatalf("FindBy rating_count ascend: %v", err)
		}
		if len(results) == 0 || results[len(results)-1].ID != c1.ID {
			t.Fatalf("asc last id: got %v, want %d", results, c1.ID)
		}
	})
}

func TestCourseRepository_GetDetail(t *testing.T) {
	db := newTestDB(t)
	courseRepo := repository.NewCourseRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "offered_courses", "reviews", "courses", "teachers", "semesters")

	teacher := seedTeacher(t, db)
	courseEntity := seedCourse(t, db, teacher.ID)
	user := seedUser(t, db)

	oc := seedOfferedCourseRaw(t, db, courseEntity.ID, "2024-2025-1", "zh", []string{"核心课"}, []string{"2022"})
	db.Model(&repository.OfferedCourseEntity{}).Where("id = ?", oc.ID).
		Update("teacher_ids", pq.Int64Array{int64(teacher.ID)})

	now := time.Now()
	r1 := review.Review{
		CourseID:  courseEntity.ID,
		Semester:  "2024-2025-1",
		UserID:    user.ID,
		Rating:    5,
		Content:   "很好的课程，老师讲解清晰。",
		Score:     "A",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := reviewRepo.Create(ctx, &r1); err != nil {
		t.Fatalf("create review 1: %v", err)
	}
	r2 := review.Review{
		CourseID:  courseEntity.ID,
		Semester:  "2024-2025-1",
		UserID:    user.ID,
		Rating:    3,
		Content:   "一般般。",
		Score:     "B",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := reviewRepo.Create(ctx, &r2); err != nil {
		t.Fatalf("create review 2: %v", err)
	}

	detail, err := courseRepo.GetDetail(ctx, courseEntity.ID)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if detail.Code != courseEntity.Code {
		t.Errorf("Code: got %q, want %q", detail.Code, courseEntity.Code)
	}
	if detail.Rating.Count != 2 {
		t.Errorf("Rating.Count: got %d, want 2", detail.Rating.Count)
	}
	avg := float64(5+3) / 2
	if detail.Rating.Avg != avg {
		t.Errorf("Rating.Avg: got %v, want %v", detail.Rating.Avg, avg)
	}
	expectedDist := [5]int{0, 0, 1, 0, 1}
	if detail.Rating.Distribution != expectedDist {
		t.Errorf("Rating.Distribution: got %v, want %v", detail.Rating.Distribution, expectedDist)
	}
	if len(detail.OfferedCourses) != 1 {
		t.Fatalf("OfferedCourses count: got %d, want 1", len(detail.OfferedCourses))
	}
	if detail.OfferedCourses[0].Semester != "2024-2025-1" {
		t.Errorf("OfferedCourses[0].Semester: got %q, want 2024-2025-1", detail.OfferedCourses[0].Semester)
	}
	if len(detail.OfferedCourses[0].TeacherGroup) != 1 {
		t.Errorf("TeacherGroup count: got %d, want 1", len(detail.OfferedCourses[0].TeacherGroup))
	}
}

func TestCourseRepository_FindOfferedCourses(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "offered_courses", "courses", "teachers")

	teacher := seedTeacher(t, db)
	courseEntity := seedCourse(t, db, teacher.ID)

	t2 := repository.TeacherEntity{Code: "T002", Name: "李老师", Department: "计算机学院", Title: "副教授", Pinyin: "lilaoshi", PinyinAbbr: "lls"}
	if err := db.Create(&t2).Error; err != nil {
		t.Fatalf("seed teacher t2: %v", err)
	}

	oc1 := seedOfferedCourseRaw(t, db, courseEntity.ID, "2023-2024-1", "zh", []string{"核心课"}, []string{"2021"})
	oc2 := seedOfferedCourseRaw(t, db, courseEntity.ID, "2024-2025-1", "en", []string{"选修课"}, []string{"2022"})

	db.Model(&repository.OfferedCourseEntity{}).Where("id = ?", oc1.ID).
		Update("teacher_ids", pq.Int64Array{int64(teacher.ID)})
	db.Model(&repository.OfferedCourseEntity{}).Where("id = ?", oc2.ID).
		Update("teacher_ids", pq.Int64Array{int64(t2.ID)})

	ocs, err := repo.FindOfferedCourses(ctx, courseEntity.ID)
	if err != nil {
		t.Fatalf("FindOfferedCourses: %v", err)
	}
	if len(ocs) != 2 {
		t.Fatalf("count: got %d, want 2", len(ocs))
	}
	if len(ocs[0].TeacherGroup) != 1 {
		t.Errorf("first offered course teacher group count: got %d, want 1", len(ocs[0].TeacherGroup))
	}
}

func TestCourseRepository_GetFilters(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "courses", "teachers")

	t1 := seedTeacher(t, db)
	t2 := repository.TeacherEntity{Code: "T002", Name: "李老师", Department: "数学学院", Title: "副教授", Pinyin: "lilaoshi", PinyinAbbr: "lls"}
	if err := db.Create(&t2).Error; err != nil {
		t.Fatalf("seed teacher t2: %v", err)
	}

	seedCourseRaw(t, db, "CS101", "数据结构", 3.0, "计算机学院", t1.ID, "zh", []string{"核心课"}, []string{"2021"})
	seedCourseRaw(t, db, "CS102", "算法设计", 3.0, "计算机学院", t1.ID, "en", []string{"选修课"}, []string{"2022"})
	seedCourseRaw(t, db, "MA101", "高等数学", 4.0, "数学学院", t2.ID, "zh", []string{"核心课"}, []string{"2021", "2022"})

	filters, err := repo.GetFilters(ctx)
	if err != nil {
		t.Fatalf("GetFilters: %v", err)
	}

	t.Run("credits", func(t *testing.T) {
		if len(filters.Credits) != 2 {
			t.Fatalf("credits count: got %d, want 2", len(filters.Credits))
		}
		if filters.Credits[0].Name != "3" || filters.Credits[0].Count != 2 {
			t.Errorf("credit 3: got %+v, want name=3 count=2", filters.Credits[0])
		}
		if filters.Credits[1].Name != "4" || filters.Credits[1].Count != 1 {
			t.Errorf("credit 4: got %+v, want name=4 count=1", filters.Credits[1])
		}
	})

	t.Run("departments", func(t *testing.T) {
		if len(filters.Departments) != 2 {
			t.Fatalf("departments count: got %d, want 2", len(filters.Departments))
		}
		if filters.Departments[0].Name != "数学学院" || filters.Departments[0].Count != 1 {
			t.Errorf("department 数学学院: got %+v", filters.Departments[0])
		}
		if filters.Departments[1].Name != "计算机学院" || filters.Departments[1].Count != 2 {
			t.Errorf("department 计算机学院: got %+v", filters.Departments[1])
		}
	})

	t.Run("categories", func(t *testing.T) {
		if len(filters.Categories) != 2 {
			t.Fatalf("categories count: got %d, want 2", len(filters.Categories))
		}
		if filters.Categories[0].Name != "核心课" || filters.Categories[0].Count != 2 {
			t.Errorf("category 核心课: got %+v", filters.Categories[0])
		}
		if filters.Categories[1].Name != "选修课" || filters.Categories[1].Count != 1 {
			t.Errorf("category 选修课: got %+v", filters.Categories[1])
		}
	})

	t.Run("target_years", func(t *testing.T) {
		if len(filters.TargetYears) != 2 {
			t.Fatalf("target_years count: got %d, want 2", len(filters.TargetYears))
		}
		if filters.TargetYears[0].Name != "2021" || filters.TargetYears[0].Count != 2 {
			t.Errorf("target_year 2021: got %+v", filters.TargetYears[0])
		}
		if filters.TargetYears[1].Name != "2022" || filters.TargetYears[1].Count != 2 {
			t.Errorf("target_year 2022: got %+v", filters.TargetYears[1])
		}
	})
}

func TestCourseRepository_OfferedCourseExists(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "offered_courses", "courses", "teachers")

	teacher := seedTeacher(t, db)
	courseEntity := seedCourse(t, db, teacher.ID)

	seedOfferedCourseRaw(t, db, courseEntity.ID, "2024-2025-1", "zh", nil, nil)

	exists, err := repo.OfferedCourseExists(ctx, courseEntity.ID, "2024-2025-1")
	if err != nil {
		t.Fatalf("OfferedCourseExists: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}

	exists, err = repo.OfferedCourseExists(ctx, courseEntity.ID, "2020-2021-1")
	if err != nil {
		t.Fatalf("OfferedCourseExists for missing: %v", err)
	}
	if exists {
		t.Error("expected exists=false for missing semester")
	}
}
