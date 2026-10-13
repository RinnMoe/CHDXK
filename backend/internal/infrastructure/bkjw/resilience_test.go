//go:build test

package bkjw

import (
	"context"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"jcourse/internal/domain/course"
)

func TestPersonalScriptWhitespaceCommentsAndQuotedExamples(t *testing.T) {
	for _, script := range []string{
		`var table0 = new CourseTable (); var a = new  TaskActivity ('','', '101(CS)', '程序设计', '', '', '01');`,
		`/* new TaskActivity('','','invalid','example') */ var table0 = new CourseTable(); var x="new TaskActivity('','','bad','fake')"; var a = new /* annotation */ window.TaskActivity (teachers.join(','), '', '101(CS)', '\u7a0b\u5e8f设计', '', '', '01');`,
	} {
		rows, err := parsePersonal([]byte("<script>" + script + "</script>"))
		if err != nil || len(rows) != 1 || rows[0].SourceLessonID != "101" || rows[0].Name != "程序设计" {
			t.Fatalf("rows=%v err=%v", rows, err)
		}
	}
	for _, page := range []string{
		`<script>var note="new CourseTable()";</script>`,
		`<script>/* new CourseTable() */</script>`,
		`<script>new CourseTable(); new TaskActivity('', '', fetch('x'), 'bad')</script>`,
	} {
		if _, err := parsePersonal([]byte(page)); err == nil {
			t.Fatal("invalid/unknown response accepted as empty")
		}
	}
}

func TestHTTPRejectsPermanentErrorsAndHonorsRetryAfter(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(404) }))
	defer server.Close()
	c := &connection{ctx: context.Background(), client: &http.Client{Transport: mappedTransport{server.URL}}, retryDelay: func(int) time.Duration { return 0 }}
	err := c.retry(func() error { _, _, e := c.eams("GET", "/eams/missing.action", nil); return e }, func(int) {})
	if err != codeError("UPSTREAM_REJECTED") || calls.Load() != 1 {
		t.Fatal(err, calls.Load())
	}
	now := time.Now().UTC().Truncate(time.Second)
	if retryAfter("120", now) != now.Add(120*time.Second) || retryAfter(now.Add(30*time.Second).Format(http.TimeFormat), now) != now.Add(30*time.Second) {
		t.Fatal("Retry-After ignored")
	}
	if !retryAfter("bad", now).IsZero() {
		t.Fatal("bad Retry-After accepted")
	}
}

func TestEAMSRenewsExistingCASSessionWithoutNewQRCode(t *testing.T) {
	var queries atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Host + r.URL.Path {
		case "ids.chd.edu.cn/authserver/login":
			http.Redirect(w, r, eamsOrigin+"/eams/home.action", 302)
		case "bkjw.chd.edu.cn/eams/home.action":
			http.SetCookie(w, &http.Cookie{Name: "JSESSIONID", Value: "renewed-test", Path: "/"})
			fmt.Fprint(w, "<html>教务首页</html>")
		default:
			queries.Add(1)
			if _, err := r.Cookie("JSESSIONID"); err != nil {
				w.WriteHeader(401)
				return
			}
			w.Write(fixture(t, "detail"))
		}
	}))
	defer server.Close()
	jar, _ := cookiejar.New(nil)
	c := &connection{ctx: context.Background(), client: &http.Client{Jar: jar, Transport: mappedTransport{server.URL}, CheckRedirect: checkRedirect}}
	data, _, err := c.eams("GET", "/eams/stdSyllabus!info.action", nil)
	if err != nil || !strings.Contains(string(data), "CS101") || queries.Load() != 2 {
		t.Fatal(err, queries.Load())
	}
}

func TestDetailFailureDoesNotPreventOtherDetailsAndIsRetried(t *testing.T) {
	listing, detail := fixture(t, "listing"), fixture(t, "detail")
	var first, second atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "search") {
			w.Write(listing)
			return
		}
		if r.URL.Query().Get("lesson.id") == "101" {
			if first.Add(1) <= 3 {
				w.WriteHeader(500)
				return
			}
			w.Write(detail)
			return
		}
		second.Add(1)
		fmt.Fprint(w, strings.ReplaceAll(strings.ReplaceAll(string(detail), "CS101", "MA201"), ".01", ".02"))
	}))
	defer server.Close()
	c := &connection{ctx: context.Background(), client: &http.Client{Transport: mappedTransport{server.URL}}, cache: &checkpointStore{root: t.TempDir()}, events: make(chan course.SourceEvent, 40), retryDelay: func(int) time.Duration { return 0 }}
	event, err := c.adminArchive(semester{ID: "101", Name: "2025-2026-1"})
	if err != nil || event.Type != "archive" || first.Load() != 4 || second.Load() != 1 || c.failedDetails != 0 {
		t.Fatal(err, event.Type, first.Load(), second.Load())
	}
}

