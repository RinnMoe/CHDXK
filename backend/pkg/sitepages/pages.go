package sitepages

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

const Key = "content.pages"

//go:embed defaults.json
var Defaults string

type Page struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Published bool   `json:"published"`
	Footer    bool   `json:"footer"`
}

var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,47}$`)

func Parse(value string) ([]Page, error) {
	if len(value) > 512<<10 {
		return nil, fmt.Errorf("页面内容总量过大")
	}
	var pages []Page
	decoder := json.NewDecoder(bytes.NewBufferString(value))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&pages); err != nil {
		return nil, fmt.Errorf("页面内容格式无效")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("页面内容格式无效")
	}
	if pages == nil || len(pages) > 30 {
		return nil, fmt.Errorf("最多添加 30 个页面")
	}
	seen := map[string]bool{}
	for _, page := range pages {
		if !slugPattern.MatchString(page.Slug) || seen[page.Slug] {
			return nil, fmt.Errorf("页面地址无效或重复")
		}
		seen[page.Slug] = true
		if strings.TrimSpace(page.Title) == "" || utf8.RuneCountInString(page.Title) > 80 || utf8.RuneCountInString(page.Content) > 32000 {
			return nil, fmt.Errorf("标题不能为空且不超过 80 字，内容不超过 32000 字")
		}
	}
	return pages, nil
}
