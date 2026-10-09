//go:build test

package bkjw

import (
	"context"
	"encoding/json"
	"fmt"
	"jcourse/internal/domain/course"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type mappedTransport struct{ origin string }

func (m mappedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	copy := req.Clone(req.Context())
	u, _ := url.Parse(m.origin)
	copy.URL.Scheme = u.Scheme
	copy.URL.Host = u.Host
	copy.Host = req.URL.Host
	resp, err := http.DefaultTransport.RoundTrip(copy)
	if resp != nil {
		resp.Request = req
	}
	return resp, err
}
func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name + ".html")
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func next(t *testing.T, c *connection, kind string) course.SourceEvent {
	t.Helper()
	for {
		select {
		case e, ok := <-c.Events():
			if !ok {
				t.Fatal("closed before", kind)
			}
			if e.Type == "error" {
				t.Fatal("collector error", e.Error)
			}
			if e.Type == kind {
				return e
			}
		case <-time.After(12 * time.Second):
			t.Fatal("timeout waiting", kind)
		}
	}
}

const entryPage = `<input class="calendar-text" id="querySemester"><script>semesterCalendar({value:101});function searchTable(){if(jQuery("#courseTableType").val()=="std"){bg.form.addInput(form,"ids","77");}else{bg.form.addInput(form,"ids","88");}}</script>`
const calendar = `{semesters:{y0:[{id:101,schoolYear:2025-2026,name:1},{id:102,schoolYear:2025-2026,name:2}],semesterId:101}}`

