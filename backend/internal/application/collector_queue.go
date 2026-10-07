package application

import (
	"context"
	"sync"

	"jcourse/pkg/apperr"
)

// CollectorQueue shares one browser slot across administrative and personal imports.
// Waiting tasks contain no browser, QR code or upstream login state.
type CollectorQueue struct {
	mu       sync.Mutex
	tickets  []*collectorTicket
	capacity int
}

type collectorTicket struct {
	queue *CollectorQueue
	ready chan struct{}
	once  sync.Once
}

func NewCollectorQueue(capacity int) *CollectorQueue {
	return &CollectorQueue{capacity: capacity}
}

func (q *CollectorQueue) enqueue() (*collectorTicket, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tickets) >= q.capacity {
		return nil, apperr.Conflict("导入队列已满，请稍后重试")
	}
	t := &collectorTicket{queue: q, ready: make(chan struct{})}
	q.tickets = append(q.tickets, t)
	if len(q.tickets) == 1 {
		close(t.ready)
	}
	return t, nil
}

func (t *collectorTicket) wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.ready:
		return ctx.Err()
	}
}

func (t *collectorTicket) position() int {
	t.queue.mu.Lock()
	defer t.queue.mu.Unlock()
	for i, item := range t.queue.tickets {
		if item == t {
			return i
		}
	}
	return 0
}

func (t *collectorTicket) release() {
	t.once.Do(func() {
		q := t.queue
		q.mu.Lock()
		defer q.mu.Unlock()
		for i, item := range q.tickets {
			if item != t {
				continue
			}
			q.tickets = append(q.tickets[:i], q.tickets[i+1:]...)
			if i == 0 && len(q.tickets) > 0 {
				close(q.tickets[0].ready)
			}
			return
		}
	})
}
