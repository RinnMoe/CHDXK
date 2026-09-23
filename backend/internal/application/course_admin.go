package application

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"

	"jcourse/internal/domain/course"
	"jcourse/pkg/apperr"
)

const maxCourseArchiveSize = 64 << 20

var semesterPattern = regexp.MustCompile(`^\d{4}-\d{4}-\d+$`)

type CourseManagementService struct {
	repo course.CourseDataRepository
}

func NewCourseManagementService(repo course.CourseDataRepository) *CourseManagementService {
	return &CourseManagementService{repo: repo}
}

type CourseImportResult struct {
	Semester      string `json:"semester"`
	CourseCount   int    `json:"course_count"`
	SectionCount  int    `json:"section_count"`
	ReplacedCount int    `json:"replaced_count"`
}

func (s *CourseManagementService) ListSemesters(ctx context.Context) ([]string, error) {
	return s.repo.ListCourseSemesters(ctx)
}

func (s *CourseManagementService) List(ctx context.Context, filter course.ManagedOfferingFilter) (*PaginatedResult[course.ManagedOffering], error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 50
	}
	items, total, err := s.repo.ListManagedOfferings(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &PaginatedResult[course.ManagedOffering]{Items: items, Total: total, Page: filter.Page, PageSize: filter.PageSize}, nil
}

func (s *CourseManagementService) ImportArchive(ctx context.Context, archive []byte) (*CourseImportResult, error) {
	semester, offerings, err := parseCourseArchive(archive)
	if err != nil {
		return nil, err
	}
	oldCount, err := s.countSemesterOfferings(ctx, semester)
	if err != nil {
		return nil, err
	}
	if err := s.repo.ReplaceSemester(ctx, semester, offerings); err != nil {
		return nil, err
	}
	sectionCount := 0
	for _, offering := range offerings {
		sectionCount += len(offering.Sections)
	}
	return &CourseImportResult{
		Semester:      semester,
		CourseCount:   len(offerings),
		SectionCount:  sectionCount,
		ReplacedCount: oldCount,
	}, nil
}

