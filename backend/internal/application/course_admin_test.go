package application

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"jcourse/internal/domain/course"
)

const testSemester = "2026-2027-1"

func courseArchive(t *testing.T, lessons []any, mutateDocument func(map[string]any), mutateReport func(map[string]any)) []byte {
	t.Helper()
	semester := map[string]any{
		"schema_version": 1,
		"source":         map[string]any{"host": "bkjw.chd.edu.cn"},
		"semester":       testSemester,
		"expected_count": len(lessons),
		"lessons":        lessons,
	}
	report := map[string]any{
		"schema_version":       1,
		"semester":             testSemester,
		"completed":            true,
		"list_complete":        true,
		"expected_count":       len(lessons),
		"lesson_count":         len(lessons),
		"detail_success_count": len(lessons),
		"missing_detail_count": 0,
		"duplicate_lesson_ids": []string{},
		"failed_lesson_ids":    []string{},
		"source_page":          map[string]any{"host": "bkjw.chd.edu.cn"},
	}
	if mutateDocument != nil {
		mutateDocument(semester)
	}
	if mutateReport != nil {
		mutateReport(report)
	}

	var buffer bytes.Buffer
	w := zip.NewWriter(&buffer)
	write := func(name string, value any) {
		t.Helper()
		file, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.NewEncoder(file).Encode(value); err != nil {
			t.Fatal(err)
		}
	}
	write("collector/semester.json", semester)
	write("collector/report.json", report)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func testLesson(id, code, name string, teachers []string) map[string]any {
	return map[string]any{
		"id": id,
		"summary": map[string]any{
			"lesson_no": "CS101.01", "course_code": code, "course_name": name,
			"course_type": "专业课", "teach_class": "一班", "teachers": teachers,
			"actual_count": 30, "limit_count": 40, "credit": 3, "course_period": "48/16", "week_range": "1-16",
			"raw": map[string]any{"课程名称": "ignored"},
		},
		"detail": map[string]any{
			"basic": map[string]any{
				"semester": "2026-2027学年1学期", "course_code": code, "course_name": name,
				"credit": 3, "course_type": "专业课", "department": "计算机学院", "campus": "本部",
				"language": "中文", "assessment": "考试", "teachers": teachers,
			},
			"arrangement": map[string]any{
				"total_hours": 48, "weekly_hours": 3, "week_range": "1-16", "start_week": 1, "end_week": 16, "week_count": 16,
			},
			"restrictions": []any{map[string]any{"raw": "ignored"}},
			"remark":       "课程备注",
		},
		"links": map[string]any{"syllabus": "/eams/syllabus.action?lesson.id=" + id},
	}
}

func TestParseCourseArchiveGroupsSectionsAndAllowsTeacherlessLessons(t *testing.T) {
	archive := courseArchive(t, []any{
		testLesson("1", "CS101", "程序设计", []string{"  张三　"}),
		testLesson("2", "CS101", "程序设计", []string{"张三"}),
		testLesson("3", "CS102", "数据结构", nil),
	}, nil, nil)

	semester, offerings, err := parseCourseArchive(archive)
	if err != nil {
		t.Fatalf("parseCourseArchive: %v", err)
	}
	if semester != testSemester || len(offerings) != 2 {
		t.Fatalf("got semester %q and %d offerings, want %q and 2", semester, len(offerings), testSemester)
	}
	if offerings[0].Code != "CS101" || len(offerings[0].Sections) != 2 {
		t.Fatalf("same-code sections were not grouped: %#v", offerings[0])
	}
	if offerings[0].Sections[0].Teachers[0] != "张三" {
		t.Fatalf("teacher name was not normalized: %#v", offerings[0].Sections[0].Teachers)
	}
	if offerings[0].Sections[0].SyllabusURL != "http://bkjw.chd.edu.cn/eams/syllabus.action?lesson.id=1" {
		t.Fatalf("syllabus URL was not canonicalized: %q", offerings[0].Sections[0].SyllabusURL)
	}
	teacherless := offerings[1]
	if teacherless.Code != "CS102" || len(teacherless.Sections) != 1 || len(teacherless.Sections[0].Teachers) != 0 {
		t.Fatalf("teacherless lesson should remain importable: %#v", teacherless)
	}
}

func TestParseCourseArchiveRejectsIncompleteOrInconsistentData(t *testing.T) {
	baseLesson := testLesson("1", "CS101", "程序设计", []string{"张三"})
	tests := []struct {
		name           string
		lessons        []any
		mutateDocument func(map[string]any)
		mutateReport   func(map[string]any)
	}{
		{name: "incomplete report", lessons: []any{baseLesson}, mutateReport: func(report map[string]any) { report["completed"] = false }},
		{name: "semester mismatch", lessons: []any{baseLesson}, mutateReport: func(report map[string]any) { report["semester"] = "2025-2026-2" }},
		{name: "count mismatch", lessons: []any{baseLesson}, mutateReport: func(report map[string]any) { report["expected_count"] = 2 }},
		{name: "unsupported format", lessons: []any{baseLesson}, mutateDocument: func(document map[string]any) { document["schema_version"] = 2 }},
		{name: "foreign source", lessons: []any{baseLesson}, mutateDocument: func(document map[string]any) { document["source"] = map[string]any{"host": "example.com"} }},
		{name: "duplicate lesson id", lessons: []any{baseLesson, testLesson("1", "CS102", "数据结构", nil)}},
		{name: "missing lesson detail", lessons: []any{testLesson("1", "CS101", "程序设计", []string{"张三"})}, mutateDocument: func(document map[string]any) {
			lessons := document["lessons"].([]any)
			lesson := lessons[0].(map[string]any)
			lesson["detail"] = nil
		}},
		{name: "detail semester mismatch", lessons: []any{testLesson("1", "CS101", "程序设计", []string{"张三"})}, mutateDocument: func(document map[string]any) {
			lessons := document["lessons"].([]any)
			lesson := lessons[0].(map[string]any)
			detail := lesson["detail"].(map[string]any)
			basic := detail["basic"].(map[string]any)
			basic["semester"] = "2025-2026学年2学期"
		}},
		{name: "foreign syllabus", lessons: []any{testLesson("1", "CS101", "程序设计", []string{"张三"})}, mutateDocument: func(document map[string]any) {
			lessons := document["lessons"].([]any)
			lesson := lessons[0].(map[string]any)
			links := lesson["links"].(map[string]any)
			links["syllabus"] = "https://example.com/eams/syllabus"
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive := courseArchive(t, tt.lessons, tt.mutateDocument, tt.mutateReport)
			if _, _, err := parseCourseArchive(archive); err == nil {
				t.Fatal("parseCourseArchive unexpectedly accepted invalid data")
			}
		})
	}
}

type courseDataRepositoryStub struct {
	offerings    map[string][]course.ManagedOffering
	replaceCalls int
}

func (r *courseDataRepositoryStub) ListCourseSemesters(context.Context) ([]string, error) {
	return nil, nil
}

func (r *courseDataRepositoryStub) ListManagedOfferings(_ context.Context, filter course.ManagedOfferingFilter) ([]course.ManagedOffering, int64, error) {
	items := r.offerings[filter.Semester]
	return items, int64(len(items)), nil
}

func (r *courseDataRepositoryStub) ReplaceSemester(_ context.Context, semester string, offerings []course.ManagedOffering) error {
	r.replaceCalls++
	r.offerings[semester] = offerings
	return nil
}

func (r *courseDataRepositoryStub) SaveManagedOffering(_ context.Context, _ int, offering course.ManagedOffering) (*course.ManagedOffering, error) {
	return &offering, nil
}

func (r *courseDataRepositoryStub) DeleteManagedOffering(context.Context, int) error { return nil }

func TestImportArchiveReplacesTheWholeSemester(t *testing.T) {
	repo := &courseDataRepositoryStub{offerings: map[string][]course.ManagedOffering{}}
	service := NewCourseManagementService(repo)
	archive := courseArchive(t, []any{
		testLesson("1", "CS101", "程序设计", []string{"张三"}),
		testLesson("2", "CS101", "程序设计", []string{"张三"}),
	}, nil, nil)

	first, err := service.ImportArchive(context.Background(), archive)
	if err != nil {
		t.Fatalf("first ImportArchive: %v", err)
	}
	second, err := service.ImportArchive(context.Background(), archive)
	if err != nil {
		t.Fatalf("second ImportArchive: %v", err)
	}
	if repo.replaceCalls != 2 || first.ReplacedCount != 0 || second.ReplacedCount != 1 {
		t.Fatalf("replacement state: calls=%d first=%d second=%d", repo.replaceCalls, first.ReplacedCount, second.ReplacedCount)
	}
	if got := len(repo.offerings[testSemester]); got != 1 || len(repo.offerings[testSemester][0].Sections) != 2 {
		t.Fatalf("semester was not replaced with grouped offering data: %#v", repo.offerings[testSemester])
	}
}

func TestAggregateCourseRatingsAcrossSameCodeCourses(t *testing.T) {
	got := aggregateCourseRatings([]course.RatingInfo{
		{Count: 2, Avg: 4, Score: 3.5, Distribution: [5]int{0, 0, 0, 1, 1}},
		{Count: 1, Avg: 2, Score: 2, Distribution: [5]int{0, 1, 0, 0, 0}},
	})
	if got.Count != 3 || got.Avg != 10.0/3 || got.Score != 3.0 || got.Distribution != [5]int{0, 1, 0, 1, 1} {
		t.Fatalf("unexpected course-code rating summary: %#v", got)
	}
}
