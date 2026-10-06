//go:build test

package repository_test

import (
	"context"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/teacher"
	"jcourse/internal/infrastructure/repository"
	"testing"
	"time"
)

func TestTeacherProfileImportAndManualBinding(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := repository.NewTeacherRepository(db)
	a := seedTeacherRaw(t, db, "P1", "张三", "理学院", "")
	b := seedTeacherRaw(t, db, "P2", "张三", "信息工程学院", "")
	c := seedTeacherRaw(t, db, "P3", "李四", "理学院", "")
	profiles := []teacher.Profile{{URL: "https://js.chd.edu.cn/a/list.htm", Name: "张三", Department: "理学院", Title: "教授", Fields: map[string]string{"研究领域": "数学"}, UpdatedAt: time.Now()}, {URL: "https://js.chd.edu.cn/b/list.htm", Name: "张三", Department: "信息工程学院", Title: "讲师", Fields: map[string]string{}}}
	teachers, err := repo.Teachers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, teachers, profiles); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByID(ctx, a.ID)
	if err != nil || got.Profile == nil || got.Title != "教授" {
		t.Fatal(got, err)
	}
	missing, _ := repo.GetByID(ctx, c.ID)
	if missing.Profile != nil {
		t.Fatal("missing profile invented")
	}
	filtered, total, err := repo.FindBy(ctx, teacher.TeacherFilter{Title: "教授"})
	if err != nil || total != 1 || len(filtered) != 1 {
		t.Fatal(filtered, total, err)
	}
	profiles = append(profiles, teacher.Profile{URL: "https://js.chd.edu.cn/c/list.htm", Name: "张三", Department: "理学院", Title: "副教授", Fields: map[string]string{}})
	if err := repo.Save(ctx, teachers, profiles); err != nil {
		t.Fatal(err)
	}
	matches, _ := repo.Matches(ctx)
	if len(matches) != 1 || matches[0].TeacherID != a.ID {
		t.Fatal(matches)
	}
	if err := repo.Bind(ctx, a.ID, profiles[1].URL); err == nil {
		t.Fatal("different college bound")
	}
	if err := repo.Bind(ctx, a.ID, profiles[0].URL); err != nil {
		t.Fatal(err)
	}
	// Temporary detail failure must retain prior professional fields and manual choice.
	profiles[0].Fields = map[string]string{}
	profiles[0].UpdatedAt = time.Time{}
	if err := repo.Save(ctx, teachers, profiles); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.GetByID(ctx, a.ID)
	if got.Profile == nil || got.Profile.Fields["研究领域"] != "数学" {
		t.Fatal(got)
	}
	matches, _ = repo.Matches(ctx)
	if matches[0].Status != "manual" {
		t.Fatal("manual decision overwritten")
	}
	_ = b
}

func TestTeacherProfileManualDataSurvivesSync(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := repository.NewTeacherRepository(db)
	a := seedTeacherRaw(t, db, course.InternalTeacherCode("理学院", "原姓名"), "原姓名", "理学院", "")
	manual := teacher.Profile{Name: "新姓名", Department: "理学院", Title: "", Fields: map[string]string{"联系方式": "office@example.org", "办公地点": "教学楼"}}
	if err := repo.Edit(ctx, a.ID, manual); err != nil {
		t.Fatal(err)
	}
	profiles := []teacher.Profile{{Name: "新姓名", Department: "理学院", URL: "https://js.chd.edu.cn/test/list.htm", Title: "教授", Fields: map[string]string{"办公地点": "其他地点"}, UpdatedAt: time.Now()}}
	rows, _ := repo.Teachers(ctx)
	if err := repo.Save(ctx, rows, profiles); err != nil {
		t.Fatal(err)
	}
	got, _ := repo.GetByID(ctx, a.ID)
	if got.Name != "新姓名" || got.Profile == nil || got.Profile.URL != "" || got.Title != "" || got.Profile.Fields["办公地点"] != "教学楼" {
		t.Fatalf("manual data overwritten: %#v", got)
	}
	courseRepo := repository.NewCourseRepository(db)
	if err := courseRepo.ReplaceSemester(ctx, "2026-2027-1", []course.ManagedOffering{{Code: "KEEP1", Semester: "2026-2027-1", Sections: []course.CourseSection{{SourceLessonID: "keep-section", CourseName: "保留资料课程", Department: "理学院", Teachers: []string{"原姓名"}}}}}); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.GetByID(ctx, a.ID)
	if got.Name != "新姓名" || got.Profile.Fields["办公地点"] != "教学楼" {
		t.Fatal("course import overwrote manual teacher data")
	}
	rows, total, err := repo.List(ctx, teacher.TeacherFilter{Q: "新姓名", Page: 1, PageSize: 20})
	if err != nil || total != 1 || len(rows) != 1 {
		t.Fatal(rows, total, err)
	}
	if err := repo.Bind(ctx, a.ID, profiles[0].URL); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.GetByID(ctx, a.ID)
	if got.Profile.Fields["联系方式"] != "office@example.org" {
		t.Fatal("manual data overwritten on binding")
	}
}