func TestAdaptivePacingRecoversAndAggregateWaitIsCancellable(t *testing.T) {
	p := &requestPacer{interval: 600 * time.Millisecond, adaptive: true}
	for i := 0; i < 120; i++ {
		p.success()
	}
	if p.interval != 150*time.Millisecond {
		t.Fatal(p.interval)
	}
	p.slowDown()
	if p.interval < time.Second {
		t.Fatal("did not slow down")
	}
	b := NewRequestBudget()
	b.next = time.Now().Add(time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if release, err := b.begin(ctx); err == nil {
		release()
		t.Fatal("cancelled waiter admitted")
	}
	if len(b.inflight) != 0 {
		t.Fatal("cancelled waiter leaked slot")
	}
	b.observe(true, time.Now().Add(time.Minute))
	if b.interval < time.Second || time.Until(b.next) < 50*time.Second {
		t.Fatal("shared cooldown missing")
	}
}

func TestQRPreparationRetriesTransientFailureOnly(t *testing.T) {
	var reads atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reads.Add(1) == 1 {
			w.WriteHeader(502)
			return
		}
		fmt.Fprint(w, "<html>QR preparation</html>")
	}))
	defer server.Close()
	c := &connection{ctx: context.Background(), client: &http.Client{Transport: mappedTransport{server.URL}}, retryDelay: func(int) time.Duration { return 0 }}
	data, _, err := c.loginRead(loginURL(), nil, nil)
	if err != nil || !strings.Contains(string(data), "QR preparation") || reads.Load() != 2 {
		t.Fatal(err, reads.Load())
	}
}

func TestMaintenanceDetectionDoesNotRejectCourseContent(t *testing.T) {
	for _, body := range []string{`<script>var note="系统繁忙";</script><table><tr><td>课程代码</td></tr></table>`, `<form><input name="x">系统繁忙</form>`} {
		if isMaintenancePage([]byte(body)) {
			t.Fatal("ordinary content rejected")
		}
	}
	if !isMaintenancePage([]byte(`<html><div>系统繁忙，请稍后重试</div></html>`)) {
		t.Fatal("maintenance notice missed")
	}
}

func TestPermissionDetectionIgnoresSharedScriptsAndCourseDescriptions(t *testing.T) {
	for _, page := range []string{`<script>function denied(){alert("没有权限")}</script><div>课程查询</div>`, `<table><tr><td>课程代码</td><td>CS101</td></tr><tr><td>备注</td><td>没有权限时联系管理员</td></tr></table>`} {
		if isPermissionPage([]byte(page)) {
			t.Fatal("normal course response rejected as permission failure")
		}
	}
	if !isPermissionPage([]byte(`<h3>您没有权限访问此页面</h3>`)) {
		t.Fatal("permission notice missed")
	}
}

func TestTimetableHonorsDeclaredLegacyCharset(t *testing.T) {
	page := `<script>var t=new CourseTable();var a=new TaskActivity('','','101(CS)','程序设计','','','01');</script>`
	encoded, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(page))
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=GBK")
		w.Write(encoded)
	}))
	defer server.Close()
	c := &connection{ctx: context.Background(), client: &http.Client{Transport: mappedTransport{server.URL}}}
	data, _, err := c.eams("GET", "/eams/courseTableForStd!courseTable.action", nil)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := parsePersonal(data)
	if err != nil || len(rows) != 1 || rows[0].Name != "程序设计" {
		t.Fatal(rows, err)
	}
}

func TestPersonalCourseNameBindings(t *testing.T) {
	for _, script := range []string{
		`var courseName="数学"; new TaskActivity("", "", "123", courseName);`,
		`let prefix="高等"; var courseName=prefix+"数学"; new TaskActivity("", "", "123", courseName); courseName="英语"; new TaskActivity("", "", "456", courseName);`,
	} {
		rows, err := parsePersonal([]byte(script))
		if err != nil || len(rows) == 0 || rows[0].Name == "" {
			t.Fatalf("bindings failed: %v", err)
		}
	}
	for _, script := range []string{
		`var courseName="旧值"; courseName=loadName(); new TaskActivity("", "", "123", courseName);`,
		`// var courseName="注释";
new TaskActivity("", "", "123", courseName);`,
		`var example='courseName="伪造";'; new TaskActivity("", "", "123", courseName);`,
	} {
		if _, err := parsePersonal([]byte(script)); err == nil {
			t.Fatal("unsupported binding accepted")
		}
	}
}

func TestPersonalRepeatedInvalidResponsePreservesParseError(t *testing.T) {
	var reads atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		fmt.Fprint(w, `<script>new CourseTable();new TaskActivity('', '', '123', unknownName);</script>`)
	}))
	defer server.Close()
	c := &connection{ctx: context.Background(), client: &http.Client{Transport: mappedTransport{server.URL}}, events: make(chan course.SourceEvent, 20), retryDelay: func(int) time.Duration { return 0 }}
	_, err := c.personalArchive([]byte(entryPage), semester{ID: "101", Name: "2025-2026-1"})
	if errorCode(err, "") != "PERSONAL_PARSE_FAILED" || reads.Load() != 2 {
		t.Fatalf("code=%s requests=%d", errorCode(err, ""), reads.Load())
	}
}

func TestPersonalCourseNameLiteralConcatenation(t *testing.T) {
	page := `<script>var table0=new CourseTable();var a=new TaskActivity(teachers.join(','),names.join(','),'101'+'(CS)', ('程序' + ('设计')) + '（一）', '', '', '01', false, 0, '', null, false);</script>`
	rows, err := parsePersonal([]byte(page))
	if err != nil || len(rows) != 1 || rows[0].Name != "程序设计（一）" {
		t.Fatal(rows, err)
	}
	for _, expr := range []string{`'X'+dangerous()`, `name`, `'X'.toString()`, `('X')+1`, `decodeURIComponent('X')`} {
		if _, err := jsStringExpression(expr); err == nil {
			t.Fatal("nonliteral expression accepted", expr)
		}
	}
}
