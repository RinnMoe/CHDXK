package chdteachers

import (
	"jcourse/internal/domain/teacher"
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
