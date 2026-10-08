package application

import (
	"context"
	"encoding/json"
	"jcourse/internal/domain/course"
	"jcourse/pkg/apperr"
	"strings"
	"sync"
	"time"
)

// Each user has an isolated short-lived collector; one user cannot monopolize
// the administrative source or inspect another user's QR or results.
type EnrollmentSyncService struct {
	mu       sync.Mutex
	jobs     map[int]*CourseSyncService
	source   course.CourseSource
	importer course.EnrollmentImporter
	limit    int
	queue    *CollectorQueue
}

func NewEnrollmentSyncService(source course.CourseSource, importer course.EnrollmentImporter, queues ...*CollectorQueue) *EnrollmentSyncService {
	s := &EnrollmentSyncService{jobs: make(map[int]*CourseSyncService), source: source, importer: importer, limit: 128}
	if len(queues) > 0 {
		s.queue = queues[0]
	}
	return s
}
func (s *EnrollmentSyncService) Start(owner int) (*CourseSyncStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	active := 0
	for user, service := range s.jobs {
		service.mu.Lock()
		session := service.session
		service.mu.Unlock()
		status := session.snapshot()
		if !syncTerminal(status.State) && time.Now().Before(status.ExpiresAt) {
			if user == owner {
				return nil, apperr.Conflict("请先完成或取消正在进行的课程导入")
			}
			active++
		} else if time.Now().After(status.ExpiresAt) {
			delete(s.jobs, user)
		}
	}
	if active >= s.limit {
		return nil, apperr.Conflict("当前扫码导入人数较多，请稍后重试")
	}
	// Completed reports are transient. Bound retained users as well as active jobs.
	if len(s.jobs) >= s.limit {
		oldestUser := 0
		var oldest time.Time
		for user, service := range s.jobs {
			status := service.session.snapshot()
			if syncTerminal(status.State) && (oldestUser == 0 || status.ExpiresAt.Before(oldest)) {
				oldestUser = user
				oldest = status.ExpiresAt
			}
		}
		if oldestUser != 0 {
			delete(s.jobs, oldestUser)
		}
	}
	service := &CourseSyncService{source: s.source, lifetime: 15 * time.Minute, personal: true, queue: s.queue}
	service.importEvent = func(ctx context.Context, event course.SourceEvent) (*CourseImportResult, error) {
		var document struct {
			Semester string                          `json:"semester"`
			Courses  []course.EnrollmentSourceCourse `json:"courses"`
		}
		if len(event.Document) > 2<<20 || json.Unmarshal(event.Document, &document) != nil || document.Semester != event.Semester || document.Courses == nil || len(document.Courses) > 1000 {
			return nil, apperr.BadRequest("个人课表数据不完整，请重新导入")
		}
		for _, item := range document.Courses {
			if strings.TrimSpace(item.Name) == "" || (strings.TrimSpace(item.LessonNo) == "" && strings.TrimSpace(item.SourceLessonID) == "") {
				return nil, apperr.BadRequest("课程缺少教学班信息，请重新导入")
			}
		}
		result, err := s.importer.ImportSemester(ctx, owner, document.Semester, document.Courses)
		if err != nil {
			return nil, err
		}
		// Keep exact counts, but do not retain thousands of long names per semester.
		result.UnmatchedCount = len(result.Unmatched)
		sample := make([]course.EnrollmentSourceCourse, 0, min(20, len(result.Unmatched)))
		for _, item := range result.Unmatched[:min(20, len(result.Unmatched))] {
			item.Name = shortImportLabel(item.Name)
			item.LessonNo = shortImportLabel(item.LessonNo)
			item.SourceLessonID = shortImportLabel(item.SourceLessonID)
			sample = append(sample, item)
		}
		result.Unmatched = sample
		return &CourseImportResult{Semester: document.Semester, Enrollment: result}, nil
	}
	status, err := service.Start(owner)
	if err == nil {
		s.jobs[owner] = service
	}
	return status, err
}

func shortImportLabel(value string) string {
	runes := []rune(value)
	if len(runes) > 128 {
		return string(runes[:128]) + "…"
	}
	return string(runes)
}
func (s *EnrollmentSyncService) owned(owner int) (*CourseSyncService, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	service := s.jobs[owner]
	if service == nil {
		return nil, apperr.NotFound("课程导入任务不存在，请重新扫码")
	}
	return service, nil
}
func (s *EnrollmentSyncService) Status(owner int, id string) (*CourseSyncStatus, error) {
	service, err := s.owned(owner)
	if err != nil {
		return nil, err
	}
	return service.Status(owner, id)
}
func (s *EnrollmentSyncService) Sync(owner int, id string, selected []string) (*CourseSyncStatus, error) {
	service, err := s.owned(owner)
	if err != nil {
		return nil, err
	}
	return service.Sync(owner, id, selected)
}
func (s *EnrollmentSyncService) Cancel(owner int, id string) (*CourseSyncStatus, error) {
	service, err := s.owned(owner)
	if err != nil {
		return nil, err
	}
	return service.Cancel(owner, id)
}
