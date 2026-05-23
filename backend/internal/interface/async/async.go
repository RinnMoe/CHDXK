package async

import (
	"github.com/hibiken/asynq"

	"jcourse/internal/app"
	domainauth "jcourse/internal/domain/auth"
	"jcourse/internal/domain/stat"
)

func NewMux(container *app.ServiceContainer) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.Handle(domainauth.TaskTypeClearExpiredSuspension, newClearExpiredSuspensionHandler(container.CurrentUserService))
	mux.Handle(stat.TaskTypeCollectDailySiteStats, newCollectDailySiteStatsHandler(container.SiteStatsCommand))
	return mux
}
