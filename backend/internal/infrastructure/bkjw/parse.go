package bkjw

import (
	"fmt"
	"golang.org/x/net/html"
	stdhtml "html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type semester struct{ ID, Name string }

var objectPattern = regexp.MustCompile(`\{[^{}]{1,512}\}`)
var idField = regexp.MustCompile(`(?:["']?id["']?)\s*:\s*["']?(\d+)`)
var yearField = regexp.MustCompile(`(?:["']?schoolYear["']?)\s*:\s*["']?(\d{4}-\d{4})`)
var termField = regexp.MustCompile(`(?:["']?name["']?)\s*:\s*["']?([1-9]\d*)`)
var yearPattern = regexp.MustCompile(`\d{4}-\d{4}`)
var numberPattern = regexp.MustCompile(`-?(?:\d+\.?\d*|\.\d+)`)
var digitPattern = regexp.MustCompile(`^\d+$`)
var rangePattern = regexp.MustCompile(`(\d+)\s*-\s*(\d+)\s+of\s+(\d+)`)
var pageInfoPattern = regexp.MustCompile(`\.pageInfo\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)`)

func parseSemesters(data []byte) ([]semester, error) {
	out := []semester{}
	seen := map[string]string{}
	add := func(id, name string) error {
		if !digitPattern.MatchString(id) || !regexp.MustCompile(`^\d{4}-\d{4}-[1-9]\d*$`).MatchString(name) {
			return nil
		}
		if old, ok := seen[name]; ok {
			if old != id {
				return codeError("PAGE_STRUCTURE")
			}
			return nil
		}
		seen[name] = id
		out = append(out, semester{id, name})
		return nil
	}
	for _, object := range objectPattern.FindAllString(string(data), -1) {
		id := match(idField, object)
		year := match(yearField, object)
		term := match(termField, object)
		if id != "" && year != "" && term != "" {
			if err := add(id, year+"-"+term); err != nil {
				return nil, err
			}
		}
	}
	for _, n := range nodes(document(data), func(n *html.Node) bool { return attr(n, "data-semester-id") != "" }) {
		if err := add(attr(n, "data-semester-id"), attr(n, "data-academic-year")+"-"+attr(n, "data-term")); err != nil {
			return nil, err
		}
	}
	if len(out) == 0 {
		return nil, codeError("SEMESTER_DISCOVERY_FAILED")
	}
	return out, nil
}
func (c *connection) semesters(page []byte) ([]semester, error) {
	// Always fetch the complete catalog; the rendered selector can contain only one year.
	tagID := ""
	value := ""
	doc := document(page)
	for _, n := range tag(doc, "input") {
		if hasClass(n, "calendar-text") {
			tagID = attr(n, "id")
			break
		}
	}
	value = match(regexp.MustCompile(`semesterCalendar\s*\(\s*\{[^}]*\bvalue\s*:\s*["']?(\d+)`), string(page))
	if tagID == "" {
		return parseSemesters(page)
	}
	data, _, err := c.eams("POST", "/eams/dataQuery.action", url.Values{"tagId": {tagID}, "dataType": {"semesterCalendar"}, "value": {value}, "empty": {"false"}})
	if err != nil {
		return nil, err
	}
	return parseSemesters(data)
}

type listing struct {
	Lessons           []map[string]any
	Page, Size, Total int
	Signature         string
}

func label(s string) string { return strings.TrimRight(strings.TrimSpace(s), ":：") }
func people(s string) []string {
	out := []string{}
	for _, v := range regexp.MustCompile(`[,，;；、/\s]+`).Split(s, -1) {
		if v != "" && v != "0" {
			out = append(out, v)
		}
	}
	return out
}
func number(s string) any {
	s = numberPattern.FindString(strings.ReplaceAll(s, ",", ""))
	if s == "" {
		return nil
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return n
}
func integer(s string) int { n, _ := strconv.Atoi(s); return n }
func parseListing(data []byte) (listing, error) {
	result := listing{Lessons: []map[string]any{}, Page: 1}
	doc := document(data)
	var grid *html.Node
	for _, table := range tag(doc, "table") {
		if hasClass(table, "gridtable") {
			grid = table
			break
		}
	}
	if grid == nil {
		return result, codeError("PAGE_STRUCTURE")
	}
	headers := []string{}
	for _, row := range tag(grid, "tr") {
		cs := cells(row)
		count := 0
		for _, cell := range cs {
			if cell.Data == "th" {
				count++
			}
		}
		if count < 5 {
			continue
		}
		headers = nil
		for _, cell := range cs {
			v := text(cell)
			if title := attr(cell, "title"); title != "" {
				v = title
			}
			if sorted := match(regexp.MustCompile(`\[\s*([^\]]+?)\s*\]`), v); sorted != "" {
				v = sorted
			}
			headers = append(headers, label(v))
		}
	}
	for _, required := range []string{"课程序号", "课程代码", "课程名称", "课程类别", "教学班", "教师"} {
		found := false
		for _, h := range headers {
			if h == required {
				found = true
			}
		}
		if !found {
			return result, codeError("PAGE_STRUCTURE")
		}
	}
	result.Signature = strings.Join(headers, "|")
	for _, row := range tag(grid, "tr") {
		id := ""
		for _, input := range tag(row, "input") {
			if attr(input, "name") == "lesson.id" {
				id = attr(input, "value")
				break
			}
		}
		if id == "" {
			continue
		}
		if !digitPattern.MatchString(id) {
			return result, codeError("PAGE_STRUCTURE")
		}
		raw := map[string]string{}
		cs := cells(row)
		for i, cell := range cs {
			if i < len(headers) && headers[i] != "" {
				raw[headers[i]] = text(cell)
			}
		}
		detail, syllabus := "", ""
		for _, a := range tag(row, "a") {
			href := attr(a, "href")
			if strings.Contains(href, "stdSyllabus!info.action") {
				detail = href
			}
			if strings.Contains(href, "stdSyllabus!downloadSyllabus.action") {
				syllabus = href
			}
		}
		if detail == "" {
			return result, codeError("PAGE_STRUCTURE")
		}
		full, err := resolve(eamsOrigin+"/eams/", detail)
		if err != nil || !strings.HasPrefix(full, eamsOrigin+"/eams/") {
			return result, codeError("PAGE_STRUCTURE")
		}
		if syllabus != "" {
			full, e := resolve(eamsOrigin+"/eams/", syllabus)
			if e != nil || !strings.HasPrefix(full, eamsOrigin+"/eams/") {
				return result, codeError("PAGE_STRUCTURE")
			}
		}
		result.Lessons = append(result.Lessons, map[string]any{"id": id, "summary": map[string]any{
			"lesson_no": raw["课程序号"], "course_code": raw["课程代码"], "course_name": raw["课程名称"], "course_type": raw["课程类别"], "teach_class": raw["教学班"], "teachers": people(raw["教师"]), "actual_count": number(raw["实际"]), "limit_count": number(raw["上限"]), "credit": number(raw["学分"]), "course_period": raw["学时/周"], "week_range": raw["起止周"], "raw": raw}, "detail": nil, "links": map[string]any{"detail": detail, "syllabus": syllabus}, "warnings": []string{}})
	}
	// HTTP responses can expose pagination only through pageInfo(), before DOM rendering.
	if m := pageInfoPattern.FindStringSubmatch(string(data)); len(m) > 0 {
		result.Page = integer(m[1])
		result.Size = integer(m[2])
		result.Total = integer(m[3])
	} else if m := rangePattern.FindStringSubmatch(text(doc)); len(m) > 0 {
		start, end := integer(m[1]), integer(m[2])
		result.Total = integer(m[3])
		result.Size = end - start + 1
		for _, selectNode := range tag(doc, "select") {
			if strings.HasSuffix(attr(selectNode, "id"), "_page_select") {
				for _, option := range tag(selectNode, "option") {
					for _, a := range option.Attr {
						if a.Key == "selected" {
							result.Size = integer(attr(option, "value"))
						}
					}
				}
			}
		}
		if result.Size > 0 {
			result.Page = (start-1)/result.Size + 1
		}
	} else {
		return result, codeError("PAGE_STRUCTURE")
	}
	if result.Page < 1 || result.Size < 1 || result.Total < 0 || result.Total > 20000 || len(result.Lessons) > result.Size {
		return result, codeError("PAGE_STRUCTURE")
	}
	return result, nil
}
func parseDetail(data []byte) (map[string]any, error) {
	fields := map[string]string{}
	restrictions := []map[string]any{}
	for _, row := range tag(document(data), "tr") {
		nested := false
		for p := row.Parent; p != nil; p = p.Parent {
			if p.Data == "td" {
				nested = true
				break
			}
		}
		if nested {
			continue
		}
		cs := cells(row)
		for i := 0; i+1 < len(cs); i += 2 {
			key := label(text(cs[i]))
			if key != "" {
				fields[key] = text(cs[i+1])
			}
			if key == "限制条件组" {
				raw := text(cs[i+1])
				if raw != "" && raw != "无" && raw != "不限" {
					restrictions = append(restrictions, map[string]any{"title": "原始限制条件", "items": []any{}, "raw": raw})
				}
			}
		}
	}
	if fields["课程代码"] == "" && fields["课程序号"] == "" {
		return nil, codeError("PAGE_STRUCTURE")
	}
	return map[string]any{"basic": map[string]any{"semester": fields["学期"], "course_code": fields["课程代码"], "course_name": fields["课程名称"], "credit": number(fields["学分"]), "course_type": fields["课程类别"], "department": fields["开课院系"], "campus": fields["校区"], "language": fields["授课语言"], "assessment": fields["考核方式"], "teachers": people(fields["教师"])}, "arrangement": map[string]any{"total_hours": number(fields["总课时"]), "weekly_hours": number(fields["周课时"]), "week_range": fields["起止周"], "start_week": number(fields["起始周"]), "end_week": number(fields["结束周"]), "week_count": number(fields["周数"])}, "restrictions": restrictions, "remark": fields["备注"], "raw": fields}, nil
}

// Decode a JS string literal without an interpreter. Calls/expressions are rejected.
func jsString(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) < 2 || (raw[0] != '\'' && raw[0] != '"') || raw[len(raw)-1] != raw[0] {
		return "", codeError("PERSONAL_PARSE_FAILED")
	}
	var b strings.Builder
	for i := 1; i < len(raw)-1; i++ {
		if raw[i] != '\\' {
			if raw[i] == raw[0] || raw[i] == '\n' || raw[i] == '\r' {
				return "", codeError("PERSONAL_PARSE_FAILED")
			}
			b.WriteByte(raw[i])
			continue
		}
		i++
		if i >= len(raw)-1 {
			return "", codeError("PERSONAL_PARSE_FAILED")
		}
		switch raw[i] {
		case '\\', '\'', '"', '/':
			b.WriteByte(raw[i])
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case 'x', 'u':
			size := 2
			if raw[i] == 'u' {
				size = 4
			}
			if i+size >= len(raw)-1 {
				return "", codeError("PERSONAL_PARSE_FAILED")
			}
			n, e := strconv.ParseUint(raw[i+1:i+1+size], 16, 32)
			if e != nil {
				return "", e
			}
			b.WriteRune(rune(n))
			i += size
		default:
			return "", codeError("PERSONAL_PARSE_FAILED")
		}
	}
	return stdhtml.UnescapeString(b.String()), nil
}

// Split constructor arguments while respecting strings, nested calls and commas.
func arguments(raw string) ([]string, int, error) {
	args := []string{}
	start, depth := 0, 0
	quote := byte(0)
	escape := false
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if quote != 0 {
			if escape {
				escape = false
			} else if ch == '\\' {
				escape = true
			} else if ch == quote {
				quote = 0
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			quote = ch
			continue
		}
		switch ch {
		case '(', '[', '{':
			depth++
		case ')':
			if depth == 0 {
				args = append(args, strings.TrimSpace(raw[start:i]))
				return args, i + 1, nil
			}
			depth--
		case ']', '}':
			depth--
		case ',':
			if depth == 0 {
				args = append(args, strings.TrimSpace(raw[start:i]))
				start = i + 1
			}
		}
		if depth < 0 {
			return nil, 0, codeError("PERSONAL_PARSE_FAILED")
		}
	}
	return nil, 0, fmt.Errorf("unterminated constructor")
}
