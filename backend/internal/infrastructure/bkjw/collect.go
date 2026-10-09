package bkjw

import (
	"encoding/json"
	"jcourse/internal/domain/course"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func (c *connection) progress(s semester, stage string, completed, total, retries, page, pages int) {
	c.emit(course.SourceEvent{Type: "progress", Semester: s.Name, Progress: &course.SourceProgress{Stage: stage, Completed: completed, Total: &total, Retries: retries, Page: page, Pages: pages, UpdatedAt: time.Now(), Cached: c.restored, RetryAt: c.retryAt}})
}
func (c *connection) adminArchive(s semester) (course.SourceEvent, error) {
	lessons := []map[string]any{}
	seen := map[string]bool{}
	total := -1
	signature := ""
	size := 100
	retries := 0
	checkpoint, err := c.cache.load(s)
	if err != nil {
		return course.SourceEvent{}, err
	}
	if c.fresh {
		checkpoint = nil
	}
	startPage := 1
	if checkpoint != nil {
		lessons, total, signature, size = checkpoint.Lessons, checkpoint.Total, checkpoint.Signature, checkpoint.Size
		startPage = checkpoint.NextPage
		for _, lesson := range lessons {
			seen[lesson["id"].(string)] = true
		}
		c.progress(s, "listing", len(lessons), total, 0, startPage-1, max(1, (total+size-1)/size))
	}
	for page := startPage; page <= 2000; page++ {
		if total >= 0 && len(lessons) == total {
			break
		}
		var listing listing
		err := c.retry(func() error {
			data, _, err := c.eams("POST", "/eams/stdSyllabus!search.action", url.Values{"lesson.project.id": {"1"}, "lesson.semester.id": {s.ID}, "pageNo": {itoa(page)}, "pageSize": {itoa(size)}})
			if err != nil {
				return err
			}
			listing, err = parseListing(data)
			if err != nil {
				slog.Warn("academic listing parse failed", "semester", s.Name, "page", page)
				return err
			}
			if listing.Page != page || (total >= 0 && (listing.Total != total || listing.Signature != signature || listing.Size != size)) {
				slog.Warn("academic listing pagination mismatch", "semester", s.Name, "requested_page", page, "received_page", listing.Page, "expected_total", total, "received_total", listing.Total, "expected_size", size, "received_size", listing.Size, "same_columns", total < 0 || listing.Signature == signature)
				return codeError("PAGE_STRUCTURE")
			}
			expected := min(listing.Size, max(listing.Total-(page-1)*listing.Size, 0))
			if len(listing.Lessons) != expected {
				slog.Warn("academic listing count mismatch", "semester", s.Name, "page", page, "expected", expected, "received", len(listing.Lessons))
				return codeError("PAGE_STRUCTURE")
			}
			pageIDs := map[string]bool{}
			for _, lesson := range listing.Lessons {
				id := lesson["id"].(string)
				if seen[id] || pageIDs[id] {
					slog.Warn("academic listing duplicate rows", "semester", s.Name, "page", page)
					return codeError("PAGE_STRUCTURE")
				}
				pageIDs[id] = true
			}
			return nil
		}, func(int) { retries++; c.progress(s, "listing", len(lessons), max(total, 0), retries, 0, 0) })
		if err != nil {
			return course.SourceEvent{}, err
		}
		if total < 0 {
			total = listing.Total
			signature = listing.Signature
			size = listing.Size
		}
		for _, lesson := range listing.Lessons {
			id := lesson["id"].(string)
			seen[id] = true
			lessons = append(lessons, lesson)
		}
		pages := max(1, (total+size-1)/size)
		if checkpoint == nil {
			checkpoint = &catalogCheckpoint{Version: 1, CreatedAt: time.Now(), Semester: s.Name, SemesterID: s.ID}
		}
		checkpoint.Lessons, checkpoint.Total, checkpoint.Signature, checkpoint.Size, checkpoint.NextPage = lessons, total, signature, size, page+1
		if err := c.cache.saveCatalog(s, checkpoint); err != nil {
			return course.SourceEvent{}, err
		}
		c.progress(s, "listing", len(lessons), total, retries, page, pages)
		if page >= pages {
			break
		}
	}
	if len(lessons) != total {
		return course.SourceEvent{}, codeError("PAGE_STRUCTURE")
	}
	for i, lesson := range lessons {
		if detail := c.cache.loadDetail(s, checkpoint, lesson); detail != nil {
			lesson["detail"] = detail
			c.restored++
			c.progress(s, "details", i+1, total, retries, 0, 0)
			continue
		}
		path := lesson["links"].(map[string]any)["detail"].(string)
		err := c.retry(func() error {
			data, _, err := c.eams("GET", path, nil)
			if err != nil {
				return err
			}
			detail, err := parseDetail(data)
			if err != nil {
				return err
			}
			basic := detail["basic"].(map[string]any)
			parts := strings.Split(s.Name, "-")
			if strings.Join(strings.Fields(basic["semester"].(string)), "") != parts[0]+"-"+parts[1]+"学年"+parts[2]+"学期" {
				return codeError("PAGE_STRUCTURE")
			}
			summary := lesson["summary"].(map[string]any)
			if code := basic["course_code"].(string); code != "" && summary["course_code"] != code {
				return codeError("PAGE_STRUCTURE")
			}
			lesson["detail"] = detail
			return nil
		}, func(int) { retries++; c.progress(s, "details", i, total, retries, 0, 0) })
		if err != nil {
			return course.SourceEvent{}, err
		}
		if err := c.cache.saveDetail(s, checkpoint, lesson); err != nil {
			return course.SourceEvent{}, err
		}
		c.progress(s, "details", i+1, total, retries, 0, 0)
	}
	c.progress(s, "validating", total, total, retries, 0, 0)
	document := marshal(map[string]any{"schema_version": 1, "source": map[string]any{"host": "bkjw.chd.edu.cn", "semester_id": s.ID}, "semester": s.Name, "expected_count": total, "lessons": lessons})
	report := marshal(map[string]any{"schema_version": 1, "semester": s.Name, "completed": true, "list_complete": true, "expected_count": total, "lesson_count": total, "detail_success_count": total, "missing_detail_count": 0, "duplicate_lesson_ids": []string{}, "failed_lesson_ids": []string{}, "source_page": map[string]any{"host": "bkjw.chd.edu.cn", "grid_signature": signature}})
	if len(document)+len(report) > 32<<20 {
		return course.SourceEvent{}, codeError("PAGE_STRUCTURE")
	}
	return course.SourceEvent{Type: "archive", Semester: s.Name, Document: document, Report: report}, nil
}

var stdIDsPattern = regexp.MustCompile(`(?s)(?:val\(\)\s*==\s*["']std["']|["']std["']\s*==[^\n{]+)\s*\)\s*\{[^}]*?addInput\s*\(\s*form\s*,\s*["']ids["']\s*,\s*["']([\d,]+)["']`)
var addInputPattern = regexp.MustCompile(`addInput\s*\(\s*form\s*,\s*["']([^"']+)["']\s*,\s*["']([^"']*)["']`)
var sourceLessonPattern = regexp.MustCompile(`^(\d+)(?:\([^)]*\))?$`)

func personalParams(page []byte) (url.Values, error) {
	values := url.Values{}
	for _, n := range tag(document(page), "input") {
		name := attr(n, "name")
		if name != "" {
			values.Set(name, attr(n, "value"))
		}
	}
	// Do not select the class branch's ids by collecting all addInput calls.
	stdID := match(stdIDsPattern, string(page))
	if stdID != "" {
		values.Set("ids", stdID)
	}
	for _, m := range addInputPattern.FindAllStringSubmatch(string(page), -1) {
		if m[1] != "ids" && m[1] != "semester.id" {
			values.Set(m[1], m[2])
		}
	}
	if !regexp.MustCompile(`^\d+(?:,\d+)*$`).MatchString(values.Get("ids")) {
		return nil, codeError("PERSONAL_QUERY_FAILED")
	}
	values.Set("setting.kind", "std")
	values.Set("ignoreHead", "1")
	values.Set("showPrintAndExport", "1")
	return values, nil
}
func (c *connection) personalArchive(page []byte, s semester) (course.SourceEvent, error) {
	params, err := personalParams(page)
	if err != nil {
		return course.SourceEvent{}, err
	}
	params.Set("semester.id", s.ID)
	var courses []course.EnrollmentSourceCourse
	retries := 0
	err = c.retry(func() error {
		data, _, err := c.eams("POST", "/eams/courseTableForStd!courseTable.action", params)
		if err != nil {
			return err
		}
		courses, err = parsePersonal(data)
		return err
	}, func(int) { retries++; c.progress(s, "listing", 0, 0, retries, 0, 0) })
	if err != nil {
		if _, typed := err.(codeError); !typed {
			err = codeError("PERSONAL_REQUEST_FAILED")
		}
		return course.SourceEvent{}, err
	}
	c.progress(s, "validating", len(courses), len(courses), retries, 0, 0)
	return course.SourceEvent{Type: "archive", Semester: s.Name, Document: marshal(map[string]any{"semester": s.Name, "courses": courses})}, nil
}

func parsePersonal(data []byte) ([]course.EnrollmentSourceCourse, error) {
	out := []course.EnrollmentSourceCourse{}
	seen := map[string]bool{}
	add := func(id, lesson, name string) error {
		name = strings.TrimSpace(name)
		lesson = strings.TrimSpace(lesson)
		id = match(sourceLessonPattern, id)
		if name == "" || (id == "" && lesson == "") {
			return codeError("PERSONAL_PARSE_FAILED")
		}
		key := lesson
		if key == "" {
			key = id
		}
		if !seen[key] {
			seen[key] = true
			out = append(out, course.EnrollmentSourceCourse{SourceLessonID: id, LessonNo: lesson, Name: name})
		}
		if len(out) > 1000 {
			return codeError("PERSONAL_PARSE_FAILED")
		}
		return nil
	}
	// JSON fixture/API responses are parsed explicitly, never carried over between semesters.
	var payload struct {
		Activities json.RawMessage `json:"activities"`
	}
	if json.Unmarshal(data, &payload) == nil && payload.Activities != nil {
		var walk func(json.RawMessage, int) error
		walk = func(raw json.RawMessage, depth int) error {
			if depth > 2 {
				return codeError("PERSONAL_PARSE_FAILED")
			}
			if len(raw) > 0 && raw[0] == '[' {
				var items []json.RawMessage
				if json.Unmarshal(raw, &items) != nil || len(items) > 10000 {
					return codeError("PERSONAL_PARSE_FAILED")
				}
				for _, item := range items {
					if err := walk(item, depth+1); err != nil {
						return err
					}
				}
				return nil
			}
			var v struct {
				CourseID string `json:"courseId"`
				LessonNo string `json:"lessonNo"`
				Name     string `json:"courseName"`
			}
			if json.Unmarshal(raw, &v) != nil {
				return codeError("PERSONAL_PARSE_FAILED")
			}
			return add(v.CourseID, v.LessonNo, v.Name)
		}
		if err := walk(payload.Activities, 0); err != nil {
			return nil, err
		}
		return out, nil
	}
	content := string(data)
	cursor := 0
	count := 0
	for {
		index := strings.Index(content[cursor:], "new TaskActivity(")
		if index < 0 {
			break
		}
		start := cursor + index + len("new TaskActivity(")
		args, end, err := arguments(content[start:])
		if err != nil || len(args) < 4 {
			return nil, codeError("PERSONAL_PARSE_FAILED")
		}
		id, err := jsString(args[2])
		if err != nil {
			return nil, codeError("PERSONAL_PARSE_FAILED")
		}
		name, err := jsString(args[3])
		if err != nil {
			return nil, codeError("PERSONAL_PARSE_FAILED")
		}
		if err = add(id, "", name); err != nil {
			return nil, err
		}
		cursor = start + end
		count++
		if count > 10000 {
			return nil, codeError("PERSONAL_PARSE_FAILED")
		}
	}
	// A known empty timetable must still have a valid table initializer.
	if count == 0 && !regexp.MustCompile(`new\s+CourseTable\s*\(`).MatchString(content) {
		return nil, codeError("PERSONAL_PARSE_FAILED")
	}
	return out, nil
}
