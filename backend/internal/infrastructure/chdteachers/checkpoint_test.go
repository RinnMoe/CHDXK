package chdteachers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"jcourse/internal/domain/teacher"
)

func TestTeacherCollectionResumesAcrossClientsAndCanRefresh(t *testing.T) {
	root := t.TempDir()
	calls := map[string]int{}
	fail := true
	transport := profileTransport(func(r *http.Request) (*http.Response, error) {
		calls[r.URL.Path]++
		if r.URL.Path == "/b/list.htm" && fail {
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("missing"))}, nil
		}
		data := `<li>职称：教授</li>`
		if r.URL.Path == "/_wp3services/generalQuery" {
			data = `{"total":2,"pageCount":1,"data":[{"title":"张三","department":"理学院","cnUrl":"/a/list.htm","publishStatus":1},{"title":"李四","department":"理学院","cnUrl":"/b/list.htm","publishStatus":1}]}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(data))}, nil
	})
	newClient := func() *Client {
		return &Client{HTTP: &http.Client{Transport: transport}, CacheDir: root, retryDelay: func(int) time.Duration { return 0 }}
	}
	c := newClient()
	rows, err := c.List(context.Background(), func(int, int) {})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.Detail(context.Background(), rows[0]); err != nil {
		t.Fatal(err)
	}
	if _, err = c.Detail(context.Background(), rows[1]); err == nil {
		t.Fatal("expected failed second page")
	}
	fail = false
	c = newClient() // Process-local state is deliberately discarded.
	cached := 0
	ctx := c.CollectionContext(context.Background(), false, func(p teacher.FetchProgress) {
		if p.Cached {
			cached++
		}
	})
	rows, err = c.List(ctx, func(int, int) {})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range rows {
		if _, err := c.Detail(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	if calls["/_wp3services/generalQuery"] != 1 || calls["/a/list.htm"] != 1 || calls["/b/list.htm"] != 2 || cached != 2 {
		t.Fatal("completed requests repeated", calls, cached)
	}
	ctx = c.CollectionContext(context.Background(), true, nil)
	if _, err := c.Detail(ctx, rows[0]); err != nil || calls["/a/list.htm"] != 2 {
		t.Fatal("fresh request reused cache", err)
	}
}

func TestTeacherRetrySurvivesOutageAndCancellation(t *testing.T) {
	calls := 0
	c := &Client{HTTP: &http.Client{Transport: profileTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls < 12 {
			return nil, fmt.Errorf("temporary network failure")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`<li>职称：教授</li>`))}, nil
	})}, retryDelay: func(int) time.Duration { return 0 }}
	if _, err := c.request(context.Background(), origin+"/a/list.htm", nil); err != nil || calls != 12 {
		t.Fatal(err, calls)
	}
	calls = 0
	c.retryDelay = func(int) time.Duration { return time.Hour }
	ctx, cancel := context.WithCancel(context.Background())
	waiting := 0
	ctx = c.CollectionContext(ctx, false, func(p teacher.FetchProgress) {
		waiting += p.WaitDelta
		if p.Retry {
			cancel()
		}
	})
	if _, err := c.request(ctx, origin+"/a/list.htm", nil); err != context.Canceled || waiting != 0 {
		t.Fatal("cancel did not stop retry", err, waiting)
	}
}

func TestMalformedTeacherResponseIsNotCached(t *testing.T) {
	calls := 0
	c := &Client{CacheDir: t.TempDir(), HTTP: &http.Client{Transport: profileTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		data := `<html>系统维护</html>`
		if calls > 8 {
			data = `<li>职称：教授</li>`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(data))}, nil
	})}, retryDelay: func(int) time.Duration { return 0 }}
	if _, err := c.request(context.Background(), origin+"/a/list.htm", nil); err == nil {
		t.Fatal("maintenance page accepted")
	}
	if _, err := c.request(context.Background(), origin+"/a/list.htm", nil); err != nil || calls != 9 {
		t.Fatal("failed response cached", err, calls)
	}
	if err := validResponse([]byte(`{"total":0,"pageCount":0,"data":null}`), url.Values{}); err == nil {
		t.Fatal("incomplete catalog accepted")
	}
}

func TestDuplicateCachedPageIsDiscardedAndRetried(t *testing.T) {
	root := t.TempDir()
	reads := map[string]int{}
	var secondBody url.Values
	stale := false
	pageData := func(page string) string {
		return fmt.Sprintf(`{"total":3,"pageCount":3,"data":[{"title":"教师%s","department":"理学院","cnUrl":"/%s/list.htm","publishStatus":1}]}`, page, page)
	}
	transport := profileTransport(func(r *http.Request) (*http.Response, error) {
		r.ParseForm()
		page := r.Form.Get("pageIndex")
		reads[page]++
		if page == "2" {
			secondBody = r.PostForm
		}
		data := pageData(page)
		if stale && page == "2" && reads[page] == 2 {
			data = pageData("1")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(data))}, nil
	})
	c := &Client{HTTP: &http.Client{Transport: transport}, CacheDir: root, retryDelay: func(int) time.Duration { return 0 }}
	if rows, err := c.List(context.Background(), func(int, int) {}); err != nil || len(rows) != 3 {
		t.Fatal(rows, err)
	}
	// Reproduce the stale page persisted by the previous collector version.
	if err := c.saveCache(origin+"/_wp3services/generalQuery?queryObj=teacherHome", secondBody, []byte(pageData("1"))); err != nil {
		t.Fatal(err)
	}
	stale = true
	cached, retries := 0, 0
	ctx := c.CollectionContext(context.Background(), false, func(p teacher.FetchProgress) {
		if p.Cached {
			cached++
		}
		if p.Retry {
			retries++
		}
	})
	if rows, err := c.List(ctx, func(int, int) {}); err != nil || len(rows) != 3 {
		t.Fatal(rows, err)
	}
	if reads["1"] != 1 || reads["2"] != 3 || reads["3"] != 1 || cached != 2 || retries != 1 {
		t.Fatal("stale page was reused or completed pages refetched", reads, cached, retries)
	}
}
