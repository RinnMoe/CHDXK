package application

import (
	"context"
	"sync"

	"jcourse/pkg/apperr"
)

// CollectorQueue runs bounded personal imports; administrative imports are exclusive.
// Waiting tasks contain no upstream request, QR code or login state.
type CollectorQueue struct {
	mu          sync.Mutex
	tickets     []*collectorTicket
	capacity    int
	concurrency int
}

type collectorTicket struct {
	queue     *CollectorQueue
	ready     chan struct{}
	once      sync.Once
	started   bool
	exclusive bool
}

func NewCollectorQueue(capacity int, concurrency ...int) *CollectorQueue {
	n := 1
	if len(concurrency) > 0 {
		n = max(1, min(concurrency[0], capacity))
	}
	return &CollectorQueue{capacity: capacity, concurrency: n}
}

func (q *CollectorQueue) enqueue(exclusive ...bool) (*collectorTicket, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tickets) >= q.capacity {
		return nil, apperr.Conflict("导入队列已满，请稍后重试")
	}
	t := &collectorTicket{queue: q, ready: make(chan struct{})}
	if len(exclusive) > 0 {
		t.exclusive = exclusive[0]
	}
	q.tickets = append(q.tickets, t)
	q.schedule()
	return t, nil
}

// Caller holds q.mu. An exclusive waiter is a FIFO barrier, preventing starvation.
func (q *CollectorQueue) schedule() {
	active := 0
	for _, t := range q.tickets {
		if t.started {
			active++
			if t.exclusive {
				return
			}
		}
	}
	for _, t := range q.tickets {
		if t.started {
			continue
		}
		if active >= q.concurrency || (t.exclusive && active > 0) {
			return
		}
		t.started = true
		close(t.ready)
		active++
		if t.exclusive {
			return
		}
	}
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
	if t.started {
		return 0
	}
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
			q.schedule()
			return
		}
	})
}
