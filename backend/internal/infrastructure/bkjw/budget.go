package bkjw

import (
	"context"
	"sync"
	"time"
)

// RequestBudget limits aggregate EAMS traffic without sharing authentication.
// One budget is shared by administrative and personal sources in an API process.
type RequestBudget struct {
	mu       sync.Mutex
	next     time.Time
	interval time.Duration
	stable   int
	inflight chan struct{}
}

func NewRequestBudget() *RequestBudget {
	return &RequestBudget{interval: 500 * time.Millisecond, inflight: make(chan struct{}, 3)}
}

func (b *RequestBudget) begin(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case b.inflight <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	for {
		if err := ctx.Err(); err != nil {
			<-b.inflight
			return nil, err
		}
		b.mu.Lock()
		wait := time.Until(b.next)
		if wait <= 0 {
			b.next = time.Now().Add(b.interval)
			b.mu.Unlock()
			return func() { <-b.inflight }, nil
		}
		b.mu.Unlock()
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			<-b.inflight
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func (b *RequestBudget) observe(limited bool, until time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if limited {
		b.stable = 0
		b.interval = min(max(b.interval*2, time.Second), 10*time.Second)
		b.next = maxTime(b.next, maxTime(until, time.Now().Add(b.interval)))
	} else {
		b.stable++
		if b.stable >= 20 {
			b.stable = 0
			b.interval = max(250*time.Millisecond, b.interval-100*time.Millisecond)
		}
	}
}
