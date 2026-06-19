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
	t2 := seedTeacherRaw(t, db, "T002", "李老师", "数学学院", "副教授")

	c1 := seedCourseRaw(t, db, "CS101", "数据结构", 3.0, "计算机学院", t1.ID, "zh", []string{"核心课"}, []string{"2021"})
	seedCourseRaw(t, db, "CS102", "算法设计", 3.0, "计算机学院", t1.ID, "en", []string{"选修课"}, []string{"2022"})
	seedCourseRaw(t, db, "MA101", "高等数学", 4.0, "数学学院", t2.ID, "zh", []string{"核心课"}, []string{"2021", "2022"})
	if err := db.Model(&repository.CourseEntity{}).Where("id = ?", c1.ID).Update("rating_count", 1).Error; err != nil {
		t.Fatalf("mark course reviewed: %v", err)
	}
	if err := repository.RefreshCourseSearchVectors(db); err != nil {
		t.Fatalf("refresh course search vectors: %v", err)
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

	t.Run("search by q course name", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, course.CourseFilter{Q: "数据结构"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 1 {
			t.Errorf("total: got %d, want 1", total)
		}
		if len(results) == 0 || results[0].Code != "CS101" {
			t.Errorf("Code: got %v, want CS101", results)
		}
	})

	t.Run("search by q course name substring", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, course.CourseFilter{Q: "数学"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 1 {
			t.Errorf("total: got %d, want 1", total)
		}
		if len(results) == 0 || results[0].Code != "MA101" {
			t.Errorf("Code: got %v, want MA101", results)
		}
	})

	t.Run("search by q course code prefix", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, course.CourseFilter{Q: "cs10"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 2 {
			t.Errorf("total: got %d, want 2", total)
		}
		gotCodes := map[string]bool{}
		for _, result := range results {
			gotCodes[result.Code] = true
		}
		if !gotCodes["CS101"] || !gotCodes["CS102"] {
			t.Errorf("codes: got %v, want CS101 and CS102", results)
		}
	})

	t.Run("search by q course code with symbols", func(t *testing.T) {
		symbolCode := seedCourseRaw(t, db, "CS-10:3", "带符号课程代码", 2.0, "计算机学院", t1.ID, "zh", []string{"选修课"}, []string{"2023"})
		t.Cleanup(func() {
			_ = db.Delete(&repository.CourseEntity{}, symbolCode.ID).Error
		})

		results, total, err := repo.FindBy(ctx, course.CourseFilter{Q: "CS-10:"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 1 {
			t.Errorf("total: got %d, want 1", total)
		}
		if len(results) == 0 || results[0].Code != "CS-10:3" {
			t.Errorf("Code: got %v, want CS-10:3", results)
		}
	})

	t.Run("search by q teacher name", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, course.CourseFilter{Q: "张三"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 2 {
			t.Errorf("total: got %d, want 2", total)
		}
		if len(results) == 0 || results[0].MainTeacher == nil || results[0].MainTeacher.Name != "张三" {
			t.Errorf("MainTeacher: got %v, want 张三", results)
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

	t.Run("filter by main_teacher_name", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, course.CourseFilter{MainTeacherName: "张三"})
		if err != nil {
			t.Fatalf("FindBy main_teacher_name=张三: %v", err)
		}
		if total != 2 {
			t.Errorf("total: got %d, want 2", total)
		}
		if len(results) != 2 {
			t.Fatalf("count: got %d, want 2", len(results))
		}
		for _, r := range results {
			if r.MainTeacher == nil || r.MainTeacher.Name != "张三" {
				t.Errorf("MainTeacher.Name: got %v, want 张三", r.MainTeacher)
			}
		}

		results, total, err = repo.FindBy(ctx, course.CourseFilter{MainTeacherName: "李老师"})
		if err != nil {
			t.Fatalf("FindBy main_teacher_name=李老师: %v", err)
		}
		if total != 1 {
			t.Errorf("total: got %d, want 1", total)
		}
		if len(results) == 0 || results[0].Code != "MA101" {
			t.Errorf("Code: got %v, want MA101", results)
		}

		results, total, err = repo.FindBy(ctx, course.CourseFilter{MainTeacherName: "不存在"})
		if err != nil {
			t.Fatalf("FindBy main_teacher_name=不存在: %v", err)
		}
		if total != 0 {
			t.Errorf("total: got %d, want 0", total)
		}
	})

	t.Run("default sort by rating_score", func(t *testing.T) {
		highScore := seedCourseRaw(t, db, "RS100", "综合评分更高", 3.0, "计算机学院", t1.ID, "zh", []string{"核心课"}, []string{"2021"})
		highCount := seedCourseRaw(t, db, "RS101", "综合评分相同点评更多", 3.0, "计算机学院", t1.ID, "zh", []string{"核心课"}, []string{"2021"})
		codeFirst := seedCourseRaw(t, db, "RS102", "综合评分相同点评相同", 3.0, "计算机学院", t1.ID, "zh", []string{"核心课"}, []string{"2021"})
		t.Cleanup(func() {
			_ = db.Delete(&repository.CourseEntity{}, []int{highScore.ID, highCount.ID, codeFirst.ID}).Error
		})

		for _, tc := range []struct {
			courseID int
			count    int
			avg      float64
			score    float64
		}{
			{courseID: c1.ID, count: 2, avg: 4, score: 4.1},
			{courseID: highScore.ID, count: 1, avg: 5, score: 4.5},
			{courseID: highCount.ID, count: 3, avg: 4, score: 4.1},
			{courseID: codeFirst.ID, count: 2, avg: 4, score: 4.1},
		} {
			if err := db.Model(&repository.CourseEntity{}).
				Where("id = ?", tc.courseID).
				Updates(map[string]any{"rating_count": tc.count, "rating_avg": tc.avg, "rating_score": tc.score}).Error; err != nil {
				t.Fatalf("update course rating: %v", err)
			}
		}

		results, _, err := repo.FindBy(ctx, course.CourseFilter{})
		if err != nil {
			t.Fatalf("FindBy default rating_score desc: %v", err)
		}
		if len(results) < 4 {
			t.Fatalf("results: got %d, want at least 4", len(results))
		}
		if results[0].ID != highScore.ID || results[1].ID != highCount.ID || results[2].ID != c1.ID || results[3].ID != codeFirst.ID {
			t.Fatalf("order ids: got %d,%d,%d,%d", results[0].ID, results[1].ID, results[2].ID, results[3].ID)
		}
	})

	t.Run("sort by rating_count", func(t *testing.T) {
		highScore := seedCourseRaw(t, db, "RC100", "点评数相同综合评分更高", 3.0, "计算机学院", t1.ID, "zh", []string{"核心课"}, []string{"2021"})
		highCount := seedCourseRaw(t, db, "RC101", "点评数更多", 3.0, "计算机学院", t1.ID, "zh", []string{"核心课"}, []string{"2021"})
		codeFirst := seedCourseRaw(t, db, "RC102", "点评数相同code更后", 3.0, "计算机学院", t1.ID, "zh", []string{"核心课"}, []string{"2021"})
		t.Cleanup(func() {
			_ = db.Delete(&repository.CourseEntity{}, []int{highScore.ID, highCount.ID, codeFirst.ID}).Error
		})

		for _, tc := range []struct {
			courseID int
			count    int
			score    float64
		}{
			{courseID: c1.ID, count: 2, score: 4.0},
			{courseID: highScore.ID, count: 2, score: 4.5},
			{courseID: highCount.ID, count: 3, score: 1.0},
			{courseID: codeFirst.ID, count: 2, score: 4.0},
		} {
			if err := db.Model(&repository.CourseEntity{}).
				Where("id = ?", tc.courseID).
				Updates(map[string]any{"rating_count": tc.count, "rating_score": tc.score}).Error; err != nil {
				t.Fatalf("update course rating: %v", err)
			}
		}

		results, _, err := repo.FindBy(ctx, course.CourseFilter{OrderBy: "rating_count"})
		if err != nil {
			t.Fatalf("FindBy rating_count desc: %v", err)
		}
		if len(results) < 4 {
			t.Fatalf("results: got %d, want at least 4", len(results))
		}
		if results[0].ID != highCount.ID {
			t.Fatalf("desc first id: got %v, want %d", results, highCount.ID)
		}
		if results[1].ID != highScore.ID {
			t.Fatalf("desc second id: got %v, want %d", results, highScore.ID)
		}
		if results[2].ID != c1.ID {
			t.Fatalf("desc third id: got %v, want %d", results, c1.ID)
		}
		if results[3].ID != codeFirst.ID {
			t.Fatalf("desc fourth id: got %v, want %d", results, codeFirst.ID)
		}

		results, _, err = repo.FindBy(ctx, course.CourseFilter{OrderBy: "rating_count", Ascend: true})
		if err != nil {
			t.Fatalf("FindBy rating_count ascend: %v", err)
		}
		if len(results) == 0 || results[len(results)-1].ID != highCount.ID {
			t.Fatalf("asc last id: got %v, want %d", results, highCount.ID)
		}
	})

	t.Run("rating_avg order is unsupported", func(t *testing.T) {
		lowAvgCodeFirst := seedCourseRaw(t, db, "AA001", "均分低但编码靠前", 3.0, "计算机学院", t1.ID, "zh", []string{"核心课"}, []string{"2021"})
		highAvgCodeLast := seedCourseRaw(t, db, "ZZ999", "均分高但编码靠后", 3.0, "计算机学院", t1.ID, "zh", []string{"核心课"}, []string{"2021"})
		t.Cleanup(func() {
			_ = db.Delete(&repository.CourseEntity{}, []int{lowAvgCodeFirst.ID, highAvgCodeLast.ID}).Error
		})

		for _, tc := range []struct {
			courseID int
			count    int
			avg      float64
		}{
			{courseID: lowAvgCodeFirst.ID, count: 1, avg: 1},
			{courseID: highAvgCodeLast.ID, count: 1, avg: 5},
		} {
			if err := db.Model(&repository.CourseEntity{}).
				Where("id = ?", tc.courseID).
				Updates(map[string]any{"rating_count": tc.count, "rating_avg": tc.avg}).Error; err != nil {
				t.Fatalf("update course rating: %v", err)
			}
		}

		results, _, err := repo.FindBy(ctx, course.CourseFilter{OrderBy: "rating_avg"})
		if err != nil {
			t.Fatalf("FindBy unsupported rating_avg: %v", err)
		}
		if len(results) == 0 || results[0].ID != lowAvgCodeFirst.ID {
			t.Fatalf("first id: got %v, want %d", results, lowAvgCodeFirst.ID)
		}
	})
}

func TestCourseRepository_FindBy_DefaultSortByCode(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "courses", "teachers", "reviews")

	teacher := seedTeacher(t, db)
	otherTeacher := seedTeacherRaw(t, db, "T999", "排序老师", "测试学院", "讲师")

	seedCourseRaw(t, db, "ZZ200", "后插入课程", 3.0, "测试学院", teacher.ID, "zh", []string{"核心课"}, []string{"2024"})
	seedCourseRaw(t, db, "AA100", "先排序课程", 3.0, "测试学院", otherTeacher.ID, "zh", []string{"核心课"}, []string{"2024"})
	seedCourseRaw(t, db, "MM150", "中间排序课程", 3.0, "测试学院", teacher.ID, "zh", []string{"核心课"}, []string{"2024"})
	if err := repository.RefreshCourseSearchVectors(db); err != nil {
		t.Fatalf("refresh course search vectors: %v", err)
	}

	results, total, err := repo.FindBy(ctx, course.CourseFilter{})
	if err != nil {
		t.Fatalf("FindBy: %v", err)
	}
	if total != 3 {
		t.Fatalf("total: got %d, want 3", total)
	}
	if got := []string{results[0].Code, results[1].Code, results[2].Code}; got[0] != "AA100" || got[1] != "MM150" || got[2] != "ZZ200" {
		t.Fatalf("codes: got %v, want [AA100 MM150 ZZ200]", got)
	}
}

func TestRefreshCourseSearchVectorsWithSemester(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "courses", "teachers", "reviews")

	currentTeacher := seedTeacherRaw(t, db, "T-CURRENT", "当前老师", "测试学院", "教授")
	previousTeacher := seedTeacherRaw(t, db, "T-PREVIOUS", "历史老师", "测试学院", "教授")
	current := seedCourseRaw(t, db, "CUR101", "当前学期课程", 3.0, "测试学院", currentTeacher.ID, "zh", []string{"核心课"}, []string{"2025"})
	previous := seedCourseRaw(t, db, "OLD101", "历史学期课程", 3.0, "测试学院", previousTeacher.ID, "zh", []string{"核心课"}, []string{"2024"})
	if err := db.Model(&repository.CourseEntity{}).
		Where("id = ?", current.ID).
		Update("last_semester", "2025-2026-1").Error; err != nil {
		t.Fatalf("update current semester: %v", err)
	}
	if err := db.Model(&repository.CourseEntity{}).
		Where("id = ?", previous.ID).
		Update("last_semester", "2024-2025-2").Error; err != nil {
		t.Fatalf("update previous semester: %v", err)
	}

	if err := repository.RefreshCourseSearchVectors(db, "2025-2026-1"); err != nil {
		t.Fatalf("refresh course search vectors: %v", err)
	}

	results, total, err := repo.FindBy(ctx, course.CourseFilter{Q: "当前老师"})
	if err != nil {
		t.Fatalf("FindBy current course: %v", err)
	}
	if total != 1 || len(results) != 1 || results[0].ID != current.ID {
		t.Fatalf("current search results: total=%d results=%+v, want only %d", total, results, current.ID)
	}

	results, total, err = repo.FindBy(ctx, course.CourseFilter{Q: "历史老师"})
	if err != nil {
		t.Fatalf("FindBy previous course: %v", err)
	}
	if total != 0 || len(results) != 0 {
		t.Fatalf("previous search results: total=%d results=%+v, want none", total, results)
	}
}

func TestCourseRepository_FindBy_ZeroCredit(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "courses", "teachers")

	teacher := seedTeacher(t, db)
	seedCourseRaw(t, db, "PE001", "体育", 0, "体育部", teacher.ID, "zh", []string{"公共课"}, []string{"2021"})
	seedCourseRaw(t, db, "CS101", "数据结构", 3.0, "计算机学院", teacher.ID, "zh", []string{"核心课"}, []string{"2021"})

	results, total, err := repo.FindBy(ctx, course.CourseFilter{Credit: new(float32(0))})
	if err != nil {
		t.Fatalf("FindBy: %v", err)
	}
	if total != 1 {
		t.Errorf("total: got %d, want 1", total)
	}
	if len(results) != 1 || results[0].Code != "PE001" {
		t.Errorf("results: got %+v, want PE001", results)
	}
}

