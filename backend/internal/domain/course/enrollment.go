package course

import (
	"context"
	"errors"
	"strings"
	"time"

	"jcourse/internal/domain/jaccount"
)

var (
	ErrSemesterRequired       = errors.New("semester required")
	ErrOfferedCourseNotFound  = errors.New("offered course not found")
	ErrInvalidSyncSemester    = errors.New("invalid semester")
	ErrEnrollmentSyncDisabled = errors.New("course enrollment sync disabled")
)

type CourseEnrollment struct {
	ID        int
	UserID    int
	CourseID  int
	Semester  string
	CreatedAt time.Time
}

type CourseEnrollmentView struct {
	ID        int
	UserID    int
	Course    CourseView
	Semester  string
	CreatedAt time.Time
}

type CourseEnrollmentSyncResult struct {
	Matched int64 `json:"matched"`
	Total   int   `json:"total"`
}

type CourseEnrollmentFilter struct {
	UserID   int
	CourseID int
}

type CourseEnrollmentRepository interface {
	Create(ctx context.Context, enrollment *CourseEnrollment) error
	SyncFromCoursePairs(ctx context.Context, userID int, semester string, pairs []CourseCodeTeacher) (int64, error)
	Delete(ctx context.Context, enrollmentID, userID int) error
}

type CourseEnrollmentQuery interface {
	FindUserEnrollments(ctx context.Context, userID int) ([]CourseEnrollmentView, error)
	FindUserCourseEnrollments(ctx context.Context, userID, courseID int) ([]CourseEnrollmentView, error)
}

type EnrollmentService struct {
	courseRepo     CourseRepository
	enrollmentRepo CourseEnrollmentRepository
	jaccountClient jaccount.Client
}

func NewEnrollmentService(courseRepo CourseRepository, enrollmentRepo CourseEnrollmentRepository, jaccountClient jaccount.Client) *EnrollmentService {
	return &EnrollmentService{courseRepo: courseRepo, enrollmentRepo: enrollmentRepo, jaccountClient: jaccountClient}
}

func (s *EnrollmentService) Create(ctx context.Context, userID, courseID int, semester string, now time.Time) error {
	semester = strings.TrimSpace(semester)
	if semester == "" {
		return ErrSemesterRequired
	}
	c, err := s.courseRepo.Get(ctx, courseID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCourseNotFound
	}
	exists := semester == c.LastSemester
	if !exists {
		exists, err = s.courseRepo.OfferedCourseExists(ctx, courseID, semester)
	}
	if err != nil {
		return err
	}
	if !exists {
		return ErrOfferedCourseNotFound
	}
	return s.enrollmentRepo.Create(ctx, &CourseEnrollment{
		UserID:    userID,
		CourseID:  courseID,
		Semester:  semester,
		CreatedAt: now,
	})
}

func (s *EnrollmentService) Delete(ctx context.Context, userID, enrollmentID int) error {
	return s.enrollmentRepo.Delete(ctx, enrollmentID, userID)
}

func (s *EnrollmentService) StartSync(ctx context.Context, semester, state string) (string, string, error) {
	semester, err := s.validateSyncSemester(ctx, semester)
	if err != nil {
		return "", "", err
	}
	if s.jaccountClient == nil {
		return "", "", ErrEnrollmentSyncDisabled
	}
	authURL, err := s.jaccountClient.AuthCodeURL(state)
	if err != nil {
		if errors.Is(err, jaccount.ErrDisabled) {
			return "", "", ErrEnrollmentSyncDisabled
		}
		return "", "", err
	}
	return semester, authURL, nil
}

func (s *EnrollmentService) SyncFromCode(ctx context.Context, userID int, semester, code string) (*CourseEnrollmentSyncResult, error) {
	if s.jaccountClient == nil {
		return nil, ErrEnrollmentSyncDisabled
	}
	token, err := s.jaccountClient.Exchange(ctx, code)
	if err != nil {
		if errors.Is(err, jaccount.ErrDisabled) {
			return nil, ErrEnrollmentSyncDisabled
		}
		return nil, err
	}
	lessons, err := s.jaccountClient.Lessons(ctx, token, semester)
	if err != nil {
		if errors.Is(err, jaccount.ErrDisabled) {
			return nil, ErrEnrollmentSyncDisabled
		}
		return nil, err
	}
	return s.SyncLessons(ctx, userID, semester, lessons)
}

func (s *EnrollmentService) validateSyncSemester(ctx context.Context, semester string) (string, error) {
	semester = strings.TrimSpace(semester)
	if semester == "" {
		return "", ErrSemesterRequired
	}
	allowed, err := s.courseRepo.OfferedSemesterExists(ctx, semester)
	if err != nil {
		return "", err
	}
	if !allowed {
		return "", ErrInvalidSyncSemester
	}
	return semester, nil
}

func (s *EnrollmentService) SyncLessons(ctx context.Context, userID int, semester string, lessons []jaccount.LessonCourse) (*CourseEnrollmentSyncResult, error) {
	semester = strings.TrimSpace(semester)
	if semester == "" {
		return nil, ErrSemesterRequired
	}
	pairs := make([]CourseCodeTeacher, 0, len(lessons))
	for _, lesson := range lessons {
		code := strings.TrimSpace(lesson.Code)
		teacherName := strings.TrimSpace(lesson.TeacherName)
		if code == "" || teacherName == "" {
			continue
		}
		pairs = append(pairs, CourseCodeTeacher{Code: code, TeacherName: teacherName})
	}
	matched, err := s.enrollmentRepo.SyncFromCoursePairs(ctx, userID, semester, pairs)
	if err != nil {
		return nil, err
	}
	return &CourseEnrollmentSyncResult{Matched: matched, Total: len(pairs)}, nil
}
