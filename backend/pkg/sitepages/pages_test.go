package sitepages

import "testing"

func TestPagesValidation(t *testing.T) {
	pages, err := Parse(Defaults)
	if err != nil || len(pages) != 2 {
		t.Fatalf("defaults: %v", err)
	}
	for _, value := range []string{`null`, `[{"slug":"../admin","title":"bad"}]`, `[{"slug":"links","title":""}]`, `[{"slug":"links","title":"A"},{"slug":"links","title":"B"}]`, `[] {}`} {
		if _, err := Parse(value); err == nil {
			t.Fatalf("accepted %s", value)
		}
	}
	if _, err := Parse(`[{"slug":"links","title":"友链","content":"[站点](https://example.com)","published":false,"footer":true}]`); err != nil {
		t.Fatal(err)
	}
}