func TestCourseRepository_GetDetail(t *testing.T) {
	db := newTestDB(t)
	courseRepo := repository.NewCourseRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "offered_courses", "reviews", "courses", "teachers")

	teacher := seedTeacher(t, db)
	courseEntity := seedCourse(t, db, teacher.ID)
	user := seedUser(t, db)
	otherUser := seedUserRaw(t, db, "detailuser2", "detailuser2@example.com")

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
		UserID:    otherUser.ID,
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
}

func TestCourseRepository_FindOfferedCourses(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "offered_courses", "courses", "teachers")

	teacher := seedTeacher(t, db)
	courseEntity := seedCourse(t, db, teacher.ID)

	t2 := seedTeacherRaw(t, db, "T002", "李老师", "计算机学院", "副教授")

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
	if ocs[0].Semester != "2024-2025-1" {
		t.Errorf("first offered course semester: got %q, want 2024-2025-1", ocs[0].Semester)
	}
}

func TestCourseRepository_GetFilters(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "courses", "teachers")

	t1 := seedTeacher(t, db)
	t2 := seedTeacherRaw(t, db, "T002", "李老师", "数学学院", "副教授")

	seedCourseRaw(t, db, "CS101", "数据结构", 3.0, "计算机学院", t1.ID, "zh", []string{"核心课"}, []string{"2021"})
	c2 := seedCourseRaw(t, db, "CS102", "算法设计", 3.0, "计算机学院", t1.ID, "en", []string{"选修课"}, []string{"2022"})
	c3 := seedCourseRaw(t, db, "MA101", "高等数学", 4.0, "数学学院", t2.ID, "zh", []string{"核心课"}, []string{"2021", "2022"})
	seedOfferedCourseRaw(t, db, c2.ID, "2024-2025-2", "zh", nil, nil)
	seedOfferedCourseRaw(t, db, c3.ID, "2024-2025-2", "zh", nil, nil)
	seedOfferedCourseRaw(t, db, c3.ID, "2025-2026-1", "zh", nil, nil)

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
		if filters.TargetYears[0].Name != "2022" || filters.TargetYears[0].Count != 2 {
			t.Errorf("target_year 2022: got %+v", filters.TargetYears[0])
		}
		if filters.TargetYears[1].Name != "2021" || filters.TargetYears[1].Count != 2 {
			t.Errorf("target_year 2021: got %+v", filters.TargetYears[1])
		}
	})

	t.Run("semesters", func(t *testing.T) {
		if len(filters.Semesters) != 2 {
			t.Fatalf("semesters count: got %d, want 2", len(filters.Semesters))
		}
		if filters.Semesters[0].Name != "2025-2026-1" || filters.Semesters[0].Count != 1 {
			t.Errorf("semester 2025-2026-1: got %+v, want count=1", filters.Semesters[0])
		}
		if filters.Semesters[1].Name != "2024-2025-2" || filters.Semesters[1].Count != 2 {
			t.Errorf("semester 2024-2025-2: got %+v, want count=2", filters.Semesters[1])
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

func TestCourseRepository_OfferedSemesterExists(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "offered_courses", "courses", "teachers")

	teacher := seedTeacher(t, db)
	courseEntity := seedCourse(t, db, teacher.ID)
	seedOfferedCourseRaw(t, db, courseEntity.ID, "2024-2025-1", "zh", nil, nil)

	exists, err := repo.OfferedSemesterExists(ctx, "2024-2025-1")
	if err != nil {
		t.Fatalf("OfferedSemesterExists: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}

	exists, err = repo.OfferedSemesterExists(ctx, "2020-2021-1")
	if err != nil {
		t.Fatalf("OfferedSemesterExists for missing: %v", err)
	}
	if exists {
		t.Error("expected exists=false for missing semester")
	}
}
