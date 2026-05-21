package async

import (
	"github.com/hibiken/asynq"

	"jcourse/internal/app"
	domainauth "jcourse/internal/domain/auth"
)

func NewMux(container *app.ServiceContainer) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.Handle(domainauth.TaskTypeClearExpiredSuspension, newClearExpiredSuspensionHandler(container.AuthService))
	return mux
}
