//go:build test

package application

import (
	"context"
	"encoding/json"
	"errors"
	"jcourse/internal/domain/course"
	"strings"
	"sync"
	"testing"
	"time"
)

type testSourceConnection struct {
	events   chan course.SourceEvent
	selected []string
	mu       sync.Mutex
}

func (c *testSourceConnection) Events() <-chan course.SourceEvent { return c.events }
func (c *testSourceConnection) Sync(items []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.selected = append([]string{}, items...)
	return nil
}
func (c *testSourceConnection) Close() {}
func waitSync(t *testing.T, s *CourseSyncService, id, state string) *CourseSyncStatus {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		status, err := s.Status(1, id)
		if err != nil {
			t.Fatal(err)
		}
		if status.State == state {
			return status
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("did not reach %s", state)
	return nil
}
func syncFixture() (*CourseSyncService, *testSourceConnection) {
	c := &testSourceConnection{events: make(chan course.SourceEvent, 10)}
	s := &CourseSyncService{source: &course.MockCourseSource{ConnectFn: func(context.Context) (course.SourceConnection, error) { return c, nil }}, importArchive: func(context.Context, []byte) (*CourseImportResult, error) {
		return &CourseImportResult{Semester: testSemester, CourseCount: 1}, nil
	}}
	return s, c
}

func TestCourseSyncDetailedProgressAndImportStage(t *testing.T) {
	s, c := syncFixture()
	importStarted := make(chan struct{})
	release := make(chan struct{})
	s.importArchive = func(context.Context, []byte) (*CourseImportResult, error) {
		close(importStarted)
		<-release
		return &CourseImportResult{Semester: testSemester, CourseCount: 1}, nil
	}
	job, err := s.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Cancel(1, job.ID)
	c.events <- course.SourceEvent{Type: "ready", Semesters: []string{testSemester}}
	waitSync(t, s, job.ID, "ready")
	if _, err := s.Sync(1, job.ID, []string{testSemester}); err != nil {
		t.Fatal(err)
	}
	total := 100
	c.events <- course.SourceEvent{Type: "progress", Semester: testSemester, Progress: &course.SourceProgress{Stage: "details", Completed: 21, Total: &total, Failed: 2, Retries: 3}}
	deadline := time.Now().Add(time.Second)
	for {
		status, _ := s.Status(1, job.ID)
		if status.Progress != nil {
			if status.Progress.Completed != 21 || status.Progress.Failed != 2 || status.Progress.Retries != 3 || status.Progress.UpdatedAt.IsZero() || status.StartedAt == nil {
				t.Fatalf("missing progress: %+v", status)
			}
			*status.Progress.Total = 999
			fresh, _ := s.Status(1, job.ID)
			if *fresh.Progress.Total != 100 {
				t.Fatal("snapshot modified live progress")
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("progress not consumed")
		}
		time.Sleep(time.Millisecond)
	}
	c.events <- course.SourceEvent{Type: "archive", Semester: testSemester, Document: json.RawMessage(`{"semester":"2026-2027-1"}`), Report: json.RawMessage(`{}`)}
	select {
	case <-importStarted:
	case <-time.After(time.Second):
		t.Fatal("import not started")
	}
	status, _ := s.Status(1, job.ID)
	if status.Progress.Stage != "importing" || len(status.Results) != 0 {
		t.Fatal("save reported complete before import finished")
	}
	close(release)
	c.events <- course.SourceEvent{Type: "done"}
	finished := waitSync(t, s, job.ID, "done")
	if finished.Progress.Stage != "completed" || len(finished.Results) != 1 {
		t.Fatal("missing saved semester progress")
	}
}

func TestCourseSyncLoginErrorsClearQRAndAllowReconnect(t *testing.T) {
	for _, code := range []string{"QR_LOGIN_FAILED", "AUTH_REQUIRED", "PAGE_STRUCTURE", "SEMESTER_DISCOVERY_FAILED", "PERMISSION_DENIED", "https://example.test/?code=secret"} {
		t.Run(code, func(t *testing.T) {
			s, c := syncFixture()
			status, err := s.Start(1)
			if err != nil {
				t.Fatal(err)
			}
			c.events <- course.SourceEvent{Type: "qr", QR: "data:image/png;base64,test"}
			c.events <- course.SourceEvent{Type: "error", Error: code}
			failed := waitSync(t, s, status.ID, "failed")
			if failed.QR != "" || failed.Message == "" || strings.Contains(failed.Message, "secret") {
				t.Fatalf("unsafe failure status: %+v", failed)
			}
			retry, err := s.Start(1)
			if err != nil {
				t.Fatalf("cannot reconnect after login failure: %v", err)
			}
			defer s.Cancel(1, retry.ID)
		})
	}
}
func TestCourseSyncOwnershipAndSelection(t *testing.T) {
	s, c := syncFixture()
	status, err := s.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Cancel(1, status.ID)
	if _, err := s.Status(2, status.ID); err == nil {
		t.Fatal("other administrator saw job")
	}
	if _, err := s.Start(2); err == nil {
		t.Fatal("parallel source session accepted")
	}
	c.events <- course.SourceEvent{Type: "ready", Semesters: []string{testSemester}}
	waitSync(t, s, status.ID, "ready")
	if _, err := s.Sync(1, status.ID, []string{"2000-2001-1"}); err == nil {
		t.Fatal("unknown semester accepted")
	}
	if _, err := s.Sync(1, status.ID, []string{testSemester, testSemester}); err == nil {
		t.Fatal("duplicate semester accepted")
	}
	if _, err := s.Sync(1, status.ID, []string{testSemester}); err != nil {
		t.Fatal(err)
	}
	c.events <- course.SourceEvent{Type: "archive", Semester: testSemester, Document: json.RawMessage(`{"semester":"2026-2027-1"}`), Report: json.RawMessage(`{}`)}
	c.events <- course.SourceEvent{Type: "done"}
	if result := waitSync(t, s, status.ID, "done"); len(result.Results) != 1 {
		t.Fatal("missing import result")
	}
}
func TestCourseSyncStopsOnImportFailure(t *testing.T) {
	s, c := syncFixture()
	s.importArchive = func(context.Context, []byte) (*CourseImportResult, error) { return nil, errors.New("invalid archive") }
	status, _ := s.Start(1)
	c.events <- course.SourceEvent{Type: "ready", Semesters: []string{testSemester}}
	waitSync(t, s, status.ID, "ready")
	_, _ = s.Sync(1, status.ID, []string{testSemester})
	c.events <- course.SourceEvent{Type: "archive", Semester: testSemester, Document: json.RawMessage(`{"semester":"2026-2027-1"}`)}
	if result := waitSync(t, s, status.ID, "failed"); len(result.Results) != 0 {
		t.Fatal("failed import was reported successful")
	}
}
func TestSourceArchiveUsesExistingCompletenessValidator(t *testing.T) {
	archive, err := sourceArchive(course.SourceEvent{Semester: testSemester, Document: json.RawMessage(`{"semester":"2026-2027-1"}`), Report: json.RawMessage(`{"completed":false}`)})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := parseCourseArchive(archive); err == nil {
		t.Fatal("incomplete collector result accepted")
	}
}
