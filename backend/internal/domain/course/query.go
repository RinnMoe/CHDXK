package course

import (
	"context"

	"jcourse/internal/domain/teacher"
)

type RatingInfo struct {
	Count        int
	Avg          float64
	Distribution [5]int
}

type CourseFilter struct {
	CourseIDs   []int
	TeacherID   int
	ExcludeID   int
	Code        string
	Department  string
	Categories  []string
	Language    string
	TargetYears []string
	Credit      *float32
	HasReview   *bool
	OrderBy     string // "rating_count" | "rating_avg"
	Ascend      bool
	Page        int
	PageSize    int
}

// Read model: course list/search result
type CourseView struct {
	ID            int
	Code          string
	Name          string
	Credit        float32
	Department    string
	MainTeacherID int
	MainTeacher   *teacher.TeacherView
	Categories    []string
	Language      string
	TargetYears   []string
	Rating        RatingInfo
}

// Read model: course detail with offered courses and rating distribution
type CourseDetailView struct {
	ID                int
	Code              string
	Name              string
	Credit            float32
	Department        string
	MainTeacherID     int
	MainTeacher       *teacher.TeacherView
	Categories        []string
	Language          string
	TargetYears       []string
	Rating            RatingInfo
	OfferedCourses    []OfferedCourseView
	NotificationLevel NotificationLevel
}

type FilterItem = teacher.FilterItem

type CourseFilters struct {
	Credits     []FilterItem `json:"credits"`
	Departments []FilterItem `json:"departments"`
	Categories  []FilterItem `json:"categories"`
	TargetYears []FilterItem `json:"target_years"`
}

// Read model interface
type CourseQuery interface {
	FindBy(ctx context.Context, filter CourseFilter) ([]CourseView, int64, error)
	GetDetail(ctx context.Context, courseID int) (*CourseDetailView, error)
	FindOfferedCourses(ctx context.Context, courseID int) ([]OfferedCourseView, error)
	GetFilters(ctx context.Context) (*CourseFilters, error)
}
