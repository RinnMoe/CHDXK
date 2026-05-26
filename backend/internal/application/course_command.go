package application

import (
	"context"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
)

var ErrCourseNotFound = course.ErrCourseNotFound

type CourseCommandService struct {
	courseService        *course.Service
	notificationService  *course.NotificationService
	enrollmentService    *course.EnrollmentService
	hotService           *course.CourseHotService
	ratingCommandService *course.CourseRatingCommandService
}

func NewCourseCommandService(
	courseService *course.Service,
	notificationService *course.NotificationService,
	enrollmentService *course.EnrollmentService,
	hotService *course.CourseHotService,
	ratingCommandService *course.CourseRatingCommandService,
) *CourseCommandService {
	return &CourseCommandService{
		courseService:        courseService,
		notificationService:  notificationService,
		enrollmentService:    enrollmentService,
		hotService:           hotService,
		ratingCommandService: ratingCommandService,
	}
}

type UpdateCourseModeratorRemarkCommand struct {
	ModeratorRemark string `json:"moderator_remark,omitempty"`
}

func (s *CourseCommandService) UpdateModeratorRemark(ctx context.Context, u *auth.User, courseID int, cmd *UpdateCourseModeratorRemarkCommand) error {
	return s.courseService.UpdateModeratorRemark(ctx, u, course.UpdateModeratorRemark{
		CourseID:        courseID,
		ModeratorRemark: cmd.ModeratorRemark,
	})
}

func (s *CourseCommandService) SetNotificationLevel(ctx context.Context, userID, courseID int, level course.NotificationLevel) error {
	return s.notificationService.SetLevel(ctx, userID, courseID, level)
}

func (s *CourseCommandService) CreateEnrollment(ctx context.Context, userID int, cmd CreateCourseEnrollmentCommand) error {
	return s.enrollmentService.Create(ctx, userID, cmd.CourseID, cmd.Semester, time.Now())
}

func (s *CourseCommandService) DeleteEnrollment(ctx context.Context, userID, enrollmentID int) error {
	return s.enrollmentService.Delete(ctx, userID, enrollmentID)
}

func (s *CourseCommandService) RecordActivity(ctx context.Context, payload course.RecordHotCourseActivityPayload) error {
	return s.hotService.RecordActivity(ctx, payload)
}

func (s *CourseCommandService) StartEnrollmentSync(ctx context.Context, semester, state string) (string, string, error) {
	return s.enrollmentService.StartSync(ctx, semester, state)
}

func (s *CourseCommandService) SyncEnrollmentFromCode(ctx context.Context, userID int, semester, code string) (*course.CourseEnrollmentSyncResult, error) {
	return s.enrollmentService.SyncFromCode(ctx, userID, semester, code)
}

func (s *CourseCommandService) RefreshRatingScores(ctx context.Context) error {
	return s.ratingCommandService.RefreshRatingScores(ctx)
}
