package task

import "context"

type Task interface {
	Type() string
	Payload() []byte
}

type EnqueueOption any

type Enqueuer interface {
	Enqueue(ctx context.Context, task Task, opts ...EnqueueOption) error
}

var enqueuer Enqueuer

func SetEnqueuer(e Enqueuer) {
	enqueuer = e
}

func SetEnqueuerForTest(e Enqueuer) Enqueuer {
	prev := enqueuer
	enqueuer = e
	return prev
}

func Enqueue(ctx context.Context, t Task, opts ...EnqueueOption) error {
	if enqueuer == nil {
		return nil
	}
	return enqueuer.Enqueue(ctx, t, opts...)
}
