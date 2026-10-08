//go:build test

package bkjw

import (
	"context"
	"encoding/json"
	"fmt"
	"jcourse/internal/domain/course"
	"net/http"
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
	c = &connection{ctx: ctx}
	calls := 0
	retries := 0
	err := c.retry(func() error { calls++; return codeError("AUTH_REQUIRED") }, func(int) { retries++ })
	if err == nil || calls != 1 || retries != 0 {
		t.Fatal("auth retried")
	}
	err = c.retry(func() error {
		calls++
		if calls < 4 {
			return fmt.Errorf("temporary 503")
		}
		return nil
	}, func(int) { retries++ })
	if err != nil || retries != 2 {
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
