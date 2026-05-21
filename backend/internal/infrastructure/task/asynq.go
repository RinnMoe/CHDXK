package task

import (
	"context"

	"github.com/hibiken/asynq"

	"jcourse/config"
	domaintask "jcourse/internal/domain/task"
)

func redisOpt(conf config.RedisConfig) asynq.RedisConnOpt {
	return asynq.RedisClientOpt{
		Addr:     conf.Addr,
		Username: conf.Username,
		Password: conf.Password,
		DB:       conf.DB,
	}
}

type asynqEnqueuer struct {
	client *asynq.Client
}

func NewClient(conf config.RedisConfig) *asynq.Client {
	return asynq.NewClient(redisOpt(conf))
}

func NewEnqueuer(client *asynq.Client) domaintask.Enqueuer {
	return &asynqEnqueuer{client: client}
}

func (e *asynqEnqueuer) Enqueue(ctx context.Context, t domaintask.Task, opts ...domaintask.EnqueueOption) error {
	asynqOpts := make([]asynq.Option, 0, len(opts))
	for _, o := range opts {
		if ao, ok := o.(asynq.Option); ok {
			asynqOpts = append(asynqOpts, ao)
		}
	}
	_, err := e.client.EnqueueContext(ctx, asynq.NewTask(t.Type(), t.Payload()), asynqOpts...)
	return err
}

func NewServer(conf config.AppConfig) *asynq.Server {
	concurrency := conf.Asynq.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}
	return asynq.NewServer(redisOpt(conf.Redis), asynq.Config{
		Concurrency: concurrency,
	})
}
