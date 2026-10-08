//go:build test

package application

import (
	"context"
	"sync"
	"testing"
	"time"

	"jcourse/internal/domain/course"
)

type queueTestConnection struct {
	testSourceConnection
	closing chan struct{}
	closed  chan struct{}
	once    sync.Once
}

func (c *queueTestConnection) Close()      { c.once.Do(func() { close(c.closing) }) }
func (c *queueTestConnection) WaitClosed() { <-c.closed }

func TestCollectorQueueSerializesAdminAndPersonalAndWaitsForCollectorExit(t *testing.T) {
	queue := NewCollectorQueue(3)
	connected := make(chan *queueTestConnection, 3)
	source := &course.MockCourseSource{ConnectFn: func(context.Context) (course.SourceConnection, error) {
		c := &queueTestConnection{testSourceConnection: testSourceConnection{events: make(chan course.SourceEvent, 4)}, closing: make(chan struct{}), closed: make(chan struct{})}
		connected <- c
		return c, nil
	}}
	admin := &CourseSyncService{source: source, queue: queue}
	personal := NewEnrollmentSyncService(source, nil, queue)
	first, err := admin.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	c := <-connected
	second, err := personal.Start(2)
	if err != nil {
		t.Fatal(err)
	}
	third, err := personal.Start(3)
	if err != nil {
		t.Fatal(err)
	}
	if second.State != "queued" || second.QueuePosition != 1 || third.QueuePosition != 2 {
		t.Fatalf("wrong queue positions: %+v %+v", second, third)
	}
	if _, err := personal.Start(4); err == nil {
		t.Fatal("unbounded queue accepted")
	}
	if _, err := personal.Start(2); err == nil {
		t.Fatal("duplicate user queued twice")
	}
	if _, err := personal.Cancel(2, third.ID); err == nil {
		t.Fatal("cross-user cancellation accepted")
	}
	if _, err := personal.Cancel(2, second.ID); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		status, _ := personal.Status(3, third.ID)
		if status.QueuePosition == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("cancelled ticket retained")
		}
		time.Sleep(time.Millisecond)
	}
	if _, err := admin.Cancel(1, first.ID); err != nil {
		t.Fatal(err)
	}
	select {
	case <-c.closing:
	case <-time.After(time.Second):
		t.Fatal("collector not closed")
	}
	select {
	case <-connected:
		t.Fatal("next collector started before previous exit")
	case <-time.After(20 * time.Millisecond):
	}
	close(c.closed)
	var next *queueTestConnection
	select {
	case next = <-connected:
	case <-time.After(time.Second):
		t.Fatal("queue stalled after collector exit")
	}
	_, _ = personal.Cancel(3, third.ID)
	close(next.closed)
}

func TestCollectorQueueTimeoutAndConnectFailureReleaseSlot(t *testing.T) {
	q := NewCollectorQueue(2)
	first, _ := q.enqueue()
	second, _ := q.enqueue()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if err := second.wait(ctx); err == nil {
		t.Fatal("waiting ignored deadline")
	}
	second.release()
	first.release()
	source := &course.MockCourseSource{ConnectFn: func(context.Context) (course.SourceConnection, error) { return nil, context.DeadlineExceeded }}
	s := &CourseSyncService{source: source, queue: q}
	job, err := s.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	waitSync(t, s, job.ID, "failed")
	deadline := time.Now().Add(time.Second)
	for {
		q.mu.Lock()
		remaining := len(q.tickets)
		q.mu.Unlock()
		if remaining == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("failed connect leaked slot")
		}
		time.Sleep(time.Millisecond)
	}
}
