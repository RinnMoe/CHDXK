package bkjw

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"jcourse/internal/domain/course"
)

type Config struct {
	Enabled bool `mapstructure:"enabled"`
}
type Source struct {
	Config   Config
	Personal bool
}

func (s *Source) Connect(ctx context.Context) (course.SourceConnection, error) {
	if !s.Config.Enabled {
		return nil, errors.New("collector disabled")
	}
	return s.connect(ctx, nil), nil
}

// Each connection owns an in-memory cookie jar. No browser or session file is used.
func (s *Source) connect(parent context.Context, transport http.RoundTripper) *connection {
	ctx, cancel := context.WithCancel(parent)
	jar, _ := cookiejar.New(nil)
	if transport == nil {
		transport = &http.Transport{Proxy: http.ProxyFromEnvironment, MaxIdleConns: 4, MaxIdleConnsPerHost: 2, IdleConnTimeout: 30 * time.Second, TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: 30 * time.Second}
	}
	client := &http.Client{Jar: jar, Transport: transport, Timeout: 30 * time.Second, CheckRedirect: checkRedirect}
	c := &connection{ctx: ctx, cancel: cancel, client: client, events: make(chan course.SourceEvent), commands: make(chan []string, 1), done: make(chan struct{}), personal: s.Personal}
	go c.run()
	return c
}

type connection struct {
	ctx       context.Context
	cancel    context.CancelFunc
	client    *http.Client
	events    chan course.SourceEvent
	commands  chan []string
	done      chan struct{}
	personal  bool
	once      sync.Once
	mu        sync.Mutex
	submitted bool
}

func (c *connection) Events() <-chan course.SourceEvent { return c.events }
func (c *connection) WaitClosed()                       { <-c.done }
func (c *connection) Close()                            { c.once.Do(c.cancel) }
func (c *connection) Sync(semesters []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ctx.Err(); err != nil {
		return err
	}
	if c.submitted || len(semesters) == 0 {
		return errors.New("selection already submitted or empty")
	}
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.commands <- append([]string{}, semesters...):
		c.submitted = true
		return nil
	}
}
func (c *connection) emit(event course.SourceEvent) bool {
	select {
	case <-c.ctx.Done():
		return false
	case c.events <- event:
		return true
	}
}
func (c *connection) run() {
	defer close(c.done)
	defer close(c.events)
	defer c.cancel()
	defer func() { c.client.CloseIdleConnections(); c.client.Jar = nil }()
	stage := "QR_LOGIN_FAILED"
	fail := func(err error) {
		if c.ctx.Err() == nil {
			c.emit(course.SourceEvent{Type: "error", Error: errorCode(err, stage)})
		}
	}
	if err := c.login(); err != nil {
		fail(err)
		return
	}
	c.emit(course.SourceEvent{Type: "discovering"})
	stage = "SEMESTER_DISCOVERY_FAILED"
	entry := "/eams/stdSyllabus.action"
	if c.personal {
		entry = "/eams/courseTableForStd.action"
	}
	page, _, err := c.eams("GET", entry, nil)
	if err != nil {
		fail(err)
		return
	}
	options, err := c.semesters(page)
	if err != nil {
		fail(err)
		return
	}
	ids := make([]string, 0, len(options))
	for _, s := range options {
		ids = append(ids, s.Name)
	}
	if !c.emit(course.SourceEvent{Type: "ready", Semesters: ids}) {
		return
	}
	var selected []string
	select {
	case <-c.ctx.Done():
		return
	case selected = <-c.commands:
	}
	chosen := make([]semester, 0, len(selected))
	seen := map[string]bool{}
	for _, name := range selected {
		found := false
		for _, s := range options {
			if s.Name == name && !seen[name] {
				chosen = append(chosen, s)
				seen[name] = true
				found = true
				break
			}
		}
		if !found {
			fail(codeError("INVALID_SEMESTER"))
			return
		}
	}
	stage = "COURSE_COLLECTION_FAILED"
	for _, s := range chosen {
		if !c.emit(course.SourceEvent{Type: "collecting", Semester: s.Name}) {
			return
		}
		var event course.SourceEvent
		if c.personal {
			event, err = c.personalArchive(page, s)
		} else {
			event, err = c.adminArchive(s)
		}
		if err != nil {
			fail(err)
			return
		}
		if !c.emit(event) {
			return
		}
	}
	c.emit(course.SourceEvent{Type: "done"})
}
func marshal(value any) json.RawMessage { data, _ := json.Marshal(value); return data }

type codeError string

func (e codeError) Error() string { return string(e) }
func errorCode(err error, fallback string) string {
	var e codeError
	if errors.As(err, &e) {
		return string(e)
	}
	return fallback
}
