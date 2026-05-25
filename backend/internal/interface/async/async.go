package async

import (
	"github.com/hibiken/asynq"

	"jcourse/config"
	"jcourse/internal/app"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/stat"
)

func NewMux(container *app.ServiceContainer) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.Use(newTaskLoggingMiddleware())
	mux.Handle(auth.TaskTypeClearExpiredSuspension, newClearExpiredSuspensionHandler(container.AuthUserService))
	mux.Handle(auth.TaskTypeFlushAccess, newFlushAccessHandler(container.AccessTracker))
	mux.Handle(course.TaskTypeRecordHotCourseActivity, newRecordHotCourseActivityHandler(container.CourseHotCommand))
	mux.Handle(stat.TaskTypeCollectDailySiteStats, newCollectDailySiteStatsHandler(container.SiteStatsCommand))
	return mux
}

func RegisterScheduledTasks(scheduler *asynq.Scheduler, conf config.AppConfig) (bool, error) {
	registered := false
	if conf.Stats.SchedulerEnabled && conf.Stats.DailyCron != "" {
		task := stat.NewCollectDailySiteStatsTask("")
		if _, err := scheduler.Register(conf.Stats.DailyCron, asynq.NewTask(task.Type(), task.Payload())); err != nil {
			return false, err
		}
		registered = true
	}
	if conf.Auth.Access.SchedulerEnabled && conf.Auth.Access.FlushCron != "" {
		task := auth.NewFlushAccessTask()
		if _, err := scheduler.Register(conf.Auth.Access.FlushCron, asynq.NewTask(task.Type(), task.Payload())); err != nil {
			return false, err
		}
		registered = true
	}
	return registered, nil
}