func TestHTTPCollectorEndToEnd(t *testing.T) {
	for _, personal := range []bool{true, false} {
		t.Run(fmt.Sprint(personal), func(t *testing.T) {
			var tables atomic.Int32
			var callbacks atomic.Int32
			var business atomic.Int32
			listing := fixture(t, "listing")
			detail := fixture(t, "detail")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Host + r.URL.Path {
				case "ids.chd.edu.cn/authserver/login":
					if cookie, e := r.Cookie("TGC"); e == nil && cookie.Value == "test-session" {
						http.Redirect(w, r, eamsOrigin+"/eams/home.action", 302)
						return
					}
					fmt.Fprint(w, `<iframe id="iframe" src="/authserver/wechat"></iframe>`)
				case "ids.chd.edu.cn/authserver/wechat":
					http.Redirect(w, r, wxOrigin+"/connect/qrconnect", 302)
				case "open.weixin.qq.com/connect/qrconnect":
					fmt.Fprint(w, `<img class="web_qrcode_img" src="/qr.png"><script>var t="https://ids.chd.edu.cn/authserver/callback?state=sample";var url="qrconnect?uuid=fixture-uuid-1234";</script>`)
				case "open.weixin.qq.com/qr.png":
					w.Write([]byte("fixture-image"))
				case "lp.open.weixin.qq.com/connect/l/qrconnect":
					if tables.Load() == 0 {
						tables.Add(1)
						fmt.Fprint(w, `window.wx_errcode=404;window.wx_code='';`)
					} else {
						fmt.Fprint(w, `window.wx_errcode=405;window.wx_code='fake-code';`)
					}
				case "ids.chd.edu.cn/authserver/callback":
					callbacks.Add(1)
					if r.URL.Query().Get("code") != "fake-code" {
						t.Error("missing callback code")
					}
					http.SetCookie(w, &http.Cookie{Name: "TGC", Value: "test-session", Path: "/"})
					fmt.Fprint(w, `<form method="post" action="http://bkjw.chd.edu.cn/eams/login.action"><input name="ticket" value="test-ticket"></form><script>document.forms[0].submit();</script>`)
				case "bkjw.chd.edu.cn/eams/login.action":
					r.ParseForm()
					if r.Method != "POST" || r.Form.Get("ticket") != "test-ticket" {
						t.Error("lost automatic form")
					}
					http.SetCookie(w, &http.Cookie{Name: "JSESSIONID", Value: "eams-test", Path: "/eams/"})
					http.Redirect(w, r, "/eams/home.action", 302)
				case "bkjw.chd.edu.cn/eams/home.action":
					if _, e := r.Cookie("JSESSIONID"); e != nil {
						t.Error("lost business cookie")
					}
					business.Add(1)
					fmt.Fprint(w, "<html>教务首页</html>")
				case "bkjw.chd.edu.cn/eams/courseTableForStd.action", "bkjw.chd.edu.cn/eams/stdSyllabus.action":
					fmt.Fprint(w, entryPage)
				case "bkjw.chd.edu.cn/eams/dataQuery.action":
					r.ParseForm()
					if r.Form.Get("dataType") != "semesterCalendar" || r.Form.Get("tagId") != "querySemester" {
						t.Error("calendar parameters")
					}
					fmt.Fprint(w, calendar)
				case "bkjw.chd.edu.cn/eams/courseTableForStd!courseTable.action":
					r.ParseForm()
					if r.Form.Get("ids") != "77" || r.Form.Get("setting.kind") != "std" {
						t.Error("wrong student branch", r.Form)
					}
					if r.Form.Get("semester.id") == "101" {
						fmt.Fprint(w, `<script>var table0=new CourseTable();var activity=new TaskActivity(teachers.join(','),names.join(','),'101(CS101.01)','程序设计','r','教室','0101');table0.addActivity(activity);var activity=new TaskActivity('', '', '101(CS101.01)', '程序设计', '', '', '01');</script>`)
					} else {
						fmt.Fprint(w, `<script>var table0=new CourseTable();</script>`)
					}
				case "bkjw.chd.edu.cn/eams/stdSyllabus!search.action":
					w.Write(listing)
				case "bkjw.chd.edu.cn/eams/stdSyllabus!info.action":
					if r.URL.Query().Get("lesson.id") == "102" {
						data := strings.ReplaceAll(string(detail), "CS101", "MA201")
						data = strings.ReplaceAll(data, "程序设计", "高等数学（下）")
						fmt.Fprint(w, data)
					} else {
						w.Write(detail)
					}
				default:
					t.Error("unexpected path", r.Host, r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			c := (&Source{Personal: personal}).connect(context.Background(), mappedTransport{server.URL})
			defer c.Close()
			if e := next(t, c, "qr"); !strings.HasPrefix(e.QR, "data:image/png;base64,") {
				t.Fatal("qr")
			}
			next(t, c, "scanned")
			next(t, c, "authenticating")
			ready := next(t, c, "ready")
			if len(ready.Semesters) != 2 {
				t.Fatal("incomplete catalog")
			}
			selected := ready.Semesters
			if !personal {
				selected = selected[:1]
			}
			if err := c.Sync(selected); err != nil {
				t.Fatal(err)
			}
			if err := c.Sync(selected); err == nil {
				t.Fatal("accepted duplicate run")
			}
			for i, s := range selected {
				e := next(t, c, "archive")
				if e.Semester != s {
					t.Fatal("semester mismatch")
				}
				if personal {
					var d struct {
						Courses []course.EnrollmentSourceCourse `json:"courses"`
					}
					json.Unmarshal(e.Document, &d)
					if i == 0 && (len(d.Courses) != 1 || d.Courses[0].SourceLessonID != "101") {
						t.Fatal("personal identity", d)
					}
					if i == 1 && len(d.Courses) != 0 {
						t.Fatal("stale previous semester")
					}
				} else {
					var d map[string]any
					json.Unmarshal(e.Document, &d)
					if len(d["lessons"].([]any)) != 2 {
						t.Fatal("lessons")
					}
					var report map[string]any
					json.Unmarshal(e.Report, &report)
					if report["completed"] != true || report["detail_success_count"] != float64(2) {
						t.Fatal("report")
					}
				}
			}
			next(t, c, "done")
			c.WaitClosed()
			if callbacks.Load() != 1 || business.Load() < 1 {
				t.Fatal("unverified business session")
			}
		})
	}
}

func TestLoginThroughCASAccountPortal(t *testing.T) {
	var portal, business atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Host + r.URL.Path {
		case "ids.chd.edu.cn/authserver/callback":
			http.SetCookie(w, &http.Cookie{Name: "TGC", Value: "test-session", Path: "/", Secure: true})
			http.Redirect(w, r, casOrigin+":443"+casLandingPath, http.StatusFound)
		case "ids.chd.edu.cn:443" + casLandingPath:
			portal.Add(1)
			fmt.Fprint(w, `<script>function logout(){location.href='https://evil.invalid/logout';}</script>`)
		case "ids.chd.edu.cn/authserver/login":
			cookie, err := r.Cookie("TGC")
			if err != nil || cookie.Value != "test-session" || r.URL.Query().Get("service") != eamsOrigin+"/eams/login.action" {
				t.Error("CAS service login lost the authenticated session")
			}
			http.Redirect(w, r, eamsOrigin+"/eams/login.action?ticket=test-ticket", http.StatusFound)
		case "bkjw.chd.edu.cn/eams/login.action":
			if r.URL.Query().Get("ticket") != "test-ticket" {
				t.Error("missing service ticket")
			}
			http.SetCookie(w, &http.Cookie{Name: "JSESSIONID", Value: "eams-session", Path: "/eams/"})
			http.Redirect(w, r, "/eams/home.action", http.StatusFound)
		case "bkjw.chd.edu.cn/eams/home.action":
			if _, err := r.Cookie("JSESSIONID"); err != nil {
				t.Error("business session missing")
			}
			business.Add(1)
			fmt.Fprint(w, "<html>教务首页</html>")
		default:
			t.Error("unexpected portal script navigation", r.Host, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	jar, _ := cookiejar.New(nil)
	c := &connection{ctx: context.Background(), client: &http.Client{Jar: jar, Transport: mappedTransport{server.URL}, CheckRedirect: checkRedirect}}
	defer c.client.CloseIdleConnections()
	if err := c.completeLogin(casOrigin + "/authserver/callback?code=fake-code"); err != nil {
		t.Fatal(err)
	}
	if portal.Load() != 1 || business.Load() < 2 {
		t.Fatal("portal redirect or authenticated business verification missing")
	}
	for _, raw := range []string{casOrigin + ":443/authserver/login", casOrigin + ":443" + casLandingPath, "http://bkjw.chd.edu.cn:80/eams/home.action", "https://bkjw.chd.edu.cn:443/eams/home.action"} {
		if !allowedURL(raw) {
			t.Errorf("rejected trusted address with default port: %s", raw)
		}
	}
	for _, raw := range []string{casOrigin + "/personalInfo/other", "http://ids.chd.edu.cn" + casLandingPath, casOrigin + ":8443" + casLandingPath, "https://evil.invalid" + casLandingPath} {
		if allowedURL(raw) {
			t.Errorf("accepted unintended portal address: %s", raw)
		}
	}
}

func TestAdminArchiveRetriesStalePagination(t *testing.T) {
	var secondPageRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/eams/stdSyllabus!info.action" {
			w.Write(fixture(t, "detail"))
			return
		}
		r.ParseForm()
		page := integer(r.Form.Get("pageNo"))
		if page == 2 && secondPageRequests.Add(1) == 1 {
			page = 1 // A transient stale response must not append duplicate rows.
		}
		fmt.Fprintf(w, `<table class="gridtable"><tr><th></th><th>课程序号</th><th>课程代码</th><th>课程名称</th><th>课程类别</th><th>教学班</th><th>教师</th></tr><tr><td><input name="lesson.id" value="%d"></td><td>CS101.0%d</td><td>CS101</td><td><a href="/eams/stdSyllabus!info.action?lesson.id=%d">程序设计</a></td><td>通识课程</td><td>班级</td><td>张三</td></tr></table><script>grid.pageInfo(%d,1,2);</script>`, 100+page, page, 100+page, page)
	}))
	defer server.Close()
	c := &connection{ctx: context.Background(), client: &http.Client{Transport: mappedTransport{server.URL}, CheckRedirect: checkRedirect}, events: make(chan course.SourceEvent, 20)}
	defer c.client.CloseIdleConnections()
	event, err := c.adminArchive(semester{ID: "101", Name: "2025-2026-1"})
	if err != nil {
		t.Fatal(err)
	}
	var archive struct {
		Lessons []struct {
			ID string `json:"id"`
		} `json:"lessons"`
	}
	if err := json.Unmarshal(event.Document, &archive); err != nil {
		t.Fatal(err)
	}
	if secondPageRequests.Load() != 2 || len(archive.Lessons) != 2 || archive.Lessons[0].ID == archive.Lessons[1].ID {
		t.Fatal("stale page was not retried before saving complete archive", archive)
	}
}

