package course

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// CourseSection contains the CHD course fields captured for one teaching section.
type CourseSection struct {
	ID              int                    `json:"id,omitempty"`
	SourceLessonID  string                 `json:"source_lesson_id,omitempty"`
	LessonNo        string                 `json:"lesson_no"`
	CourseName      string                 `json:"course_name"`
	CourseType      string                 `json:"course_type"`
	Credit          float32                `json:"credit"`
	Department      string                 `json:"department"`
	Language        string                 `json:"language"`
	Campus          string                 `json:"campus"`
	Assessment      string                 `json:"assessment"`
	TeachClass      string                 `json:"teach_class"`
	ActualCount     *int                   `json:"actual_count,omitempty"`
	LimitCount      *int                   `json:"limit_count,omitempty"`
	CoursePeriod    string                 `json:"course_period"`
	TotalHours      *float32               `json:"total_hours,omitempty"`
	WeeklyHours     *float32               `json:"weekly_hours,omitempty"`
	WeekRange       string                 `json:"week_range"`
	StartWeek       *int                   `json:"start_week,omitempty"`
	EndWeek         *int                   `json:"end_week,omitempty"`
	WeekCount       *int                   `json:"week_count,omitempty"`
	Remark          string                 `json:"remark"`
	SyllabusURL     string                 `json:"syllabus_url"`
	Teachers        []string               `json:"teachers"`
	TeacherProfiles []CourseSectionTeacher `json:"teacher_profiles,omitempty"`
}

type CourseSectionTeacher struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Position int    `json:"position"`
}

// ManagedOffering is the admin write/read model for one course in one semester.
type ManagedOffering struct {
	ID       int             `json:"id,omitempty"`
	CourseID int             `json:"course_id,omitempty"`
	Code     string          `json:"code"`
	Semester string          `json:"semester"`
	Sections []CourseSection `json:"sections"`
}

type ManagedOfferingFilter struct {
	Semester string
	Query    string
	Page     int
	PageSize int
}

type CourseDataRepository interface {
	ListCourseSemesters(ctx context.Context) ([]string, error)
	ListManagedOfferings(ctx context.Context, filter ManagedOfferingFilter) ([]ManagedOffering, int64, error)
	ReplaceSemester(ctx context.Context, semester string, offerings []ManagedOffering) error
	SaveManagedOffering(ctx context.Context, id int, offering ManagedOffering) (*ManagedOffering, error)
	DeleteManagedOffering(ctx context.Context, id int) error
}

func NormalizeTeacherName(name string) string {
	return strings.Join(strings.Fields(norm.NFKC.String(name)), " ")
}

func InternalTeacherCode(department, name string) string {
	department = strings.ToLower(strings.Join(strings.Fields(norm.NFKC.String(department)), " "))
	name = strings.ToLower(NormalizeTeacherName(name))
	sum := sha256.Sum256([]byte(department + "\x00" + name))
	return "CHD-" + hex.EncodeToString(sum[:])
}
