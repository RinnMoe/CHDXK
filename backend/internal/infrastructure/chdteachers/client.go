package chdteachers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"jcourse/internal/domain/teacher"
)

const origin = "https://js.chd.edu.cn"

type Client struct{ HTTP *http.Client }

func NewClient() *Client {
	return &Client{HTTP: &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if r.URL.Hostname() != "js.chd.edu.cn" || len(via) > 5 {
			return fmt.Errorf("教师主页跳转地址无效")
		}
		return nil
	}}}
}
func sourceURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	u = urlBase.ResolveReference(u)
	if u.Hostname() != "js.chd.edu.cn" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return "", fmt.Errorf("教师主页地址无效")
	}
	u.Scheme = "https"
	u.Fragment = ""
	return u.String(), nil
}

var urlBase, _ = url.Parse(origin)

func (c *Client) request(ctx context.Context, address string, body url.Values) ([]byte, error) {
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}
		method := http.MethodGet
		var reader io.Reader
		if body != nil {
			method = http.MethodPost
			reader = strings.NewReader(body.Encode())
		}
		req, err := http.NewRequestWithContext(ctx, method, address, reader)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "CHDXK teacher profile sync")
		if body != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
		}
		resp, err := c.HTTP.Do(req)
		if err != nil {
			if attempt == 2 {
				return nil, fmt.Errorf("教师主页连接失败: %w", err)
			}
			continue
		}
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		resp.Body.Close()
		if resp.StatusCode == 200 && readErr == nil {
			return data, nil
		}
		if resp.StatusCode < 500 && resp.StatusCode != 429 {
			return nil, fmt.Errorf("教师主页返回 HTTP %d", resp.StatusCode)
		}
	}
	return nil, fmt.Errorf("教师主页请求重试后仍失败")
}
func (c *Client) List(ctx context.Context, progress func(int, int)) ([]teacher.Profile, error) {
	return c.list(ctx, "", progress)
}
func (c *Client) Find(ctx context.Context, name, department string) ([]teacher.Profile, error) {
	rows, err := c.list(ctx, strings.TrimSpace(name), func(int, int) {})
	if err != nil {
		return nil, err
	}
	matches := []teacher.Profile{}
	for _, p := range rows {
		if teacher.ProfileIdentity(p.Name, p.Department) == teacher.ProfileIdentity(name, department) {
			matches = append(matches, p)
		}
	}
	return matches, nil
}
func (c *Client) list(ctx context.Context, name string, progress func(int, int)) ([]teacher.Profile, error) {
	result := []teacher.Profile{}
	seen := map[string]bool{}
	for page := 1; page <= 500; page++ {
		body := url.Values{"siteId": {"2"}, "pageIndex": {strconv.Itoa(page)}, "rows": {"100"}, "articleType": {"1"}, "level": {"0"}, "orders": {""}, "pageEvent": {"dataSearchByPageIndex"}, "conditions": {`[{"field":"language","value":"1","judge":"="},{"field":"ownDepartment","value":"","judge":"="},{"field":"title","value":"","judge":"like"},{"field":"published","value":"1","judge":"="}]`}, "returnInfos": {`[{"field":"title","name":"title"},{"field":"career","name":"career"},{"field":"cnUrl","name":"cnUrl"},{"field":"department","name":"department"},{"field":"publishStatus","name":"publishStatus"}]`}}
		if name != "" {
			conditions := []map[string]string{{"field": "language", "value": "1", "judge": "="}, {"field": "ownDepartment", "value": "", "judge": "="}, {"field": "title", "value": name, "judge": "like"}, {"field": "published", "value": "1", "judge": "="}}
			encoded, _ := json.Marshal(conditions)
			body.Set("conditions", string(encoded))
		}
		data, err := c.request(ctx, origin+"/_wp3services/generalQuery?queryObj=teacherHome", body)
		if err != nil {
			return nil, err
		}
		var response struct {
			Total     int `json:"total"`
			PageCount int `json:"pageCount"`
			Data      []struct {
				Name       string `json:"title"`
				Title      string `json:"career"`
				URL        string `json:"cnUrl"`
				Department string `json:"department"`
				Published  int    `json:"publishStatus"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("教师目录格式异常")
		}
		if name != "" && response.Total == 0 && response.Data != nil && len(response.Data) == 0 {
			return result, nil
		}
		if response.Total < 1 || response.PageCount < 1 || len(response.Data) == 0 {
			return nil, fmt.Errorf("教师目录格式异常，未保存本次同步")
		}
		added := 0
		for _, row := range response.Data {
			if row.Published != 1 {
				continue
			}
			address, err := sourceURL(row.URL)
			if err != nil || row.Name == "" || row.Department == "" {
				return nil, fmt.Errorf("教师目录缺少姓名、学院或有效主页地址")
			}
			if !seen[address] {
				result = append(result, teacher.Profile{URL: address, Name: strings.TrimSpace(row.Name), Title: strings.TrimSpace(row.Title), Department: strings.TrimSpace(row.Department), Fields: map[string]string{}})
				seen[address] = true
				added++
			}
		}
		progress(page, response.PageCount)
		if page >= response.PageCount {
			return result, nil
		}
		if added == 0 {
			return nil, fmt.Errorf("教师目录分页未推进，未保存本次同步")
		}
	}
	return nil, fmt.Errorf("教师目录分页超出范围")
}
func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
func class(n *html.Node, key string) bool {
	for _, v := range strings.Fields(attr(n, "class")) {
		if v == key {
			return true
		}
	}
	return false
}
func text(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	if n.Data == "script" || n.Data == "style" {
		return ""
	}
	var b strings.Builder
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		b.WriteString(text(child))
		if child.Data == "p" || child.Data == "br" || child.Data == "li" {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
func walk(n *html.Node, visit func(*html.Node)) {
	visit(n)
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		walk(child, visit)
	}
}

var labels = map[string]bool{"职称": true, "学位": true, "学历": true, "毕业院校": true, "电子邮箱": true, "联系电话": true, "联系方式": true, "办公地点": true, "办公地址": true, "教育经历": true, "个人简介": true, "研究领域": true, "研究方向": true, "工作经历": true, "社会职务": true, "科研项目": true, "论文": true, "科技成果": true, "荣誉奖励": true, "开授课程": true}

func ParseDetail(data []byte, p teacher.Profile) (teacher.Profile, error) {
	root, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return p, err
	}
	fields := map[string]string{}
	found := false
	walk(root, func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}
		if class(n, "post") {
			var heading, content string
			walk(n, func(child *html.Node) {
				if child.Data == "h3" {
					heading = strings.TrimSpace(text(child))
				}
				if class(child, "con") {
					content = strings.TrimSpace(text(child))
				}
			})
			if labels[heading] {
				found = true
				if len(content) > 30000 {
					content = string([]rune(content)[:min(10000, len([]rune(content)))])
				}
				fields[heading] = content
			}
		}
		if n.Data == "li" || n.Data == "p" {
			s := strings.TrimSpace(text(n))
			parts := strings.SplitN(s, "：", 2)
			if len(parts) == 2 && labels[strings.TrimSpace(parts[0])] {
				found = true
				fields[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	})
	if !found {
		return p, fmt.Errorf("教师详细资料模板无法识别")
	}
	p.Fields = fields
	p.UpdatedAt = time.Now().UTC()
	if fields["职称"] != "" {
		p.Title = fields["职称"]
	}
	return p, nil
}
func (c *Client) Detail(ctx context.Context, p teacher.Profile) (teacher.Profile, error) {
	address, err := sourceURL(p.URL)
	if err != nil {
		return p, err
	}
	data, err := c.request(ctx, address, nil)
	if err != nil {
		return p, err
	}
	return ParseDetail(data, p)
}
