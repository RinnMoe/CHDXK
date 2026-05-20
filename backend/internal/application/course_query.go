package application

import (
	"context"

	"jcourse/internal/domain/course"
)

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

	dto.OfferedCourses = make([]OfferedCourseDTO, 0, len(detail.OfferedCourses))
	for _, oc := range detail.OfferedCourses {
		ocDTO := OfferedCourseDTO{
			Semester:     oc.Semester,
			Language:     oc.Language,
			Grade:        oc.Grade,
			TeacherGroup: make([]TeacherDTO, 0, len(oc.TeacherGroup)),
		}
		for _, t := range oc.TeacherGroup {
			ocDTO.TeacherGroup = append(ocDTO.TeacherGroup, newTeacherDTO(t))
		}
		dto.OfferedCourses = append(dto.OfferedCourses, ocDTO)
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
