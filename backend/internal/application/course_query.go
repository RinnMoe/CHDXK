package application

import (
	"context"

	"jcourse/internal/domain/course"
)

type CourseListItem struct {
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Credit      float32    `json:"credit"`
	MainTeacher TeacherDTO `json:"main_teacher"`
	RatingCount int        `json:"rating_cnt"`
	RatingAvg   float64    `json:"rating_avg"`
}

func newCourseListItem(c *course.CourseForQuery) CourseListItem {
	item := CourseListItem{
		Code:        c.Code,
		Name:        c.Name,
		Credit:      c.Credit,
		RatingCount: c.ReviewCount,
		RatingAvg:   c.AvgRating,
	}
	if c.MainTeacher != nil {
		item.MainTeacher = newTeacherDTO(c.MainTeacher)
	}
	return item
}

type CourseDetailDTO struct {
	ID                 int              `json:"id"`
	Code               string           `json:"code"`
	Name               string           `json:"name"`
	Credit             float32          `json:"credit"`
	Department         string           `json:"department"`
	MainTeacher        TeacherDTO       `json:"main_teacher"`
	TeacherGroup       []TeacherDTO     `json:"teacher_group"`
	ReviewCount        int              `json:"review_count"`
	AvgRating          float64          `json:"avg_rating"`
	RatingDistribution [5]int           `json:"rating_distribution"`
	OtherTeachers      []CourseListItem `json:"other_teachers"`
	OtherCourses       []CourseListItem `json:"other_courses"`
}

type CourseListFilter struct {
	Code       string   `form:"code"`
	Department string   `form:"department"`
	Credit     *float32 `form:"credit"`
	HasReview  *bool    `form:"has_review"`
	OrderBy    string   `form:"order_by"`
	OrderDir   string   `form:"order_dir"`
	Page       int      `form:"page"`
	PageSize   int      `form:"page_size"`
}

type PaginatedResult[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type CourseQueryService struct {
	courseQuery course.CourseQuery
}

func NewCourseQueryService(courseQuery course.CourseQuery) *CourseQueryService {
	return &CourseQueryService{
		courseQuery: courseQuery,
	}
}

func (s *CourseQueryService) ListCourses(ctx context.Context, f CourseListFilter) (*PaginatedResult[CourseListItem], error) {
	filter := course.CourseFilter{
		Code:       f.Code,
		Department: f.Department,
		Credit:     f.Credit,
		HasReview:  f.HasReview,
		OrderBy:    f.OrderBy,
		OrderDir:   f.OrderDir,
		Page:       f.Page,
		PageSize:   f.PageSize,
	}

	courses, total, err := s.courseQuery.FindBy(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]CourseListItem, len(courses))
	for i, c := range courses {
		items[i] = newCourseListItem(&c)
	}

	return &PaginatedResult[CourseListItem]{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (s *CourseQueryService) GetCourseDetail(ctx context.Context, courseID int) (*CourseDetailDTO, error) {
	detail, err := s.courseQuery.GetDetail(ctx, courseID)
	if err != nil {
		return nil, err
	}

	dto := &CourseDetailDTO{
		ID:                 detail.ID,
		Code:               detail.Code,
		Name:               detail.Name,
		Credit:             detail.Credit,
		Department:         detail.Department,
		ReviewCount:        detail.ReviewCount,
		AvgRating:          detail.AvgRating,
		RatingDistribution: detail.RatingDistribution,
		MainTeacher:        TeacherDTO{ID: detail.MainTeacherID},
	}
	if detail.MainTeacher != nil {
		dto.MainTeacher = newTeacherDTO(detail.MainTeacher)
	}

	dto.TeacherGroup = make([]TeacherDTO, 0, len(detail.TeacherGroup))
	for _, t := range detail.TeacherGroup {
		dto.TeacherGroup = append(dto.TeacherGroup, newTeacherDTO(t))
	}

	sameCode, _, err := s.courseQuery.FindBy(ctx, course.CourseFilter{
		Code:      detail.Code,
		ExcludeID: courseID,
		OrderBy:   "avg_rating",
		OrderDir:  "desc",
	})
	if err != nil {
		return nil, err
	}
	dto.OtherTeachers = make([]CourseListItem, len(sameCode))
	for i, c := range sameCode {
		dto.OtherTeachers[i] = newCourseListItem(&c)
	}

	sameTeacher, _, err := s.courseQuery.FindBy(ctx, course.CourseFilter{
		TeacherID: detail.MainTeacherID,
		ExcludeID: courseID,
		OrderBy:   "avg_rating",
		OrderDir:  "desc",
	})
	if err != nil {
		return nil, err
	}
	dto.OtherCourses = make([]CourseListItem, len(sameTeacher))
	for i, c := range sameTeacher {
		dto.OtherCourses[i] = newCourseListItem(&c)
	}

	return dto, nil
}

func (s *CourseQueryService) ListTeacherCourses(ctx context.Context, teacherID int, f CourseListFilter) (*PaginatedResult[CourseListItem], error) {
	filter := course.CourseFilter{
		TeacherID:  teacherID,
		Code:       f.Code,
		Department: f.Department,
		Credit:     f.Credit,
		HasReview:  f.HasReview,
		OrderBy:    f.OrderBy,
		OrderDir:   f.OrderDir,
		Page:       f.Page,
		PageSize:   f.PageSize,
	}

	courses, total, err := s.courseQuery.FindBy(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]CourseListItem, len(courses))
	for i, c := range courses {
		items[i] = newCourseListItem(&c)
	}

	return &PaginatedResult[CourseListItem]{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}
