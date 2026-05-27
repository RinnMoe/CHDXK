package repository_test

import (
	"context"
	"testing"

	"jcourse/internal/domain/teacher"
	"jcourse/internal/infrastructure/repository"
)

func TestTeacherRepository_FindBy(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewTeacherRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "teachers")

	seedTeacherRaw(t, db, "T001", "张三", "计算机学院", "教授")
	seedTeacherRaw(t, db, "T002", "李四", "数学学院", "副教授")
	seedTeacherRaw(t, db, "T003", "王五", "计算机学院", "讲师")

	t.Run("list all", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, teacher.TeacherFilter{})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 3 {
			t.Errorf("total: got %d, want 3", total)
		}
		if len(results) != 3 {
			t.Errorf("results count: got %d, want 3", len(results))
		}
	})

	t.Run("filter by department", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, teacher.TeacherFilter{Department: "计算机学院"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 2 {
			t.Errorf("total: got %d, want 2", total)
		}
		for _, r := range results {
			if r.Department != "计算机学院" {
				t.Errorf("unexpected department: %q", r.Department)
			}
		}
	})

	t.Run("filter by title", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, teacher.TeacherFilter{Title: "教授"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 1 {
			t.Errorf("total: got %d, want 1", total)
		}
		if results[0].Title != "教授" {
			t.Errorf("title: got %q, want 教授", results[0].Title)
		}
	})

	t.Run("search by q compact pinyin", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, teacher.TeacherFilter{Q: "zhangsan"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 1 {
			t.Errorf("total: got %d, want 1", total)
		}
		if results[0].Name != "张三" {
			t.Errorf("Name: got %q, want 张三", results[0].Name)
		}
	})

	t.Run("search by q spaced pinyin", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, teacher.TeacherFilter{Q: "zhang san"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 1 {
			t.Errorf("total: got %d, want 1", total)
		}
		if len(results) == 0 || results[0].Name != "张三" {
			t.Errorf("Name: got %v, want 张三", results)
		}
	})

	t.Run("search by q pinyin abbr", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, teacher.TeacherFilter{Q: "zs"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 1 {
			t.Errorf("total: got %d, want 1", total)
		}
		if len(results) == 0 || results[0].Name != "张三" {
			t.Errorf("Name: got %v, want 张三", results)
		}
	})

	t.Run("search by q code", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, teacher.TeacherFilter{Q: "T002"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 1 {
			t.Errorf("total: got %d, want 1", total)
		}
		if len(results) == 0 || results[0].Name != "李四" {
			t.Errorf("Name: got %v, want 李四", results)
		}
	})

	t.Run("search by q name substring", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, teacher.TeacherFilter{Q: "李"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 1 {
			t.Errorf("total: got %d, want 1", total)
		}
		if len(results) == 0 || results[0].Name != "李四" {
			t.Errorf("Name: got %v, want 李四", results)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, teacher.TeacherFilter{Page: 1, PageSize: 2})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 3 {
			t.Errorf("total: got %d, want 3", total)
		}
		if len(results) != 2 {
			t.Errorf("page 1 results: got %d, want 2", len(results))
		}

		results2, _, err := repo.FindBy(ctx, teacher.TeacherFilter{Page: 2, PageSize: 2})
		if err != nil {
			t.Fatalf("FindBy page 2: %v", err)
		}
		if len(results2) != 1 {
			t.Errorf("page 2 results: got %d, want 1", len(results2))
		}
	})

	t.Run("no match filter", func(t *testing.T) {
		results, total, err := repo.FindBy(ctx, teacher.TeacherFilter{Department: "不存在的学院"})
		if err != nil {
			t.Fatalf("FindBy: %v", err)
		}
		if total != 0 {
			t.Errorf("total: got %d, want 0", total)
		}
		if len(results) != 0 {
			t.Errorf("results: got %d, want 0", len(results))
		}
	})
}

func TestTeacherRepository_FindBy_DefaultSortByCode(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewTeacherRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "teachers")

	seedTeacherRaw(t, db, "T300", "后排序老师", "测试学院", "讲师")
	seedTeacherRaw(t, db, "T100", "先排序老师", "测试学院", "讲师")
	seedTeacherRaw(t, db, "T200", "中间排序老师", "测试学院", "讲师")

	results, total, err := repo.FindBy(ctx, teacher.TeacherFilter{})
	if err != nil {
		t.Fatalf("FindBy: %v", err)
	}
	if total != 3 {
		t.Fatalf("total: got %d, want 3", total)
	}
	if got := []string{results[0].Code, results[1].Code, results[2].Code}; got[0] != "T100" || got[1] != "T200" || got[2] != "T300" {
		t.Fatalf("codes: got %v, want [T100 T200 T300]", got)
	}
}

func TestTeacherRepository_GetFilters(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewTeacherRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "teachers")

	seedTeacherRaw(t, db, "T001", "张三", "计算机学院", "教授")
	seedTeacherRaw(t, db, "T002", "李四", "数学学院", "副教授")
	seedTeacherRaw(t, db, "T003", "王五", "计算机学院", "讲师")

	filters, err := repo.GetFilters(ctx)
	if err != nil {
		t.Fatalf("GetFilters: %v", err)
	}

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

	t.Run("titles", func(t *testing.T) {
		if len(filters.Titles) != 3 {
			t.Fatalf("titles count: got %d, want 3", len(filters.Titles))
		}
		if filters.Titles[0].Name != "副教授" || filters.Titles[0].Count != 1 {
			t.Errorf("title 副教授: got %+v", filters.Titles[0])
		}
		if filters.Titles[1].Name != "教授" || filters.Titles[1].Count != 1 {
			t.Errorf("title 教授: got %+v", filters.Titles[1])
		}
		if filters.Titles[2].Name != "讲师" || filters.Titles[2].Count != 1 {
			t.Errorf("title 讲师: got %+v", filters.Titles[2])
		}
	})
}
