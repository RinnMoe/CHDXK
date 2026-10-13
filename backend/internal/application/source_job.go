package application

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"jcourse/internal/domain/course"
	"jcourse/pkg/apperr"
)

func (s *CourseSyncService) WithJobStore(store course.SourceJobStore) *CourseSyncService {
	s.jobStore = store
	return s
}
func (s *EnrollmentSyncService) WithJobStore(store course.SourceJobStore) *EnrollmentSyncService {
	s.jobStore = store
	return s
}

func (s *CourseSyncService) persist(session *courseSyncSession, force bool) {
	if s.jobStore == nil {
		return
	}
	session.persistMu.Lock()
	defer session.persistMu.Unlock()
	status := session.snapshot()
	if !force && status.State == session.savedState && time.Since(session.lastSaved) < time.Second {
		return
	}
	status.QR = ""
	status.Warning = ""
	status.QueuePosition = 0
	// Recovery needs counters and completed semesters, never course names or QR credentials.
	for i, result := range status.Results {
		if result.Enrollment != nil {
			v := *result.Enrollment
			v.Unmatched = []course.EnrollmentSourceCourse{}
			status.Results[i].Enrollment = &v
		}
	}
	data, err := json.Marshal(status)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err == nil {
		err = s.jobStore.Save(ctx, session.owner, s.personal, data)
	}
	if err != nil {
		slog.Warn("collector progress persistence failed", "job_id", status.ID)
		session.mu.Lock()
		session.status.Warning = "任务进度无法保存，服务重启后需重新发起；请检查采集缓存目录"
		session.mu.Unlock()
		return
	}
	session.mu.Lock()
	session.status.Warning = ""
	session.mu.Unlock()
	session.lastSaved = time.Now()
	session.savedState = status.State
}

func (s *CourseSyncService) loadStatus(owner int) *CourseSyncStatus {
	if s.jobStore == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data, err := s.jobStore.Load(ctx, owner, s.personal)
	var status CourseSyncStatus
	if err != nil || len(data) == 0 || json.Unmarshal(data, &status) != nil || status.ID == "" {
		return nil
	}
	status.QR = ""
	status.QueuePosition = 0
	for _, result := range status.Results {
		if result.Enrollment != nil && result.Enrollment.Unmatched == nil {
			result.Enrollment.Unmatched = []course.EnrollmentSourceCourse{}
		}
	}
	return &status
}

func (s *CourseSyncService) restoreStatus(owner int, id string) (*CourseSyncStatus, error) {
	status := s.loadStatus(owner)
	if status == nil || status.ID != id {
		return nil, apperr.NotFound("教务同步任务不存在，请重新连接")
	}
	if !syncTerminal(status.State) {
		status.State = "failed"
		status.Message = "服务已重启，请重新扫码继续；已完成学期保留，采集缓存会自动复用"
		if status.Progress != nil {
			status.Progress.RetryAt = nil
		}
	}
	return status, nil
}

// A resumed connection verifies school identity afresh. Never skip personal
// semesters using a previous task's results: the next QR may be another account.
func (s *CourseSyncService) resumeSelection(owner int) []string {
	status := s.loadStatus(owner)
	if status == nil || status.State == "done" {
		return []string{}
	}
	return append([]string{}, status.Selected...)
}
