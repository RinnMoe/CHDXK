//go:build test

package bkjw

import (
	"context"
	"fmt"
	"jcourse/internal/domain/course"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCatalogResumesAfterConnectionLost(t *testing.T) {
	store := &checkpointStore{root: t.TempDir()}
	s := semester{ID: "101", Name: "2025-2026-1"}
	listing, detail := fixture(t, "listing"), fixture(t, "detail")
	var listReads, firstReads, secondReads atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "search") {
			listReads.Add(1)
			w.Write(listing)
			return
		}
		if r.URL.Query().Get("lesson.id") == "101" {
			firstReads.Add(1)
			w.Write(detail)
			return
		}
		if secondReads.Add(1) == 1 {
			cancel()
			w.WriteHeader(500)
			return
		}
		fmt.Fprint(w, strings.ReplaceAll(strings.ReplaceAll(string(detail), "CS101", "MA201"), ".01", ".02"))
	}))
	defer server.Close()
	collect := func(ctx context.Context, fresh ...bool) (course.SourceEvent, error, int) {
		events := make(chan course.SourceEvent)
		c := &connection{ctx: ctx, client: &http.Client{Transport: mappedTransport{server.URL}}, cache: store, events: events, retryDelay: func(int) time.Duration { return 0 }}
		c.fresh = len(fresh) > 0 && fresh[0]
		done := make(chan struct{})
		go func() {
			for range events {
			}
			close(done)
		}()
		event, err := c.adminArchive(s)
		close(events)
		<-done
		return event, err, c.restored
	}
	if _, err, _ := collect(ctx); err == nil {
		t.Fatal("expected interrupted request")
	}
	if firstReads.Load() != 1 {
		t.Fatal("first detail was not collected")
	}
	event, err, cached := collect(context.Background())
	if err != nil || event.Type != "archive" || cached != 1 || listReads.Load() != 1 || firstReads.Load() != 1 || secondReads.Load() != 2 {
		t.Fatal("resume repeated completed requests", err, cached, listReads.Load(), firstReads.Load(), secondReads.Load())
	}
	_, err, cached = collect(context.Background())
	if err != nil || cached != 2 || secondReads.Load() != 2 {
		t.Fatal("complete cache not reused", err, cached)
	}
	_, err, cached = collect(context.Background(), true)
	if err != nil || cached != 0 || listReads.Load() != 2 || firstReads.Load() != 2 || secondReads.Load() != 3 {
		t.Fatal("fresh collection reused old cache", err, cached)
	}
	catalog, err := store.load(s)
	if err != nil || catalog == nil {
		t.Fatal("missing catalog")
	}
	catalog.CreatedAt = time.Now().Add(-25 * time.Hour)
	if err := store.saveCatalog(s, catalog); err != nil {
		t.Fatal(err)
	}
	if v, _ := store.load(s); v != nil {
		t.Fatal("expired cache reused")
	}
	if err := os.WriteFile(filepath.Join(store.dir(s), "catalog.json"), []byte("interrupted JSON"), 0600); err != nil {
		t.Fatal(err)
	}
	if v, _ := store.load(s); v != nil {
		t.Fatal("corrupt cache reused")
	}
}

func TestRetrySurvivesExtendedOutageAndCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := &connection{ctx: ctx, retryDelay: func(int) time.Duration { return 0 }}
	calls := 0
	if err := c.retry(func() error {
		calls++
		if calls < 12 {
			return fmt.Errorf("temporary upstream 500")
		}
		return nil
	}, func(int) {
		if c.retryAt == nil {
			t.Error("missing wait feedback")
		}
	}); err != nil || calls != 12 {
		t.Fatal(err, calls)
	}
	c.retryDelay = func(int) time.Duration { return time.Hour }
	go func() { time.Sleep(5 * time.Millisecond); cancel() }()
	if err := c.retry(func() error { return fmt.Errorf("network unavailable") }, func(int) {}); err != context.Canceled {
		t.Fatal("cancel did not interrupt backoff", err)
	}
}

func TestCatalogResumesFromNextListingPage(t *testing.T) {
	store := &checkpointStore{root: t.TempDir()}
	s := semester{ID: "101", Name: "2025-2026-1"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var firstPages, secondPages atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "info.action") {
			w.Write(fixture(t, "detail"))
			return
		}
		r.ParseForm()
		page := integer(r.Form.Get("pageNo"))
		if page == 1 {
			firstPages.Add(1)
		} else if secondPages.Add(1) == 1 {
			cancel()
			w.WriteHeader(500)
			return
		}
		fmt.Fprintf(w, `<table class="gridtable"><tr><th></th><th>课程序号</th><th>课程代码</th><th>课程名称</th><th>课程类别</th><th>教学班</th><th>教师</th></tr><tr><td><input name="lesson.id" value="%d"></td><td>CS101.0%d</td><td>CS101</td><td><a href="/eams/stdSyllabus!info.action?lesson.id=%d">程序设计</a></td><td>通识课程</td><td>班级</td><td>张三</td></tr></table><script>grid.pageInfo(%d,1,2);</script>`, 100+page, page, 100+page, page)
	}))
	defer server.Close()
	collect := func(ctx context.Context) error {
		c := &connection{ctx: ctx, client: &http.Client{Transport: mappedTransport{server.URL}}, cache: store, events: make(chan course.SourceEvent, 20), retryDelay: func(int) time.Duration { return 0 }}
		_, err := c.adminArchive(s)
		return err
	}
	if err := collect(ctx); err == nil {
		t.Fatal("expected interrupted listing")
	}
	if err := collect(context.Background()); err != nil || firstPages.Load() != 1 || secondPages.Load() != 2 {
		t.Fatal("completed listing page fetched again", err, firstPages.Load(), secondPages.Load())
	}
}

func TestCheckpointWriteFailureIsReported(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	store := &checkpointStore{root: path}
	if err := store.saveCatalog(semester{ID: "101", Name: "2025-2026-1"}, &catalogCheckpoint{}); err != codeError("CACHE_WRITE_FAILED") {
		t.Fatal("cache failure hidden", err)
	}
}
