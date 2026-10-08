//go:build test

package bkjw

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"jcourse/internal/domain/course"
	"jcourse/internal/infrastructure/repository"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Opt-in Linux/container capacity probe. No requests leave the local fixture server.
// Large names + maximum 1000 classes stress HTTP, parsing and archive retention.
func TestCollectorCapacity(t *testing.T) {
	raw := os.Getenv("BKJW_LOAD_CONCURRENCY")
	if raw == "" {
		t.Skip("opt-in capacity probe")
	}
	concurrency, err := strconv.Atoi(raw)
	if err != nil || concurrency < 1 || concurrency > 256 {
		t.Fatal("invalid concurrency")
	}
	var catalog strings.Builder
	catalog.WriteString("{semesters:{y0:[")
	for i := 0; i < 17; i++ {
		year := 2018 + i/2
		fmt.Fprintf(&catalog, "{id:%d,schoolYear:%d-%d,name:%d},", i+1, year, year+1, 1+i%2)
	}
	catalog.WriteString("]}}")
	var payload strings.Builder
	classCount := 1000
	var importer *repository.CourseEnrollmentRepository
	var fixtureDB *gorm.DB
	if os.Getenv("BKJW_LOAD_DB") == "1" {
		classCount = 30
		dsn := os.Getenv("JCOURSE_POSTGRES_DSN")
		if dsn == "" {
			t.Fatal("isolated database DSN required")
		}
		base, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatal("database connection unavailable")
		}
		schema := fmt.Sprintf("bkjw_capacity_%d", time.Now().UnixNano())
		if err = base.Exec("CREATE SCHEMA " + schema).Error; err != nil {
			t.Fatal(err)
		}
		defer base.Exec("DROP SCHEMA " + schema + " CASCADE")
		fixtureDB, err = gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatal("fixture database unavailable")
		}
		sqlDB, _ := fixtureDB.DB()
		sqlDB.SetMaxOpenConns(6)
		sqlDB.SetMaxIdleConns(2)
		defer sqlDB.Close()
		for _, statement := range []string{
			"CREATE TABLE offered_courses(id integer PRIMARY KEY, course_id integer NOT NULL, semester text NOT NULL)",
			"CREATE TABLE course_sections(id serial PRIMARY KEY,offered_course_id integer, source_lesson_id text, course_name text, lesson_no text)",
			"CREATE INDEX ON course_sections(source_lesson_id,course_name)",
			"CREATE TABLE course_enrollments(id serial PRIMARY KEY,user_id integer,course_id integer,semester text,created_at timestamptz,UNIQUE(user_id,course_id,semester))",
		} {
			if err = fixtureDB.Exec(statement).Error; err != nil {
				t.Fatal(err)
			}
		}
		for term := 0; term < 17; term++ {
			year := 2018 + term/2
			name := fmt.Sprintf("%d-%d-%d", year, year+1, 1+term%2)
			for i := 0; i < classCount; i++ {
				offering := term*classCount + i + 1
				if err = fixtureDB.Exec("INSERT INTO offered_courses VALUES (?,?,?)", offering, i+1, name).Error; err != nil {
					t.Fatal(err)
				}
				if err = fixtureDB.Exec("INSERT INTO course_sections(offered_course_id,source_lesson_id,course_name) VALUES (?,?,?)", offering, strconv.Itoa(i+1), strings.Repeat("模拟课程", 80)+strconv.Itoa(i)).Error; err != nil {
					t.Fatal(err)
				}
			}
		}
		importer = repository.NewCourseEnrollmentRepository(fixtureDB)
	}
	payload.WriteString("<script>var table0=new CourseTable();")
	for i := 0; i < classCount; i++ {
		fmt.Fprintf(&payload, "var activity=new TaskActivity('', '', '%d(CS%d)', '%s%d', '', '', '0111');", i+1, i+1, strings.Repeat("模拟课程", 80), i)
	}
	payload.WriteString("</script>")
	body := payload.String()
	calendar := catalog.String()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Host + r.URL.Path {
		case "ids.chd.edu.cn/authserver/login":
			http.Redirect(w, r, eamsOrigin+"/eams/home.action", 302)
		case "bkjw.chd.edu.cn/eams/home.action":
			fmt.Fprint(w, "<html>教务首页</html>")
		case "bkjw.chd.edu.cn/eams/dataQuery.action":
			fmt.Fprint(w, calendar)
		case "bkjw.chd.edu.cn/eams/courseTableForStd!courseTable.action":
			fmt.Fprint(w, body)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	// start collectors after a simulated completed login; QR polling is covered separately.
	var finished atomic.Int32
	var records atomic.Int64
	var wg sync.WaitGroup
	reports := make(chan []course.EnrollmentImportResult, concurrency)
	barrier := make(chan struct{})
	started := time.Now()
	for worker := 0; worker < concurrency; worker++ {
		owner := worker + 1
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			jar, _ := cookiejar.New(nil)
			c := &connection{ctx: ctx, client: &http.Client{Jar: jar, Transport: mappedTransport{server.URL}, Timeout: 30 * time.Second, CheckRedirect: checkRedirect}, events: make(chan course.SourceEvent)}
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case <-c.events:
					}
				}
			}()
			<-barrier
			if err := c.completeLogin(casOrigin + "/authserver/login"); err != nil {
				t.Error(err)
				return
			}
			options, err := c.semesters([]byte(entryPage))
			if err != nil {
				t.Error(err)
				return
			}
			retained := []course.EnrollmentImportResult{}
			for _, s := range options {
				event, err := c.personalArchive([]byte(entryPage), s)
				if err != nil {
					t.Error(err)
					return
				}
				if classCount == 1000 && len(event.Document) < 900000 {
					t.Error("fixture too small")
					return
				}
				records.Add(int64(classCount))
				var decoded struct {
					Courses []course.EnrollmentSourceCourse `json:"courses"`
				}
				if err := json.Unmarshal(event.Document, &decoded); err != nil || len(decoded.Courses) != classCount {
					t.Error("invalid archive")
					return
				}
				result := course.EnrollmentImportResult{UnmatchedCount: len(decoded.Courses), Unmatched: make([]course.EnrollmentSourceCourse, 20)}
				if importer != nil {
					imported, e := importer.ImportSemester(ctx, owner, s.Name, decoded.Courses)
					if e != nil || imported.Matched != classCount || imported.Added != classCount {
						t.Errorf("database import failed: %v", e)
						return
					}
					result = *imported
				}
				for i, item := range decoded.Courses[:len(result.Unmatched)] {
					item.Name = string([]rune(item.Name)[:128]) + "…"
					result.Unmatched[i] = item
				}
				retained = append(retained, result)
				time.Sleep(30 * time.Millisecond)
				runtime.KeepAlive(event)
			}
			c.client.CloseIdleConnections()
			c.client.Jar = nil
			finished.Add(1)
			reports <- retained
		}()
	}
	close(barrier)
	wg.Wait()
	time.Sleep(5 * time.Second)
	runtime.KeepAlive(reports)
	if finished.Load() != int32(concurrency) {
		t.Fatalf("completed %d/%d", finished.Load(), concurrency)
	}
	if fixtureDB != nil {
		var saved int64
		if err := fixtureDB.Table("course_enrollments").Count(&saved).Error; err != nil || saved != int64(concurrency*17*classCount) {
			t.Fatal("incorrect saved count", saved, err)
		}
	}
	t.Logf("concurrency=%d semesters=17 classes_per_semester=%d database=%v body_bytes=%d completed=%d records=%d elapsed=%s", concurrency, classCount, importer != nil, len(body), finished.Load(), records.Load(), time.Since(started))
}
