package application

import (
	"context"

	"jcourse/internal/domain/course"
)

type CourseDTO struct {
	ID          int        `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Credit      float32    `json:"credit"`
	Department  string     `json:"department"`
	MainTeacher TeacherDTO `json:"main_teacher"`
	ReviewCount int        `json:"review_count"`
	AvgRating   float64    `json:"avg_rating"`
}

func newCourseDTO(c *course.CourseForQuery) CourseDTO {
	dto := CourseDTO{
		ID:          c.ID,
		Code:        c.Code,
		Name:        c.Name,
		Credit:      c.Credit,
		Department:  c.Department,
		ReviewCount: c.ReviewCount,
		AvgRating:   c.AvgRating,
		MainTeacher: TeacherDTO{ID: c.MainTeacherID},
	}
	if c.MainTeacher != nil {
		dto.MainTeacher = newTeacherDTO(c.MainTeacher)
	}
	return dto
}

type CourseDetailDTO struct {
	CourseDTO
	RatingDistribution [5]int             `json:"rating_distribution"`
	OtherTeachers      []TeacherCourseDTO `json:"other_teachers"`
	OtherCourses       []CourseDTO        `json:"other_courses"`
}

type TeacherCourseDTO struct {
	CourseID    int        `json:"course_id"`
	Teacher     TeacherDTO `json:"teacher"`
	ReviewCount int        `json:"review_count"`
	AvgRating   float64    `json:"avg_rating"`
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

func (s *CourseQueryService) ListCourses(ctx context.Context, f CourseListFilter) (*PaginatedResult[CourseDTO], error) {
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

	dtos := make([]CourseDTO, len(courses))
	for i, c := range courses {
		dtos[i] = newCourseDTO(&c)
	}

	return &PaginatedResult[CourseDTO]{
		Items:    dtos,
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
		CourseDTO:          newCourseDTO(&detail.CourseForQuery),
		RatingDistribution: detail.RatingDistribution,
		OtherTeachers:      make([]TeacherCourseDTO, len(detail.OtherTeachers)),
		OtherCourses:       make([]CourseDTO, len(detail.OtherCourses)),
	}

	for i, t := range detail.OtherTeachers {
		dto.OtherTeachers[i] = TeacherCourseDTO{
			CourseID:    t.CourseID,
			Teacher:     newTeacherDTO(&t.Teacher),
			ReviewCount: t.ReviewCount,
			AvgRating:   t.AvgRating,
		}
	}

	for i, c := range detail.OtherCourses {
		dto.OtherCourses[i] = newCourseDTO(&c)
	}

	return dto, nil
}

func (s *CourseQueryService) ListTeacherCourses(ctx context.Context, teacherID int, f CourseListFilter) (*PaginatedResult[CourseDTO], error) {
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

	dtos := make([]CourseDTO, len(courses))
	for i, c := range courses {
		dtos[i] = newCourseDTO(&c)
	}

	return &PaginatedResult[CourseDTO]{
		Items:    dtos,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}
