package application

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"jcourse/internal/domain/teacher"
	"jcourse/pkg/apperr"
)

type TeacherProfileSyncStatus struct {
	State     string `json:"state"`
	Stage     string `json:"stage"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
	Failed    int    `json:"failed"`
	Message   string `json:"message"`
}
type TeacherProfileService struct {
	repo   teacher.ProfileStore
	source teacher.ProfileSource
	mu     sync.Mutex
	status TeacherProfileSyncStatus
	cancel context.CancelFunc
}

func NewTeacherProfileService(repo teacher.ProfileStore, source teacher.ProfileSource) *TeacherProfileService {
	return &TeacherProfileService{repo: repo, source: source, status: TeacherProfileSyncStatus{State: "idle"}}
}
func (s *TeacherProfileService) Status() TeacherProfileSyncStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}
func (s *TeacherProfileService) Matches(ctx context.Context) ([]teacher.ProfileMatch, error) {
	return s.repo.Matches(ctx)
}
func (s *TeacherProfileService) Bind(ctx context.Context, id int, address string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status.State == "running" {
		return apperr.BadRequest("请等待同步完成后处理重名")
	}
	return s.repo.Bind(ctx, id, address)
}
func (s *TeacherProfileService) Start() (TeacherProfileSyncStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status.State == "running" {
		return s.status, apperr.BadRequest("教师资料同步正在进行")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	s.cancel = cancel
	s.status = TeacherProfileSyncStatus{State: "running", Stage: "catalog"}
	go s.run(ctx, cancel)
	return s.status, nil
}
func (s *TeacherProfileService) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
}
func (s *TeacherProfileService) fail(ctx context.Context, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.State = "failed"
	s.status.Message = message
	if ctx.Err() != nil {
		s.status.State = "cancelled"
		s.status.Message = "同步已取消或超时，原有资料保留"
	}
}
func (s *TeacherProfileService) run(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	teachers, err := s.repo.Teachers(ctx)
	if err != nil {
		s.fail(ctx, "读取本站教师失败，原有资料保留")
		return
	}
	profiles, err := s.source.List(ctx, func(done, total int) { s.mu.Lock(); s.status.Completed = done; s.status.Total = total; s.mu.Unlock() })
	if err != nil {
		s.fail(ctx, "教师主页目录读取失败，原有资料保留，请稍后重试")
		return
	}
	keys := map[string]bool{}
	for _, t := range teachers {
		keys[teacher.ProfileIdentity(t.Name, t.Department)] = true
	}
	indexes := []int{}
	for i, p := range profiles {
		if keys[teacher.ProfileIdentity(p.Name, p.Department)] {
			indexes = append(indexes, i)
		}
	}
	s.mu.Lock()
	s.status.Stage = "details"
	s.status.Completed = 0
	s.status.Total = len(indexes)
	s.mu.Unlock()
	// Three HTTP workers; no browser processes or school login are needed.
	jobs := make(chan int)
	var workers sync.WaitGroup
	for worker := 0; worker < 3; worker++ {
		workers.Go(func() {
			for i := range jobs {
				detail, err := s.source.Detail(ctx, profiles[i])
				if err == nil {
					profiles[i] = detail
				}
				s.mu.Lock()
				s.status.Completed++
				if err != nil {
					s.status.Failed++
				}
				s.mu.Unlock()
				select {
				case <-ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
				}
			}
		})
	}
	for _, i := range indexes {
		select {
		case jobs <- i:
		case <-ctx.Done():
			close(jobs)
			workers.Wait()
			s.fail(ctx, "")
			return
		}
	}
	close(jobs)
	workers.Wait()
	if ctx.Err() != nil {
		s.fail(ctx, "")
		return
	}
	s.mu.Lock()
	s.status.Stage = "saving"
	s.mu.Unlock()
	if err := s.repo.Save(ctx, teachers, profiles); err != nil {
		s.fail(ctx, "保存教师资料失败，原有资料保留")
		return
	}
	s.mu.Lock()
	s.status.State = "done"
	s.status.Message = "教师资料同步完成，请在重名处理页确认无法唯一匹配的教师"
	s.mu.Unlock()
}

func (s *TeacherProfileService) Teachers(ctx context.Context, f teacher.TeacherFilter) ([]TeacherDTO, int64, error) {
	rows, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	result := make([]TeacherDTO, len(rows))
	for i := range rows {
		result[i] = newTeacherDTO(&rows[i])
	}
	return result, total, nil
}
func (s *TeacherProfileService) Edit(ctx context.Context, id int, p teacher.Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status.State == "running" {
		return apperr.BadRequest("请等待同步完成后修改资料")
	}
	p.Name = strings.TrimSpace(p.Name)
	p.Department = strings.TrimSpace(p.Department)
	p.URL = strings.TrimSpace(p.URL)
	if p.Name == "" || p.Department == "" || len(p.Name) > 300 || len(p.Department) > 300 || len(p.Title) > 300 {
		return apperr.BadRequest("请填写有效的姓名、学院和职称")
	}
	if p.URL != "" {
		u, err := url.Parse(p.URL)
		if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
			return apperr.BadRequest("教师主页请填写完整的网页地址")
		}
	}
	if len(p.URL) > 2000 {
		return apperr.BadRequest("教师主页地址过长")
	}
	allowed := map[string]bool{"学位": true, "联系方式": true, "办公地点": true, "研究方向": true, "个人简介": true}
	fields := map[string]string{}
	for k, v := range p.Fields {
		if !allowed[k] || len(v) > 30000 {
			return apperr.BadRequest("资料字段或内容长度无效")
		}
		fields[k] = strings.TrimSpace(v)
	}
	p.Fields = fields
	p.UpdatedAt = time.Now()
	return s.repo.Edit(ctx, id, p)
}

type TeacherProfilePreview struct {
	Candidates []teacher.Profile `json:"candidates"`
	Profile    *teacher.Profile  `json:"profile,omitempty"`
}

// Preview is read-only: it never saves or starts a global synchronization.
func (s *TeacherProfileService) Preview(ctx context.Context, id int, name, department, address string) (*TeacherProfilePreview, error) {
	name = strings.TrimSpace(name)
	department = strings.TrimSpace(department)
	if name == "" || department == "" || len(name) > 300 || len(department) > 300 {
		return nil, apperr.BadRequest("请先填写姓名和学院")
	}
	rows, _, err := s.repo.List(ctx, teacher.TeacherFilter{TeacherIDs: []int{id}})
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, apperr.BadRequest("教师不存在")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	candidates, err := s.source.Find(ctx, name, department)
	if err != nil {
		return nil, apperr.BadRequest("教师主页查询失败，请稍后重试，现有资料保留")
	}
	result := &TeacherProfilePreview{Candidates: candidates}
	if address == "" && len(candidates) == 1 {
		address = candidates[0].URL
	}
	if address == "" {
		return result, nil
	}
	for _, candidate := range candidates {
		if candidate.URL == address {
			detail, err := s.source.Detail(ctx, candidate)
			if err != nil {
				return nil, apperr.BadRequest("该教师主页资料读取失败，请稍后重试，现有资料保留")
			}
			result.Profile = &detail
			return result, nil
		}
	}
	return nil, apperr.BadRequest("请选择姓名与学院匹配的候选主页")
}