func (s *CourseManagementService) countSemesterOfferings(ctx context.Context, semester string) (int, error) {
	_, total, err := s.repo.ListManagedOfferings(ctx, course.ManagedOfferingFilter{
		Semester: semester,
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

func (s *CourseManagementService) Save(ctx context.Context, id int, offering course.ManagedOffering) (*course.ManagedOffering, error) {
	cleaned, err := validateManagedOffering(offering)
	if err != nil {
		return nil, err
	}
	return s.repo.SaveManagedOffering(ctx, id, cleaned)
}

func (s *CourseManagementService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return apperr.BadRequest("开课记录 ID 无效")
	}
	return s.repo.DeleteManagedOffering(ctx, id)
}

type chdSemesterDocument struct {
	SchemaVersion int               `json:"schema_version"`
	Source        chdDocumentSource `json:"source"`
	Semester      string            `json:"semester"`
	ExpectedCount *int              `json:"expected_count"`
	Lessons       []chdLesson       `json:"lessons"`
}

type chdDocumentSource struct {
	Host string `json:"host"`
}

type chdLesson struct {
	ID      string `json:"id"`
	Summary struct {
		LessonNo     string   `json:"lesson_no"`
		CourseCode   string   `json:"course_code"`
		CourseName   string   `json:"course_name"`
		CourseType   string   `json:"course_type"`
		TeachClass   string   `json:"teach_class"`
		Teachers     []string `json:"teachers"`
		ActualCount  *int     `json:"actual_count"`
		LimitCount   *int     `json:"limit_count"`
		Credit       *float32 `json:"credit"`
		CoursePeriod string   `json:"course_period"`
		WeekRange    string   `json:"week_range"`
	} `json:"summary"`
	Detail *struct {
		Basic struct {
			Semester   string   `json:"semester"`
			CourseCode string   `json:"course_code"`
			CourseName string   `json:"course_name"`
			Credit     *float32 `json:"credit"`
			CourseType string   `json:"course_type"`
			Department string   `json:"department"`
			Campus     string   `json:"campus"`
			Language   string   `json:"language"`
			Assessment string   `json:"assessment"`
			Teachers   []string `json:"teachers"`
		} `json:"basic"`
		Arrangement struct {
			TotalHours  *float32 `json:"total_hours"`
			WeeklyHours *float32 `json:"weekly_hours"`
			WeekRange   string   `json:"week_range"`
			StartWeek   *int     `json:"start_week"`
			EndWeek     *int     `json:"end_week"`
			WeekCount   *int     `json:"week_count"`
		} `json:"arrangement"`
		Remark string `json:"remark"`
	} `json:"detail"`
	Links struct {
		Syllabus string `json:"syllabus"`
	} `json:"links"`
}

type chdImportReport struct {
	SchemaVersion      int      `json:"schema_version"`
	Semester           string   `json:"semester"`
	Completed          bool     `json:"completed"`
	ListComplete       bool     `json:"list_complete"`
	ExpectedCount      *int     `json:"expected_count"`
	LessonCount        int      `json:"lesson_count"`
	DetailSuccessCount int      `json:"detail_success_count"`
	MissingDetailCount int      `json:"missing_detail_count"`
	DuplicateLessonIDs []string `json:"duplicate_lesson_ids"`
	FailedLessonIDs    []string `json:"failed_lesson_ids"`
	SourcePage         struct {
		Host string `json:"host"`
	} `json:"source_page"`
}

func parseCourseArchive(data []byte) (string, []course.ManagedOffering, error) {
	if len(data) == 0 || len(data) > maxCourseArchiveSize {
		return "", nil, apperr.BadRequest("数据 ZIP 为空或超过 64 MiB")
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", nil, apperr.BadRequest("无法读取课程数据 ZIP")
	}
	var documentData, reportData []byte
	var totalDataSize int64
	for _, file := range archive.File {
		base := path.Base(file.Name)
		if base != "semester.json" && base != "report.json" {
			continue
		}
		if file.UncompressedSize64 > maxCourseArchiveSize {
			return "", nil, apperr.BadRequest("ZIP 中的课程数据文件过大")
		}
		reader, err := file.Open()
		if err != nil {
			return "", nil, apperr.BadRequest("无法读取 ZIP 中的课程数据文件")
		}
		content, readErr := io.ReadAll(io.LimitReader(reader, maxCourseArchiveSize+1))
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil || len(content) > maxCourseArchiveSize {
			return "", nil, apperr.BadRequest("ZIP 中的课程数据文件读取失败或过大")
		}
		totalDataSize += int64(len(content))
		if totalDataSize > maxCourseArchiveSize {
			return "", nil, apperr.BadRequest("ZIP 中的课程数据总量超过 64 MiB")
		}
		switch base {
		case "semester.json":
			if documentData != nil {
				return "", nil, apperr.BadRequest("ZIP 中包含多个 semester.json")
			}
			documentData = content
		case "report.json":
			if reportData != nil {
				return "", nil, apperr.BadRequest("ZIP 中包含多个 report.json")
			}
			reportData = content
		}
	}
	if documentData == nil || reportData == nil {
		return "", nil, apperr.BadRequest("ZIP 必须同时包含 semester.json 和 report.json")
	}

	var document chdSemesterDocument
	if err := json.Unmarshal(documentData, &document); err != nil {
		return "", nil, apperr.BadRequest("semester.json 格式无效")
	}
	var report chdImportReport
	if err := json.Unmarshal(reportData, &report); err != nil {
		return "", nil, apperr.BadRequest("report.json 格式无效")
	}
	if document.SchemaVersion != 1 || report.SchemaVersion != 1 {
		return "", nil, apperr.BadRequest("仅支持 CHD 课程数据格式版本 1")
	}
	if !semesterPattern.MatchString(document.Semester) || document.Semester != report.Semester {
		return "", nil, apperr.BadRequest("课程数据与采集报告的学期不匹配")
	}
	semesterParts := strings.Split(document.Semester, "-")
	expectedDetailSemester := semesterParts[0] + "-" + semesterParts[1] + "学年" + semesterParts[2] + "学期"
	if !isCHDHost(document.Source.Host) || !isCHDHost(report.SourcePage.Host) {
		return "", nil, apperr.BadRequest("课程数据来源必须是 CHD 教务系统")
	}
	if !report.Completed || !report.ListComplete || report.MissingDetailCount != 0 ||
		len(report.DuplicateLessonIDs) != 0 || len(report.FailedLessonIDs) != 0 {
		return "", nil, apperr.BadRequest("采集报告未通过完整性检查")
	}
	if report.ExpectedCount == nil || *report.ExpectedCount != len(document.Lessons) ||
		report.LessonCount != len(document.Lessons) || report.DetailSuccessCount != len(document.Lessons) {
		return "", nil, apperr.BadRequest("课程记录数量与采集报告不一致")
	}
	if document.ExpectedCount != nil && *document.ExpectedCount != len(document.Lessons) {
		return "", nil, apperr.BadRequest("课程记录数量与数据文件不一致")
	}
	if len(document.Lessons) == 0 {
		return "", nil, apperr.BadRequest("课程数据中没有课节记录")
	}

	seenIDs := make(map[string]struct{}, len(document.Lessons))
	groups := make(map[string]*course.ManagedOffering)
	order := make([]string, 0)
	for _, lesson := range document.Lessons {
		lessonID := strings.TrimSpace(lesson.ID)
		if lessonID == "" {
			return "", nil, apperr.BadRequest("课节 ID 不能为空")
		}
		if _, ok := seenIDs[lessonID]; ok {
			return "", nil, apperr.BadRequest("课程数据包含重复课节 ID")
		}
		seenIDs[lessonID] = struct{}{}
		if lesson.Detail == nil {
			return "", nil, apperr.BadRequest("课程数据包含详情缺失的课节")
		}
		basic := lesson.Detail.Basic
		if strings.Join(strings.Fields(basic.Semester), "") != expectedDetailSemester {
			return "", nil, apperr.BadRequest("课节详情中的学期与课程数据不一致")
		}
		code := strings.TrimSpace(basic.CourseCode)
		if code == "" {
			code = strings.TrimSpace(lesson.Summary.CourseCode)
		}
		name := firstNonEmpty(basic.CourseName, lesson.Summary.CourseName)
		department := strings.TrimSpace(basic.Department)
		credit := basic.Credit
		if credit == nil {
			credit = lesson.Summary.Credit
		}
		if code == "" || name == "" || department == "" || credit == nil {
			return "", nil, apperr.BadRequest("课程代码、名称、学分和开课院系不能为空")
		}
		if *credit < 0 {
			return "", nil, apperr.BadRequest("课程学分不能小于 0")
		}
		if summaryCode := strings.TrimSpace(lesson.Summary.CourseCode); summaryCode != "" && summaryCode != code {
			return "", nil, apperr.BadRequest("课程列表与详情中的课程代码不一致")
		}
		teacherNames := basic.Teachers
		if len(teacherNames) == 0 {
			teacherNames = lesson.Summary.Teachers
		}
		teachers := cleanTeacherNames(teacherNames)
		mainKey := ""
		if len(teachers) > 0 {
			mainKey = normalizedTeacherIdentity(department, teachers[0])
		}
		groupKey := code + "\x00" + mainKey
		group, ok := groups[groupKey]
		if !ok {
			group = &course.ManagedOffering{Code: code, Semester: document.Semester}
			groups[groupKey] = group
			order = append(order, groupKey)
		}
		courseType := firstNonEmpty(basic.CourseType, lesson.Summary.CourseType)
		weekRange := firstNonEmpty(lesson.Detail.Arrangement.WeekRange, lesson.Summary.WeekRange)
		syllabusURL, err := canonicalSyllabusURL(lesson.Links.Syllabus)
		if err != nil {
			return "", nil, err
		}
		group.Sections = append(group.Sections, course.CourseSection{
			SourceLessonID: lessonID,
			LessonNo:       strings.TrimSpace(lesson.Summary.LessonNo),
			CourseName:     name,
			CourseType:     courseType,
			Credit:         *credit,
			Department:     department,
			Language:       strings.TrimSpace(basic.Language),
			Campus:         strings.TrimSpace(basic.Campus),
			Assessment:     strings.TrimSpace(basic.Assessment),
			TeachClass:     strings.TrimSpace(lesson.Summary.TeachClass),
			ActualCount:    lesson.Summary.ActualCount,
			LimitCount:     lesson.Summary.LimitCount,
			CoursePeriod:   strings.TrimSpace(lesson.Summary.CoursePeriod),
			TotalHours:     lesson.Detail.Arrangement.TotalHours,
			WeeklyHours:    lesson.Detail.Arrangement.WeeklyHours,
			WeekRange:      weekRange,
			StartWeek:      lesson.Detail.Arrangement.StartWeek,
			EndWeek:        lesson.Detail.Arrangement.EndWeek,
			WeekCount:      lesson.Detail.Arrangement.WeekCount,
			Remark:         strings.TrimSpace(lesson.Detail.Remark),
			SyllabusURL:    syllabusURL,
			Teachers:       teachers,
		})
	}

	offerings := make([]course.ManagedOffering, 0, len(order))
	for _, key := range order {
		offerings = append(offerings, *groups[key])
	}
	sort.SliceStable(offerings, func(i, j int) bool {
		if offerings[i].Code == offerings[j].Code {
			return mainTeacherSortKey(offerings[i]) < mainTeacherSortKey(offerings[j])
		}
		return offerings[i].Code < offerings[j].Code
	})
	return document.Semester, offerings, nil
}

func validateManagedOffering(offering course.ManagedOffering) (course.ManagedOffering, error) {
	offering.Code = strings.TrimSpace(offering.Code)
	offering.Semester = strings.TrimSpace(offering.Semester)
	if offering.Code == "" || !semesterPattern.MatchString(offering.Semester) || len(offering.Sections) == 0 {
		return course.ManagedOffering{}, apperr.BadRequest("课程代码、学期和至少一个教学班为必填项")
	}
	seenIDs := map[string]struct{}{}
	mainIdentity := ""
	for i := range offering.Sections {
		section := &offering.Sections[i]
		section.SourceLessonID = strings.TrimSpace(section.SourceLessonID)
		section.LessonNo = strings.TrimSpace(section.LessonNo)
		section.CourseName = strings.TrimSpace(section.CourseName)
		section.CourseType = strings.TrimSpace(section.CourseType)
		section.Department = strings.TrimSpace(section.Department)
		section.Language = strings.TrimSpace(section.Language)
		section.Campus = strings.TrimSpace(section.Campus)
		section.Assessment = strings.TrimSpace(section.Assessment)
		section.TeachClass = strings.TrimSpace(section.TeachClass)
		section.CoursePeriod = strings.TrimSpace(section.CoursePeriod)
		section.WeekRange = strings.TrimSpace(section.WeekRange)
		section.Remark = strings.TrimSpace(section.Remark)
		if section.CourseName == "" || section.Department == "" || section.Credit < 0 {
			return course.ManagedOffering{}, apperr.BadRequest("课程名称、开课院系不能为空，学分不能小于 0")
		}
		if section.SourceLessonID != "" {
			if _, exists := seenIDs[section.SourceLessonID]; exists {
				return course.ManagedOffering{}, apperr.BadRequest("同一学期内课节 ID 不能重复")
			}
			seenIDs[section.SourceLessonID] = struct{}{}
		}
		section.Teachers = cleanTeacherNames(section.Teachers)
		identity := ""
		if len(section.Teachers) > 0 {
			identity = normalizedTeacherIdentity(section.Department, section.Teachers[0])
		}
		if i == 0 {
			mainIdentity = identity
		} else if identity != mainIdentity {
			return course.ManagedOffering{}, apperr.BadRequest("一个开课记录内的教学班必须使用相同主讲教师")
		}
		syllabusURL, err := canonicalSyllabusURL(section.SyllabusURL)
		if err != nil {
			return course.ManagedOffering{}, err
		}
		section.SyllabusURL = syllabusURL
	}
	return offering, nil
}

func canonicalSyllabusURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Fragment != "" {
		return "", apperr.BadRequest("课程大纲链接无效")
	}
	if u.Scheme != "" || u.Host != "" {
		if (u.Scheme != "http" && u.Scheme != "https") || !strings.EqualFold(u.Hostname(), "bkjw.chd.edu.cn") || u.Port() != "" {
			return "", apperr.BadRequest("课程大纲链接只能指向 CHD 教务系统")
		}
	} else if !strings.HasPrefix(u.Path, "/eams/") {
		return "", apperr.BadRequest("课程大纲链接路径无效")
	}
	cleanPath := path.Clean(u.Path)
	if !strings.HasPrefix(cleanPath, "/eams/") || strings.Contains(u.Path, "\\") {
		return "", apperr.BadRequest("课程大纲链接路径无效")
	}
	return "http://bkjw.chd.edu.cn" + u.RequestURI(), nil
}

func isCHDHost(host string) bool {
	host = strings.TrimSpace(strings.ToLower(host))
	if strings.Contains(host, "://") {
		u, err := url.Parse(host)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Port() != "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			return false
		}
		host = u.Hostname()
	}
	return host == "bkjw.chd.edu.cn"
}

func cleanTeacherNames(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	result := make([]string, 0, len(names))
	for _, name := range names {
		name = course.NormalizeTeacherName(name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, name)
	}
	return result
}

func normalizedTeacherIdentity(department, name string) string {
	return course.InternalTeacherCode(department, name)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func mainTeacherSortKey(offering course.ManagedOffering) string {
	if len(offering.Sections) == 0 || len(offering.Sections[0].Teachers) == 0 {
		return ""
	}
	return normalizedTeacherIdentity(offering.Sections[0].Department, offering.Sections[0].Teachers[0])
}
