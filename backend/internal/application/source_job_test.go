//go:build test

package application

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"jcourse/internal/domain/course"
)

func TestPersistedJobSurvivesRestartWithoutLoginOrPersonalDetails(t *testing.T) {
	var mu sync.Mutex
	values := map[int]json.RawMessage{}
	store := &course.MockSourceJobStore{
		SaveFn: func(_ context.Context, owner int, personal bool, value json.RawMessage) error {
			mu.Lock()
			defer mu.Unlock()
			values[owner] = append(json.RawMessage{}, value...)
			return nil
		},
		LoadFn: func(_ context.Context, owner int, personal bool) (json.RawMessage, error) {
			mu.Lock()
			defer mu.Unlock()
			return values[owner], nil
		},
	}
	s := (&CourseSyncService{personal: true}).WithJobStore(store)
	session := &courseSyncSession{owner: 1, status: CourseSyncStatus{ID: "restart-job", State: "running", QR: "private-qr", Selected: []string{testSemester}, Results: []CourseImportResult{{Semester: testSemester, Enrollment: &course.EnrollmentImportResult{Added: 2, Unmatched: []course.EnrollmentSourceCourse{{Name: "private-course"}}, UnmatchedCount: 1}}}}}
	s.persist(session, true)
	if strings.Contains(string(values[1]), "private-") {
		t.Fatal("QR or personal details persisted")
	}
	restarted := NewEnrollmentSyncService(nil, nil).WithJobStore(store)
	status, err := restarted.Status(1, "restart-job")
	if err != nil || status.State != "failed" || status.QR != "" || len(status.Results) != 1 || !strings.Contains(status.Message, "重新扫码") {
		t.Fatal(status, err)
	}
	if _, err := restarted.Status(2, "restart-job"); err == nil {
		t.Fatal("another owner read recovery data")
	}
	if _, err := restarted.Status(1, "another-job"); err == nil {
		t.Fatal("different ID accepted")
	}
	if selected := s.resumeSelection(1); len(selected) != 1 {
		t.Fatal(selected)
	}
	// Starting again retains selection but not old results or login credentials.
	c := &testSourceConnection{events: make(chan course.SourceEvent, 10)}
	s.source = &course.MockCourseSource{ConnectFn: func(context.Context) (course.SourceConnection, error) { return c, nil }}
	job, err := s.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Cancel(1, job.ID)
	if len(job.Selected) != 1 || len(job.Results) != 0 || job.ID == "restart-job" {
		t.Fatal(job)
	}
}

func TestCompletedJobDoesNotAutomaticallySelectPreviousSemesters(t *testing.T) {
	data, _ := json.Marshal(CourseSyncStatus{ID: "completed", State: "done", Selected: []string{testSemester}})
	s := (&CourseSyncService{}).WithJobStore(&course.MockSourceJobStore{LoadFn: func(context.Context, int, bool) (json.RawMessage, error) { return data, nil }})
	if len(s.resumeSelection(1)) != 0 {
		t.Fatal("completed selection reused")
	}
	status, err := s.Status(1, "completed")
	if err != nil || status.State != "done" {
		t.Fatal(status, err)
	}
}
