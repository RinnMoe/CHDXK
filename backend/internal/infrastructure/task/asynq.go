package task

import (
	"context"
	"time"

	"github.com/hibiken/asynq"

	domaintask "jcourse/internal/domain/task"
	"jcourse/internal/infrastructure/persistence"
)

type Config struct {
	Concurrency int `mapstructure:"concurrency"`
}

func redisOpt(conf persistence.RedisConfig) asynq.RedisConnOpt {
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

func NewClient(conf persistence.RedisConfig) *asynq.Client {
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

func NewServer(redisConf persistence.RedisConfig, conf Config) *asynq.Server {
	concurrency := conf.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}
	return asynq.NewServer(redisOpt(redisConf), asynq.Config{
		Concurrency: concurrency,
	})
}

func NewScheduler(redisConf persistence.RedisConfig, loc *time.Location) *asynq.Scheduler {
	return asynq.NewScheduler(redisOpt(redisConf), &asynq.SchedulerOpts{Location: loc})
}

func RegisterScheduledTask(s *asynq.Scheduler, cronspec string, t domaintask.Task, opts ...domaintask.EnqueueOption) (string, error) {
	asynqOpts := make([]asynq.Option, 0, len(opts))
	for _, o := range opts {
		if ao, ok := o.(asynq.Option); ok {
			asynqOpts = append(asynqOpts, ao)
		}
	}
	return s.Register(cronspec, asynq.NewTask(t.Type(), t.Payload()), asynqOpts...)
}
