package bkjw

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"jcourse/internal/domain/course"
)

type Config struct {
	Enabled     bool   `mapstructure:"enabled"`
	Node        string `mapstructure:"node"`
	Runner      string `mapstructure:"runner"`
	BrowserPath string `mapstructure:"browser_path"`
}
type Source struct {
	Config   Config
	Personal bool
}

func (s *Source) Connect(ctx context.Context) (course.SourceConnection, error) {
	if !s.Config.Enabled || s.Config.Runner == "" {
		return nil, errors.New("collector not configured")
	}
	file, err := filepath.Abs(s.Config.Runner)
	if err != nil {
		return nil, err
	}
	node := s.Config.Node
	if node == "" {
		node = "node"
	}
	args := []string{file}
	if s.Personal {
		args = append(args, "--personal")
	}
	cmd := exec.Command(node, args...)
	configureCollectorProcess(cmd)
	cmd.Dir = filepath.Dir(file)
	if s.Config.BrowserPath != "" {
		browserPath, err := filepath.Abs(s.Config.BrowserPath)
		if err != nil {
			return nil, err
		}
		cmd.Env = append(os.Environ(), "PLAYWRIGHT_BROWSERS_PATH="+browserPath)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	// stderr is intentionally discarded: upstream URLs can contain CAS credentials.
	cmd.Stderr = io.Discard
	if err = cmd.Start(); err != nil {
		_ = stdin.Close()
		return nil, err
	}
	c := &connection{cmd: cmd, stdin: stdin, events: make(chan course.SourceEvent, 8), done: make(chan struct{}), stop: make(chan struct{})}
	go func() {
		defer close(c.events)
		defer close(c.done)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 4096), 64<<20)
		terminal := false
		protocolFailed := false
		for scanner.Scan() {
			var event course.SourceEvent
			if json.Unmarshal(scanner.Bytes(), &event) != nil {
				protocolFailed = true
				break
			}
			select {
			case c.events <- event:
			case <-c.stop:
			}
			terminal = event.Type == "done" || event.Type == "error"
		}
		if protocolFailed || scanner.Err() != nil {
			select {
			case c.events <- course.SourceEvent{Type: "error", Error: "SOURCE_UNAVAILABLE"}:
			case <-c.stop:
			}
			terminal = true
			c.Close()
			// Drain outstanding output so the child can close gracefully instead of
			// blocking forever on an oversized JSON event.
			_, _ = io.Copy(io.Discard, stdout)
		}
		_ = cmd.Wait()
		// A crashed Node process must not leave Chromium children behind.
		terminateCollectorProcess(cmd)
		if !terminal {
			select {
			case c.events <- course.SourceEvent{Type: "error", Error: "SOURCE_UNAVAILABLE"}:
			case <-c.stop:
			}
		}
	}()
	go func() {
		select {
		case <-ctx.Done():
			c.Close()
		case <-c.done:
		}
	}()
	return c, nil
}

type connection struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	events chan course.SourceEvent
	done   chan struct{}
	stop   chan struct{}
	mu     sync.Mutex
	once   sync.Once
}

func (c *connection) Events() <-chan course.SourceEvent { return c.events }
func (c *connection) WaitClosed()                       { <-c.done }
func (c *connection) Sync(semesters []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.NewEncoder(c.stdin).Encode(struct {
		Type      string   `json:"type"`
		Semesters []string `json:"semesters"`
	}{"sync", semesters})
}
func (c *connection) Close() {
	c.once.Do(func() {
		close(c.stop)
		c.mu.Lock()
		_ = json.NewEncoder(c.stdin).Encode(map[string]string{"type": "cancel"})
		_ = c.stdin.Close()
		c.mu.Unlock()
		go func() {
			select {
			case <-c.done:
			case <-time.After(35 * time.Second):
				terminateCollectorProcess(c.cmd)
			}
		}()
	})
}
