package async

import (
	"github.com/hibiken/asynq"

	"jcourse/config"
	"jcourse/internal/app"
	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	domainemail "jcourse/internal/domain/email"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/stat"
	asynchandler "jcourse/internal/interface/async/handler"
)

func NewMux(container *app.ServiceContainer) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.Use(newTaskLoggingMiddleware())
	mux.Handle(auth.TaskTypeClearExpiredSuspension, asynchandler.NewClearExpiredSuspensionHandler(container.AuthUserService))
	mux.Handle(auth.TaskTypeSuspendUser, asynchandler.NewSuspendUserHandler(container.AuthUserService))
	mux.Handle(auth.TaskTypeFlushAccess, asynchandler.NewFlushAccessHandler(container.AccessTracker))
	mux.Handle(course.TaskTypeRecordHotCourseActivity, asynchandler.NewRecordHotCourseActivityHandler(container.CourseCommand))
	mux.Handle(course.TaskTypeRefreshRatingScores, asynchandler.NewRefreshCourseRatingScoresHandler(container.CourseCommand))
	mux.Handle(point.TaskTypeGrantReward, asynchandler.NewGrantPointRewardHandler(container.PointRewardCommand))
	mux.Handle(domainemail.TaskTypeSendEmail, asynchandler.NewSendEmailHandler(container.EmailSender))
	mux.Handle(stat.TaskTypeCollectDailySiteStats, asynchandler.NewCollectDailySiteStatsHandler(container.SiteStatsCommand))
	mux.Handle(audit.TaskTypeRecordLog, asynchandler.NewRecordAuditLogHandler(container.AuditLogCommand))
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
	ratingScore := conf.Course.RatingScore.Normalized()
	if ratingScore.SchedulerEnabled && ratingScore.RefreshCron != "" {
		task := course.NewRefreshRatingScoresTask()
		if _, err := scheduler.Register(ratingScore.RefreshCron, asynq.NewTask(task.Type(), task.Payload())); err != nil {
			return false, err
		}
		registered = true
	}
	return registered, nil
}
