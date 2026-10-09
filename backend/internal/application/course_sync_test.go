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

func TestCourseSyncReportsPhoneConfirmationAndSchoolLogin(t *testing.T) {
	s, c := syncFixture()
	job, err := s.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Cancel(1, job.ID)
	c.events <- course.SourceEvent{Type: "qr", QR: "data:image/png;base64,test"}
	if waitSync(t, s, job.ID, "waiting").QR == "" {
		t.Fatal("missing QR before scan")
	}
	c.events <- course.SourceEvent{Type: "scanned"}
	if waitSync(t, s, job.ID, "scanned").QR == "" {
		t.Fatal("QR removed before phone confirmation")
	}
	c.events <- course.SourceEvent{Type: "authenticating"}
	if waitSync(t, s, job.ID, "authenticating").QR != "" {
		t.Fatal("QR retained after phone confirmation")
	}
	c.events <- course.SourceEvent{Type: "discovering"}
	status := waitSync(t, s, job.ID, "discovering")
	if status.QR != "" || len(status.Semesters) != 0 {
		t.Fatal("discovery reported ready prematurely")
	}
	if _, err := s.Sync(1, job.ID, []string{testSemester}); err == nil {
		t.Fatal("import started before semester discovery")
	}
	c.events <- course.SourceEvent{Type: "ready", Semesters: []string{testSemester}}
	if len(waitSync(t, s, job.ID, "ready").Semesters) != 1 {
		t.Fatal("missing semesters")
	}
}

func TestCourseSyncClassifiesPersonalFailureAndPreservesResults(t *testing.T) {
	s, c := syncFixture()
	s.personal = true
	job, err := s.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	c.events <- course.SourceEvent{Type: "ready", Semesters: []string{testSemester, "2025-2026-2"}}
	waitSync(t, s, job.ID, "ready")
	if _, err := s.Sync(1, job.ID, []string{testSemester, "2025-2026-2"}); err != nil {
		t.Fatal(err)
	}
	c.events <- course.SourceEvent{Type: "archive", Semester: testSemester, Document: json.RawMessage(`{"semester":"2026-2027-1"}`), Report: json.RawMessage(`{}`)}
	c.events <- course.SourceEvent{Type: "collecting", Semester: "2025-2026-2"}
	c.events <- course.SourceEvent{Type: "error", Error: "PERSONAL_REQUEST_FAILED"}
	status := waitSync(t, s, job.ID, "failed")
	if len(status.Results) != 1 || !strings.Contains(status.Message, "2025-2026-2") || !strings.Contains(status.Message, "自动重试") {
		t.Fatalf("lost results or vague failure: %+v", status)
	}
	if safeSourceErrorCode("http://example.com/?credential=secret") != "SOURCE_UNAVAILABLE" {
		t.Fatal("unsafe source error exposed")
	}
}

func TestSafeSourceErrorCodePreservesLoginDiagnostics(t *testing.T) {
	for _, code := range []string{"QR_ENTRY_CHANGED", "QR_PAGE_CHANGED", "QR_CALLBACK_INVALID", "QR_IMAGE_INVALID", "URL_NOT_ALLOWED", "WECHAT_UNBOUND", "INVALID_SEMESTER"} {
		if got := safeSourceErrorCode(code); got != code {
			t.Errorf("safeSourceErrorCode(%q) = %q", code, got)
		}
	}
	for _, unsafe := range []string{"http://example.com/?credential=secret", "URL_NOT_ALLOWED: token=secret", "unknown upstream response"} {
		if got := safeSourceErrorCode(unsafe); got != "SOURCE_UNAVAILABLE" {
			t.Errorf("unrecognized diagnostic was not redacted")
		}
	}
}

func TestCourseSyncReportsRejectedLoginRedirect(t *testing.T) {
	s, c := syncFixture()
	job, err := s.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	c.events <- course.SourceEvent{Type: "authenticating"}
	waitSync(t, s, job.ID, "authenticating")
	c.events <- course.SourceEvent{Type: "error", Error: "URL_NOT_ALLOWED"}
	status := waitSync(t, s, job.ID, "failed")
	if !strings.Contains(status.Message, "跳转地址") || strings.Contains(status.Message, "是否可访问") {
		t.Fatalf("rejected redirect reported as generic connectivity failure: %q", status.Message)
	}
}

func TestCourseSyncClassifiesCollectionStructureFailure(t *testing.T) {
	s, c := syncFixture()
	job, err := s.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	c.events <- course.SourceEvent{Type: "ready", Semesters: []string{testSemester}}
	waitSync(t, s, job.ID, "ready")
	if _, err := s.Sync(1, job.ID, []string{testSemester}); err != nil {
		t.Fatal(err)
	}
	c.events <- course.SourceEvent{Type: "collecting", Semester: testSemester}
	c.events <- course.SourceEvent{Type: "error", Error: "PAGE_STRUCTURE"}
	status := waitSync(t, s, job.ID, "failed")
	if !strings.Contains(status.Message, "完整性校验") || strings.Contains(status.Message, "学期列表") || len(status.Results) != 0 {
		t.Fatalf("wrong collection failure: %+v", status)
	}
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
