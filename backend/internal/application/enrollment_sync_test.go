//go:build test

package application

import (
	"context"
	"encoding/json"
	"jcourse/internal/domain/course"
	"strings"
	"testing"
	"time"
)

func TestEnrollmentSyncBoundsUnmatchedReportAndHistory(t *testing.T) {
	c := &testSourceConnection{events: make(chan course.SourceEvent, 10)}
	items := make([]course.EnrollmentSourceCourse, 1000)
	for i := range items {
		items[i] = course.EnrollmentSourceCourse{LessonNo: "L1", Name: strings.Repeat("课", 300)}
	}
	s := NewEnrollmentSyncService(&course.MockCourseSource{ConnectFn: func(context.Context) (course.SourceConnection, error) { return c, nil }}, &course.MockEnrollmentImporter{ImportSemesterFn: func(context.Context, int, string, []course.EnrollmentSourceCourse) (*course.EnrollmentImportResult, error) {
		return &course.EnrollmentImportResult{Unmatched: items}, nil
	}})
	s.limit = 1
	job, err := s.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	c.events <- course.SourceEvent{Type: "ready", Semesters: []string{testSemester}}
	waitSync(t, s.jobs[1], job.ID, "ready")
	s.Sync(1, job.ID, []string{testSemester})
	doc, _ := json.Marshal(map[string]any{"semester": testSemester, "courses": []course.EnrollmentSourceCourse{{LessonNo: "L1", Name: "课程"}}})
	c.events <- course.SourceEvent{Type: "archive", Semester: testSemester, Document: doc}
	c.events <- course.SourceEvent{Type: "done"}
	status := waitSync(t, s.jobs[1], job.ID, "done")
	result := status.Results[0].Enrollment
	if result.UnmatchedCount != 1000 || len(result.Unmatched) != 20 || len([]rune(result.Unmatched[0].Name)) != 129 {
		t.Fatal("report unbounded or counts lost")
	}
	second, err := s.Start(2)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Cancel(2, second.ID)
	if len(s.jobs) != 1 || s.jobs[1] != nil {
		t.Fatal("completed history grew past capacity")
	}
}

func TestEnrollmentSyncIsolationImportAndCapacity(t *testing.T) {
	connections := []*testSourceConnection{}
	source := &course.MockCourseSource{ConnectFn: func(context.Context) (course.SourceConnection, error) {
		c := &testSourceConnection{events: make(chan course.SourceEvent, 10)}
		connections = append(connections, c)
		return c, nil
	}}
	imported := make(chan int, 1)
	importer := &course.MockEnrollmentImporter{ImportSemesterFn: func(_ context.Context, owner int, semester string, items []course.EnrollmentSourceCourse) (*course.EnrollmentImportResult, error) {
		if semester != testSemester || len(items) != 1 {
			t.Errorf("wrong import: %s %+v", semester, items)
		}
		imported <- owner
		return &course.EnrollmentImportResult{Matched: 1, Added: 1, Unmatched: []course.EnrollmentSourceCourse{}}, nil
	}}
	s := NewEnrollmentSyncService(source, importer)
	s.limit = 2
	first, err := s.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Cancel(1, first.ID)
	second, err := s.Start(2)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Cancel(2, second.ID)
	if _, err = s.Start(1); err == nil {
		t.Fatal("duplicate user accepted")
	}
	if _, err = s.Start(3); err == nil {
		t.Fatal("capacity exceeded")
	}
	for _, op := range []func() error{
		func() error { _, e := s.Status(2, first.ID); return e },
		func() error { _, e := s.Sync(2, first.ID, []string{testSemester}); return e },
		func() error { _, e := s.Cancel(2, first.ID); return e },
	} {
		if op() == nil {
			t.Fatal("cross-user access allowed")
		}
	}
	connections[0].events <- course.SourceEvent{Type: "ready", Semesters: []string{testSemester}}
	waitSync(t, s.jobs[1], first.ID, "ready")
	if _, err = s.Sync(1, first.ID, []string{testSemester}); err != nil {
		t.Fatal(err)
	}
	doc, _ := json.Marshal(map[string]any{"semester": testSemester, "courses": []course.EnrollmentSourceCourse{{LessonNo: "C.01", Name: "课程"}}})
	connections[0].events <- course.SourceEvent{Type: "archive", Semester: testSemester, Document: doc}
	connections[0].events <- course.SourceEvent{Type: "done"}
	status := waitSync(t, s.jobs[1], first.ID, "done")
	if status.Results[0].Enrollment.Added != 1 {
		t.Fatal("missing enrollment result")
	}
	select {
	case owner := <-imported:
		if owner != 1 {
			t.Fatal("wrong owner")
		}
	case <-time.After(time.Second):
		t.Fatal("not imported")
	}
	if _, err = s.Start(3); err != nil {
		t.Fatal("completed job did not release capacity")
	}
	s.Cancel(3, s.jobs[3].session.status.ID)
}

func TestEnrollmentSyncRejectsIncompleteSemester(t *testing.T) {
	c := &testSourceConnection{events: make(chan course.SourceEvent, 10)}
	s := NewEnrollmentSyncService(&course.MockCourseSource{ConnectFn: func(context.Context) (course.SourceConnection, error) { return c, nil }}, &course.MockEnrollmentImporter{ImportSemesterFn: func(context.Context, int, string, []course.EnrollmentSourceCourse) (*course.EnrollmentImportResult, error) {
		t.Fatal("imported incomplete document")
		return nil, nil
	}})
	job, err := s.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Cancel(1, job.ID)
	c.events <- course.SourceEvent{Type: "ready", Semesters: []string{testSemester}}
	waitSync(t, s.jobs[1], job.ID, "ready")
	s.Sync(1, job.ID, []string{testSemester})
	c.events <- course.SourceEvent{Type: "archive", Semester: testSemester, Document: json.RawMessage(`{"semester":"` + testSemester + `"}`)}
	waitSync(t, s.jobs[1], job.ID, "failed")
}
