package application

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"

	"jcourse/internal/domain/course"
	"jcourse/pkg/apperr"
)

type CourseSyncStatus struct {
	ID        string                 `json:"id"`
	State     string                 `json:"state"`
	QR        string                 `json:"qr,omitempty"`
	Semesters []string               `json:"semesters"`
	Selected  []string               `json:"selected"`
	Current   string                 `json:"current,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Results   []CourseImportResult   `json:"results"`
	ExpiresAt time.Time              `json:"expires_at"`
	StartedAt *time.Time             `json:"started_at,omitempty"`
	Progress  *course.SourceProgress `json:"progress,omitempty"`
}
type courseSyncSession struct {
	mu         sync.Mutex
	owner      int
	status     CourseSyncStatus
	connection course.SourceConnection
	cancel     context.CancelFunc
}
type CourseSyncService struct {
	mu            sync.Mutex
	session       *courseSyncSession
	source        course.CourseSource
	importArchive func(context.Context, []byte) (*CourseImportResult, error)
	importEvent   func(context.Context, course.SourceEvent) (*CourseImportResult, error)
	lifetime      time.Duration
	personal      bool
}

func NewCourseSyncService(source course.CourseSource, management *CourseManagementService) *CourseSyncService {
	return &CourseSyncService{source: source, importArchive: management.ImportArchive}
}
func (s *CourseSyncService) Start(owner int) (*CourseSyncStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old := s.session; old != nil {
		old.mu.Lock()
		busy := !syncTerminal(old.status.State) && time.Now().Before(old.status.ExpiresAt)
		old.mu.Unlock()
		if busy {
			return nil, apperr.Conflict("已有教务连接或同步任务，请完成或取消后重试")
		}
		old.cancel()
	}
	lifetime := s.lifetime
	if lifetime == 0 {
		lifetime = 24 * time.Hour
	}
	ctx, cancel := context.WithTimeout(context.Background(), lifetime)
	connection, err := s.source.Connect(ctx)
	if err != nil {
		cancel()
		return nil, apperr.ServiceUnavailable("教务同步暂不可用，请联系技术管理员检查采集器配置")
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		cancel()
		connection.Close()
		return nil, err
	}
	session := &courseSyncSession{owner: owner, connection: connection, cancel: cancel, status: CourseSyncStatus{ID: hex.EncodeToString(id), State: "connecting", Semesters: []string{}, Selected: []string{}, Results: []CourseImportResult{}, ExpiresAt: time.Now().Add(lifetime)}}
	s.session = session
	go s.consume(ctx, session)
	return session.snapshot(), nil
}
func syncTerminal(state string) bool {
	return state == "done" || state == "failed" || state == "cancelled"
}
func (session *courseSyncSession) snapshot() *CourseSyncStatus {
	session.mu.Lock()
	defer session.mu.Unlock()
	status := session.status
	status.Semesters = append([]string{}, status.Semesters...)
	status.Selected = append([]string{}, status.Selected...)
	status.Results = append([]CourseImportResult{}, status.Results...)
	if status.Progress != nil {
		progress := *status.Progress
		if progress.Total != nil {
			total := *progress.Total
			progress.Total = &total
		}
		status.Progress = &progress
	}
	if status.StartedAt != nil {
		started := *status.StartedAt
		status.StartedAt = &started
	}
	return &status
}
func (s *CourseSyncService) owned(owner int, id string) (*courseSyncSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session := s.session
	if session == nil || session.owner != owner || session.status.ID != id {
		return nil, apperr.NotFound("教务同步任务不存在，请重新连接")
	}
	return session, nil
}
func (s *CourseSyncService) Status(owner int, id string) (*CourseSyncStatus, error) {
	session, err := s.owned(owner, id)
	if err != nil {
		return nil, err
	}
	return session.snapshot(), nil
}
func (s *CourseSyncService) Sync(owner int, id string, selected []string) (*CourseSyncStatus, error) {
	session, err := s.owned(owner, id)
	if err != nil {
		return nil, err
	}
	session.mu.Lock()
	if session.status.State != "ready" || len(selected) == 0 {
		session.mu.Unlock()
		return nil, apperr.BadRequest("请完成扫码登录并选择学期")
	}
	allowed := make(map[string]bool)
	for _, semester := range session.status.Semesters {
		allowed[semester] = true
	}
	unique := make(map[string]bool)
	for _, semester := range selected {
		if !allowed[semester] || unique[semester] {
			session.mu.Unlock()
			return nil, apperr.BadRequest("请选择有效且不重复的学期")
		}
		unique[semester] = true
	}
	session.status.Selected = append([]string{}, selected...)
	session.status.State = "running"
	now := time.Now()
	session.status.StartedAt = &now
	session.status.QR = ""
	err = session.connection.Sync(selected)
	session.mu.Unlock()
	if err != nil {
		s.fail(session, "连接已中断，请重新扫码登录")
		return nil, apperr.ServiceUnavailable("连接已中断，请重新扫码登录")
	}
	return session.snapshot(), nil
}
func (s *CourseSyncService) Cancel(owner int, id string) (*CourseSyncStatus, error) {
	session, err := s.owned(owner, id)
	if err != nil {
		return nil, err
	}
	session.mu.Lock()
	if !syncTerminal(session.status.State) {
		session.status.State = "cancelled"
		session.status.QR = ""
		session.cancel()
	}
	session.mu.Unlock()
	return session.snapshot(), nil
}
func (s *CourseSyncService) fail(session *courseSyncSession, message string) {
	session.mu.Lock()
	if !syncTerminal(session.status.State) {
		session.status.State = "failed"
		session.status.Message = message
		session.status.QR = ""
	}
	session.mu.Unlock()
	session.cancel()
}
func (s *CourseSyncService) consume(ctx context.Context, session *courseSyncSession) {
	defer session.connection.Close()
	defer session.cancel()
	for {
		select {
		case <-ctx.Done():
			s.fail(session, "连接已到期，请重新扫码登录")
			return
		case event, ok := <-session.connection.Events():
			if !ok {
				s.fail(session, "连接已中断，请重新扫码登录")
				return
			}
			session.mu.Lock()
			if syncTerminal(session.status.State) {
				session.mu.Unlock()
				return
			}
			switch event.Type {
			case "qr":
				if strings.HasPrefix(event.QR, "data:image/png;base64,") && len(event.QR) < 1<<20 {
					session.status.State = "waiting"
					session.status.QR = event.QR
				}
			case "scanned":
				session.status.State = "scanned"
			case "authenticating", "discovering":
				session.status.State = event.Type
				session.status.QR = ""
			case "ready":
				session.status.QR = ""
				for _, semester := range event.Semesters {
					if !semesterPattern.MatchString(semester) {
						session.mu.Unlock()
						s.fail(session, "教务学期列表读取失败")
						return
					}
				}
				if len(event.Semesters) == 0 {
					session.mu.Unlock()
					s.fail(session, "未找到可同步的学期")
					return
				}
				session.status.State = "ready"
				session.status.Semesters = append([]string{}, event.Semesters...)
			case "collecting", "progress":
				if session.status.State == "running" && containsSemester(session.status.Selected, event.Semester) {
					session.status.Current = event.Semester
					if event.Type == "collecting" {
						session.status.Progress = &course.SourceProgress{Stage: "listing", UpdatedAt: time.Now()}
					} else if validSourceProgress(event.Progress) {
						progress := *event.Progress
						progress.UpdatedAt = time.Now()
						session.status.Progress = &progress
					}
				}
			case "archive":
				session.status.Progress = &course.SourceProgress{Stage: "validating", UpdatedAt: time.Now()}
				status := session.status
				session.mu.Unlock()
				duplicate := false
				for _, previous := range status.Results {
					if previous.Semester == event.Semester {
						duplicate = true
					}
				}
				if status.State != "running" || !containsSemester(status.Selected, event.Semester) || duplicate {
					s.fail(session, "同步学期与所选学期不一致")
					return
				}
				session.mu.Lock()
				session.status.Progress = &course.SourceProgress{Stage: "importing", UpdatedAt: time.Now()}
				session.mu.Unlock()
				var result *CourseImportResult
				var err error
				if s.importEvent != nil {
					result, err = s.importEvent(ctx, event)
				} else {
					var archive []byte
					archive, err = sourceArchive(event)
					if err != nil {
						s.fail(session, "本学期采集结果不完整，未导入，请重新同步")
						return
					}
					result, err = s.importArchive(ctx, archive)
				}
				if err != nil {
					s.fail(session, "本学期未通过导入校验或保存失败，已停止同步")
					return
				}
				session.mu.Lock()
				session.status.Results = append(session.status.Results, *result)
				session.status.Progress = &course.SourceProgress{Stage: "completed", UpdatedAt: time.Now()}
			case "done":
				if len(session.status.Results) != len(session.status.Selected) || session.status.State != "running" {
					session.mu.Unlock()
					s.fail(session, "同步未完成，请检查学期结果后重新连接")
					return
				}
				session.status.State = "done"
			case "error":
				slog.Warn("academic collection stopped", "job_id", session.status.ID, "semester", session.status.Current, "error_code", safeSourceErrorCode(event.Error))
				session.mu.Unlock()
				message := "教务连接或采集失败，请检查教务系统是否可访问后重新连接"
				if event.Error == "QR_EXPIRED" {
					message = "二维码已过期，请重新连接"
				}
				if event.Error == "WECHAT_UNBOUND" {
					message = "请先在学校统一登录页面绑定微信，再重新连接"
				}
				if event.Error == "QR_LOGIN_FAILED" {
					message = "扫码登录未完成，请重新连接并在手机上确认登录"
				}
				if event.Error == "AUTH_REQUIRED" {
					message = "学校登录已确认，但未能进入本科教务系统，请重新连接"
				}
				if s.personal && event.Error == "AUTH_REQUIRED" {
					message = "教务登录已过期，请重新扫码；已完成学期的记录保留"
				}
				if event.Error == "PERSONAL_REQUEST_FAILED" {
					message = "本学期课表请求失败，自动重试后仍未成功；已完成学期的记录保留，请稍后重试"
				}
				if event.Error == "PERSONAL_RENDER_FAILED" {
					message = "本学期课表未能加载，自动重试后仍未成功；已完成学期的记录保留"
				}
				if event.Error == "PERSONAL_PARSE_FAILED" {
					message = "本学期课表格式无法读取，未保存本学期，请联系技术管理员；已完成学期的记录保留"
				}
				if s.personal && session.snapshot().Current != "" {
					message = session.snapshot().Current + "：" + message
				}
				if event.Error == "SEMESTER_DISCOVERY_FAILED" || event.Error == "PAGE_STRUCTURE" {
					message = "未能读取教务学期列表，请稍后重新连接；如持续失败，请联系技术管理员"
				}
				if event.Error == "PERSONAL_MENU_MISSING" || event.Error == "PERSONAL_QUERY_FAILED" {
					message = "未能打开个人课表，请确认学校账号可查看本科教务的个人课表后重新连接"
				}
				if event.Error == "PERSONAL_CALENDAR_FAILED" {
					message = "个人课表的学期列表未能加载，请稍后重新连接"
				}
				if event.Error == "PERMISSION_DENIED" {
					message = "当前学校账号无权访问全校开课查询，请使用有访问权限的账号重新连接"
					if s.personal {
						message = "当前学校账号无权访问个人课表，请使用学生账号重新连接"
					}
				}
				s.fail(session, message)
				return
			}
			done := syncTerminal(session.status.State)
			session.mu.Unlock()
			if done {
				return
			}
		}
	}
}

func safeSourceErrorCode(code string) string {
	switch code {
	case "PERSONAL_REQUEST_FAILED", "PERSONAL_RENDER_FAILED", "PERSONAL_PARSE_FAILED", "PERSONAL_MENU_MISSING", "PERSONAL_QUERY_FAILED", "PERSONAL_CALENDAR_FAILED", "AUTH_REQUIRED", "PERMISSION_DENIED", "QR_EXPIRED", "QR_LOGIN_FAILED", "COURSE_COLLECTION_FAILED", "SEMESTER_DISCOVERY_FAILED", "PAGE_STRUCTURE":
		return code
	default:
		return "SOURCE_UNAVAILABLE"
	}
}
func containsSemester(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}

func validSourceProgress(p *course.SourceProgress) bool {
	if p == nil || p.Completed < 0 || p.Failed < 0 || p.Retries < 0 || p.Page < 0 || p.Pages < 0 {
		return false
	}
	if p.Stage != "listing" && p.Stage != "details" && p.Stage != "validating" {
		return false
	}
	if p.Total != nil && (*p.Total < 0 || p.Completed > *p.Total || p.Failed > *p.Total-p.Completed) {
		return false
	}
	return p.Page <= p.Pages
}
func sourceArchive(event course.SourceEvent) ([]byte, error) {
	var document struct {
		Semester string `json:"semester"`
	}
	if json.Unmarshal(event.Document, &document) != nil || document.Semester != event.Semester {
		return nil, apperr.BadRequest("采集学期不一致")
	}
	if len(event.Document)+len(event.Report) > maxCourseArchiveSize {
		return nil, apperr.BadRequest("采集结果过大")
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, file := range []struct {
		name  string
		value []byte
	}{{"semester.json", event.Document}, {"report.json", event.Report}} {
		entry, err := writer.Create(file.name)
		if err != nil {
			return nil, err
		}
		if _, err := entry.Write(file.value); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
