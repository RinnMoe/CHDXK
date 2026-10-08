package application

import (
	"context"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/setting"
)

var ErrCourseNotFound = course.ErrCourseNotFound

type CourseCommandService struct {
	courseService        *course.Service
	notificationService  *course.NotificationService
	enrollmentService    *course.EnrollmentService
	hotService           *course.CourseHotService
	ratingCommandService *course.CourseRatingCommandService
	settings             SiteSettingsProvider
	importSettings       *setting.SystemSettingsService
}

func NewCourseCommandService(
	courseService *course.Service,
	notificationService *course.NotificationService,
	enrollmentService *course.EnrollmentService,
	hotService *course.CourseHotService,
	ratingCommandService *course.CourseRatingCommandService,
	settings SiteSettingsProvider,
) *CourseCommandService {
	if settings == nil {
		defaults := NewDefaultSiteSettingsProvider()
		settings = defaults
	}
	return &CourseCommandService{
		courseService:        courseService,
		notificationService:  notificationService,
		enrollmentService:    enrollmentService,
		hotService:           hotService,
		ratingCommandService: ratingCommandService,
		settings:             settings,
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

func (s *CourseCommandService) DeleteEnrollment(ctx context.Context, userID, enrollmentID int) error {
	return s.enrollmentService.Delete(ctx, userID, enrollmentID)
}

func (s *CourseCommandService) SetImportSettings(settings *setting.SystemSettingsService) {
	s.importSettings = settings
}

func (s *CourseCommandService) SaveEnrollment(ctx context.Context, userID, enrollmentID, sectionID int) error {
	if enrollmentID == 0 {
		if err := requireImportEnabled(ctx, s.importSettings, ImportManualEnabled); err != nil {
			return err
		}
	}
	return s.enrollmentService.Save(ctx, userID, enrollmentID, sectionID)
}

func (s *CourseCommandService) RecordActivity(ctx context.Context, payload course.RecordHotCourseActivityPayload) error {
	scores, err := s.settings.HotScoreConfig(ctx)
	if err != nil {
		return err
	}
	return s.hotService.RecordActivityWithScores(ctx, payload, scores)
}

func (s *CourseCommandService) RefreshRatingScores(ctx context.Context) error {
	return s.ratingCommandService.RefreshRatingScores(ctx)
}