func TestParsersRejectIncompleteAndUnsafeData(t *testing.T) {
	good := fixture(t, "listing")
	got, err := parseListing(good)
	if err != nil || got.Total != 2 || got.Lessons[0]["summary"].(map[string]any)["teachers"].([]string)[1] != "李四" {
		t.Fatal(got, err)
	}
	for _, data := range []string{"<html>错误</html>", strings.ReplaceAll(string(good), "课程代码", "unknown"), strings.ReplaceAll(string(good), "/eams/stdSyllabus!info.action", "https://evil.invalid/eams/stdSyllabus!info.action")} {
		if _, err := parseListing([]byte(data)); err == nil {
			t.Fatal("accepted malformed listing")
		}
	}
	detail, err := parseDetail(fixture(t, "detail"))
	if err != nil || detail["basic"].(map[string]any)["department"] != "信息学院" {
		t.Fatal(detail, err)
	}
	if _, err := parseSemesters([]byte(strings.Replace(calendar, "id:102", "id:101", 1))); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{`<html>系统繁忙</html>`, `new TaskActivity('', '', fetch('secret'), 'X')`, `new TaskActivity('', '', '101', 'X' + dangerous())`, `{"activities":null}`} {
		if _, err := parsePersonal([]byte(raw)); err == nil {
			t.Fatal("accepted invalid personal response", raw)
		}
	}
	courses, err := parsePersonal([]byte(`{"activities":[[{"courseId":"101(CS101)","courseName":"A","lessonNo":"L1"},{"courseId":"101(CS101)","courseName":"A","lessonNo":"L1"}],[]]}`))
	if err != nil || len(courses) != 1 {
		t.Fatal(courses, err)
	}
	value, err := jsString(`'课程\u540d\x41\'称'`)
	if err != nil || value != "课程名A'称" {
		t.Fatal(value, err)
	}
}
func TestURLPolicy(t *testing.T) {
	for _, raw := range []string{"http://ids.chd.edu.cn/authserver/login", "https://evil.invalid/", "http://bkjw.chd.edu.cn:90/eams/home.action", "http://bkjw.chd.edu.cn/logout", "http://user:secret@bkjw.chd.edu.cn/eams/home.action"} {
		if allowedURL(raw) {
			t.Fatal("unsafe URL", raw)
		}
	}
}
func TestCancellationAndRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := (&Source{}).connect(ctx, nil)
	cancel()
	c.WaitClosed()
	if err := c.Sync([]string{"2025-2026-1"}); err == nil {
		t.Fatal("accepted closed connection")
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	c = &connection{ctx: ctx, retryDelay: func(int) time.Duration { return 0 }}
	calls := 0
	retries := 0
	err := c.retry(func() error { calls++; return codeError("AUTH_REQUIRED") }, func(int) { retries++ })
	if err == nil || calls != 3 || retries != 2 {
		t.Fatal("auth confirmation retries", calls, retries)
	}
	calls, retries = 0, 0
	err = c.retry(func() error {
		calls++
		if calls < 4 {
			return fmt.Errorf("temporary 503")
		}
		return nil
	}, func(int) { retries++ })
	if err != nil || retries != 3 {
		t.Fatal("retry", err, retries)
	}
}
func TestRejectCrossHostCallbackAndUnbound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `首次使用微信联合登录`) }))
	defer server.Close()
	jarClient := (&Source{}).connect(context.Background(), mappedTransport{server.URL})
	jarClient.Close()
	jarClient.WaitClosed()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jarClient.ctx = ctx
	_, _, err := jarClient.navigate(casOrigin + "/authserver/callback")
	if err != codeError("WECHAT_UNBOUND") {
		t.Fatal(err)
	}
}
func TestLiveQRCode(t *testing.T) {
	if os.Getenv("BKJW_LIVE_QR") != "1" {
		t.Skip("opt-in network smoke; no phone confirmation needed")
	}
	c, err := (&Source{Config: Config{Enabled: true}}).Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	timer := time.NewTimer(50 * time.Second)
	defer timer.Stop()
	select {
	case e := <-c.Events():
		if e.Type != "qr" || len(e.QR) < 100 {
			t.Fatal("QR acquisition failed", e.Type, e.Error)
		}
		t.Log("real CAS/WeChat QR acquired without browser")
	case <-timer.C:
		t.Fatal("QR timed out")
	}
	c.Close()
	c.(*connection).WaitClosed()
}

