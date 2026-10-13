package bkjw

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"jcourse/internal/domain/course"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func (c *connection) progress(s semester, stage string, completed, total, retries, page, pages int) {
	c.emit(course.SourceEvent{Type: "progress", Semester: s.Name, Progress: &course.SourceProgress{Stage: stage, Completed: completed, Total: &total, Retries: retries, Failed: c.failedDetails, Page: page, Pages: pages, UpdatedAt: time.Now(), Cached: c.restored, RetryAt: c.retryAt}})
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
	completed := 0
	fetch := func(lesson map[string]any) error {
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
		}, func(int) { retries++; c.progress(s, "details", completed, total, retries, 0, 0) })
		if err != nil {
			return err
		}

		if err := c.cache.saveDetail(s, checkpoint, lesson); err != nil {
			return err
		}
		return nil
	}
	pending := []map[string]any{}
	for _, lesson := range lessons {
		if detail := c.cache.loadDetail(s, checkpoint, lesson); detail != nil {
			lesson["detail"] = detail
			c.restored++
			completed++
		} else {
			// Read other details before giving an individual outage the full retry budget.
			c.shortRetries = true
			err := fetch(lesson)
			c.shortRetries = false
			if err != nil {
				if fatalDetailError(err) || c.ctx.Err() != nil {
					return course.SourceEvent{}, err
				}
				pending = append(pending, lesson)
				c.failedDetails = len(pending)
			} else {
				completed++
			}
		}
		c.progress(s, "details", completed, total, retries, 0, 0)
	}
	checkpoint.FailedIDs = nil
	for _, lesson := range pending {
		checkpoint.FailedIDs = append(checkpoint.FailedIDs, lesson["id"].(string))
	}
	if err := c.cache.saveCatalog(s, checkpoint); err != nil {
		return course.SourceEvent{}, err
	}
	var lastError error
	failed := []string{}
	for _, lesson := range pending {
		if err := fetch(lesson); err != nil {
			if fatalDetailError(err) || c.ctx.Err() != nil {
				return course.SourceEvent{}, err
			}
			failed = append(failed, lesson["id"].(string))
			lastError = err
		} else {
			completed++
			c.failedDetails--
		}
		c.progress(s, "details", completed, total, retries, 0, 0)
	}
	checkpoint.FailedIDs = failed
	if err := c.cache.saveCatalog(s, checkpoint); err != nil {
		return course.SourceEvent{}, err
	}
	if lastError != nil {
		return course.SourceEvent{}, lastError
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
	if !c.fresh {
		if rows, ok := c.cache.loadPersonal(s, params.Get("ids")); ok {
			c.restored = len(rows)
			c.progress(s, "validating", len(rows), len(rows), 0, 0, 0)
			return course.SourceEvent{Type: "archive", Semester: s.Name, Document: marshal(map[string]any{"semester": s.Name, "courses": rows})}, nil
		}
	}
	var courses []course.EnrollmentSourceCourse
	retries := 0
	var previousBad [32]byte
	badSeen := false
	err = c.retry(func() error {
		data, _, err := c.eams("POST", "/eams/courseTableForStd!courseTable.action", params)
		if err != nil {
			return err
		}
		courses, err = parsePersonal(data)
		if err != nil {
			digest := sha256.Sum256(data)
			if badSeen && digest == previousBad {
				return permanentSourceError{err}
			}
			previousBad, badSeen = digest, true
		}
		return err
	}, func(int) { retries++; c.progress(s, "listing", 0, 0, retries, 0, 0) })
	if err != nil {
		var code codeError
		if !errors.As(err, &code) {
			err = codeError("PERSONAL_REQUEST_FAILED")
		}
		return course.SourceEvent{}, err
	}
	if err := c.cache.savePersonal(s, params.Get("ids"), courses); err != nil {
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
	sources := []string{}
	doc := document(data)
	for _, script := range tag(doc, "script") {
		sources = append(sources, rawText(script))
	}
	if len(sources) == 0 {
		sources = append(sources, string(data))
	}
	count, initialized := 0, false
	for _, source := range sources {
		clean, err := scriptCode(source, false)
		if err != nil {
			return nil, personalParseFailure("script-comments-or-strings", len(source), 0)
		}
		mask, err := scriptCode(source, true)
		if err != nil {
			return nil, personalParseFailure("script-mask", len(source), 0)
		}
		constructors := constructorPattern.FindAllStringSubmatchIndex(mask, -1)
		for _, loc := range constructors {
			name := mask[loc[2]:loc[3]]
			args, _, err := arguments(clean[loc[1]:])
			if err != nil {
				return nil, personalParseFailure("constructor-arguments", len(source), 0)
			}
			if name == "CourseTable" {
				initialized = true
				continue
			}
			if len(args) < 4 {
				return nil, personalParseFailure("activity-argument-count", len(source), len(args))
			}
			id, err := jsStringExpression(args[2])
			if err != nil {
				return nil, personalParseFailure("activity-id-literal", len(source), len(args))
			}
			name, err = boundStringExpression(args[3], precedingStringBindings(clean[:loc[0]], mask[:loc[0]]))
			if err != nil {
				return nil, personalNameFailure(args[3], len(source), len(args))
			}
			if err = add(id, "", name); err != nil {
				return nil, personalParseFailure("activity-identity-shape", len(source), len(args))
			}
			count++
			if count > 10000 {
				return nil, codeError("PERSONAL_PARSE_FAILED")
			}
		}
		if strings.Contains(mask, "TaskActivity") && len(constructors) == 0 {
			return nil, codeError("PERSONAL_PARSE_FAILED")
		}
	}
	if count == 0 && !initialized {
		return nil, personalParseFailure("missing-timetable-initializer", len(data), len(sources))
	}
	return out, nil
}

func fatalDetailError(err error) bool {
	switch errorCode(err, "") {
	case "AUTH_REQUIRED", "PERMISSION_DENIED", "URL_NOT_ALLOWED", "CACHE_WRITE_FAILED":
		return true
	}
	return false
}

func personalParseFailure(reason string, responseBytes, items int) error {
	// Fixed reason names and numeric structure only; never dump response text,
	// argument values, identity, URLs or cookies from a private timetable.
	slog.Warn("academic timetable format rejected", "reason", reason, "response_bytes", responseBytes, "structure_items", items)
	return codeError("PERSONAL_PARSE_FAILED")
}
