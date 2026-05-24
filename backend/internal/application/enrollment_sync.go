package application

import (
	"context"
	"errors"
	"strings"

	"jcourse/internal/domain/course"
	"jcourse/internal/domain/jaccount"
)

type CourseEnrollmentSyncResult struct {
	Matched int64 `json:"matched"`
	Total   int   `json:"total"`
}

type CourseEnrollmentSyncService struct {
	enrollmentRepo course.CourseEnrollmentRepository
	courseRepo     course.CourseRepository
	jaccountClient jaccount.Client
}

func NewCourseEnrollmentSyncService(enrollmentRepo course.CourseEnrollmentRepository, courseRepo course.CourseRepository, jaccountClient jaccount.Client) *CourseEnrollmentSyncService {
	return &CourseEnrollmentSyncService{enrollmentRepo: enrollmentRepo, courseRepo: courseRepo, jaccountClient: jaccountClient}
}

func (s *CourseEnrollmentSyncService) Start(ctx context.Context, semester, state string) (string, string, error) {
	semester, err := s.validateSemester(ctx, semester)
	if err != nil {
		return "", "", err
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

func (s *CourseEnrollmentSyncService) SyncFromCode(ctx context.Context, userID int, semester, code string) (*CourseEnrollmentSyncResult, error) {
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

func (s *CourseEnrollmentSyncService) validateSemester(ctx context.Context, semester string) (string, error) {
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

func (s *CourseEnrollmentSyncService) SyncLessons(ctx context.Context, userID int, semester string, lessons []jaccount.LessonCourse) (*CourseEnrollmentSyncResult, error) {
	semester = strings.TrimSpace(semester)
	if semester == "" {
		return nil, ErrSemesterRequired
	}
	pairs := make([]course.CourseCodeTeacher, 0, len(lessons))
	for _, lesson := range lessons {
		code := strings.TrimSpace(lesson.Code)
		teacherName := strings.TrimSpace(lesson.TeacherName)
		if code == "" || teacherName == "" {
			continue
		}
		pairs = append(pairs, course.CourseCodeTeacher{Code: code, TeacherName: teacherName})
	}
	matched, err := s.enrollmentRepo.SyncFromCoursePairs(ctx, userID, semester, pairs)
	if err != nil {
		return nil, err
	}
	return &CourseEnrollmentSyncResult{Matched: matched, Total: len(pairs)}, nil
}

var (
	ErrInvalidSyncSemester    = errors.New("invalid semester")
	ErrEnrollmentSyncDisabled = errors.New("course enrollment sync disabled")
)