func TestEAMSDetectsRateLimitAndSlowsSharedPacer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><div>请不要过快点击</div><a href="javascript:history.back()">[返回前页]</a></html>`)
	}))
	defer server.Close()
	c := &connection{ctx: context.Background(), client: &http.Client{Transport: mappedTransport{server.URL}}, pacer: &requestPacer{interval: 100 * time.Millisecond}}
	defer c.client.CloseIdleConnections()
	if _, _, err := c.eams("GET", "/eams/stdSyllabus!info.action", nil); err != codeError("RATE_LIMITED") {
		t.Fatal("wrong rate limit classification", err)
	}
	if c.pacer.interval < time.Second {
		t.Fatal("pacing was not reduced")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if finish, err := c.pacer.begin(ctx); err == nil {
		finish()
		t.Fatal("cancelled request continued")
	}
}

func TestEAMSReplaysQueryAfterSessionRefresh(t *testing.T) {
	var reads atomic.Int32
	detail := fixture(t, "detail")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/eams/home.action" {
			fmt.Fprint(w, "<html>教务首页</html>")
			return
		}
		if reads.Add(1) == 1 {
			http.SetCookie(w, &http.Cookie{Name: "JSESSIONID", Value: "renewed-fixture", Path: "/"})
			http.Redirect(w, r, eamsOrigin+"/eams/home.action", 302)
			return
		}
		cookie, err := r.Cookie("JSESSIONID")
		if err != nil || cookie.Value != "renewed-fixture" {
			w.WriteHeader(401)
			return
		}
		w.Write(detail)
	}))
	defer server.Close()
	jar, _ := cookiejar.New(nil)
	c := &connection{ctx: context.Background(), client: &http.Client{Jar: jar, Transport: mappedTransport{server.URL}, CheckRedirect: checkRedirect}}
	defer c.client.CloseIdleConnections()
	retries := 0
	err := c.retry(func() error {
		data, _, err := c.eams("GET", "/eams/stdSyllabus!info.action", nil)
		if err != nil {
			return err
		}
		_, err = parseDetail(data)
		return err
	}, func(int) { retries++ })
	if err != nil || reads.Load() != 2 || retries != 1 {
		t.Fatal("query was not replayed after renewal", err, reads.Load(), retries)
	}
}

func TestPacingIncludesDirectLoginRequestAndResponseTime(t *testing.T) {
	var calls atomic.Int32
	var completed time.Time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			time.Sleep(80 * time.Millisecond)
			completed = time.Now()
			fmt.Fprint(w, "login complete")
			return
		}
		if time.Since(completed) < 95*time.Millisecond {
			t.Error("query sent before response cooldown")
		}
		fmt.Fprint(w, "query complete")
	}))
	defer server.Close()
	c := &connection{ctx: context.Background(), client: &http.Client{Transport: mappedTransport{server.URL}}, pacer: &requestPacer{interval: 100 * time.Millisecond}}
	defer c.client.CloseIdleConnections()
	if _, _, err := c.request("GET", eamsOrigin+"/eams/home.action", nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.eams("GET", "/eams/stdSyllabus.action", nil); err != nil {
		t.Fatal(err)
	}
}

func TestRateLimitPageRequiresActualErrorNotice(t *testing.T) {
	for _, page := range []string{
		`<html><script>var message="请不要过快点击";</script><a>[返回前页]</a></html>`,
		`<table><tr><td>课程代码</td><td>CS101</td></tr></table><div>请不要过快点击</div><a>[返回前页]</a>`,
		`<html>请稍后</html>`,
	} {
		if isRateLimitPage([]byte(page)) {
			t.Fatal("normal or unrelated page classified as rate limit")
		}
	}
	if !isRateLimitPage([]byte(`<div class="actionError"><font>请不要过快点击</font></div><a href="javascript:history.go(-1)">[返回前页]</a>`)) {
		t.Fatal("school error notice missed")
	}
}

func TestCASRedirectWaitsBeforeBusinessVerification(t *testing.T) {
	var homeCalls atomic.Int32
	var landed time.Time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Host + r.URL.Path {
		case "ids.chd.edu.cn" + casLandingPath:
			fmt.Fprint(w, "<html>account portal</html>")
		case "ids.chd.edu.cn/authserver/login":
			http.Redirect(w, r, eamsOrigin+"/eams/home.action", 302)
		case "bkjw.chd.edu.cn/eams/home.action":
			if homeCalls.Add(1) == 1 {
				landed = time.Now()
			} else if time.Since(landed) < 95*time.Millisecond {
				t.Error("CAS landing was not included in pacing")
			}
			fmt.Fprint(w, "<html>教务首页</html>")
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	c := &connection{ctx: context.Background(), client: &http.Client{Transport: mappedTransport{server.URL}, CheckRedirect: checkRedirect}, pacer: &requestPacer{interval: 100 * time.Millisecond}}
	defer c.client.CloseIdleConnections()
	if err := c.completeLogin(casOrigin + casLandingPath); err != nil {
		t.Fatal(err)
	}
	if homeCalls.Load() != 2 {
		t.Fatal("business page verification missing")
	}
}
