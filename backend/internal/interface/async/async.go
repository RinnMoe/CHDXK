package async

import (
	"log"

	"github.com/hibiken/asynq"

	"jcourse/internal/app"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/stat"
)

func NewMux(container *app.ServiceContainer) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.Use(newTaskLoggingMiddleware(log.Default()))
	mux.Handle(auth.TaskTypeClearExpiredSuspension, newClearExpiredSuspensionHandler(container.AuthUserService))
	mux.Handle(course.TaskTypeRecordHotCourseActivity, newRecordHotCourseActivityHandler(container.CourseHotCommand))
	mux.Handle(stat.TaskTypeCollectDailySiteStats, newCollectDailySiteStatsHandler(container.SiteStatsCommand))
	return mux
}
