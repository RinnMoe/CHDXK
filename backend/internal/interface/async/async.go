package async

import (
	"github.com/hibiken/asynq"
)

func NewMux() *asynq.ServeMux {
	mux := asynq.NewServeMux()
	registerHandlers(mux)
	return mux
}

func registerHandlers(_ *asynq.ServeMux) {
}
