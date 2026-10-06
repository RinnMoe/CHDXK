package chdteachers

import (
	"context"
	"io"
	"jcourse/internal/domain/teacher"
	"net/http"
	"strings"
	"testing"
)

func TestParseProfessionalFieldsAsPlainText(t *testing.T) {
	data := []byte(`<div class="post"><h3>个人简介</h3><div class="con"><p>从事道路研究</p><script>alert(1)</script></div></div><li><span>职称：</span><span>教授</span></li><li>电子邮箱：teacher@example.edu</li><li>学位：</li>`)
	p, err := ParseDetail(data, teacher.Profile{URL: "https://js.chd.edu.cn/a/list.htm"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Title != "教授" || p.Fields["电子邮箱"] != "teacher@example.edu" || p.Fields["学位"] != "" || p.Fields["个人简介"] != "从事道路研究" {
		t.Fatalf("unexpected parsed fields: %+v", p)
	}
	if _, err := ParseDetail([]byte(`<html>系统维护</html>`), teacher.Profile{}); err == nil {
		t.Fatal("maintenance page accepted as empty profile")
	}
}
func TestSourceURLRestrictsExternalHosts(t *testing.T) {
	for _, address := range []string{"https://example.com/list.htm", "javascript:alert(1)", "https://js.chd.edu.cn@evil.test/a"} {
		if _, err := sourceURL(address); err == nil {
			t.Fatal("external URL allowed")
		}
	}
	got, err := sourceURL("http://js.chd.edu.cn/a/list.htm")
	if err != nil || got != "https://js.chd.edu.cn/a/list.htm" {
		t.Fatal(got, err)
	}
}

type profileTransport func(*http.Request) (*http.Response, error)

func (f profileTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestFindFiltersExactNameAndCollegeAndAllowsEmptyResults(t *testing.T) {
	empty := false
	client := &Client{HTTP: &http.Client{Transport: profileTransport(func(r *http.Request) (*http.Response, error) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(r.Form.Get("conditions"), `"value":"张三"`) {
			t.Fatal("name filter missing")
		}
		body := `{"total":3,"pageCount":1,"data":[{"title":"张三","department":"理学院","cnUrl":"/a/list.htm","publishStatus":1},{"title":"张三","department":"外国语学院","cnUrl":"/b/list.htm","publishStatus":1},{"title":"张三丰","department":"理学院","cnUrl":"/c/list.htm","publishStatus":1}]}`
		if empty {
			body = `{"total":0,"pageCount":0,"data":[]}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	rows, err := client.Find(context.Background(), "张三", "理学院")
	if err != nil || len(rows) != 1 || rows[0].URL != "https://js.chd.edu.cn/a/list.htm" {
		t.Fatal(rows, err)
	}
	empty = true
	rows, err = client.Find(context.Background(), "张三", "理学院")
	if err != nil || len(rows) != 0 {
		t.Fatal(rows, err)
	}
}
