package application

import (
	"context"

	"jcourse/internal/domain/course"
)

type CourseListFilter struct {
	Code        string   `form:"code"`
	Department  string   `form:"department"`
	Language    string   `form:"language"`
	Categories  []string `form:"categories"`
	TargetYears []string `form:"target_years"`
	Credit      *float32 `form:"credit"`
	HasReview   *bool    `form:"has_review"`
	OrderBy     string   `form:"order_by"`
	OrderDir    string   `form:"order_dir"`
	Page        int      `form:"page"`
	PageSize    int      `form:"page_size"`
}

type CourseQueryService struct {
	courseQuery course.CourseQuery
}

func NewCourseQueryService(courseQuery course.CourseQuery) *CourseQueryService {
	return &CourseQueryService{
		courseQuery: courseQuery,
	}
}

func (s *CourseQueryService) ListCourses(ctx context.Context, f CourseListFilter) (*PaginatedResult[CourseListItemDTO], error) {
	filter := course.CourseFilter{
		Code:        f.Code,
		Department:  f.Department,
		Language:    f.Language,
		Categories:  f.Categories,
		TargetYears: f.TargetYears,
		Credit:      f.Credit,
		HasReview:   f.HasReview,
		OrderBy:     f.OrderBy,
		OrderDir:    f.OrderDir,
		Page:        f.Page,
		PageSize:    f.PageSize,
	}

	courses, total, err := s.courseQuery.FindBy(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]CourseListItemDTO, len(courses))
	for i, c := range courses {
		items[i] = newCourseListItemDTO(&c)
	}

	return &PaginatedResult[CourseListItemDTO]{
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
		ID:          detail.ID,
		Code:        detail.Code,
		Name:        detail.Name,
		Credit:      detail.Credit,
		Department:  detail.Department,
		Language:    detail.Language,
		TargetYears: detail.TargetYears,
		Categories:  detail.Categories,
		Rating:      newRatingInfoDTO(detail.Rating),
		MainTeacher: TeacherDTO{ID: detail.MainTeacherID},
	}
	if detail.MainTeacher != nil {
		dto.MainTeacher = newTeacherDTO(detail.MainTeacher)
	}

	dto.OfferedCourses = make([]OfferedCourseDTO, 0, len(detail.OfferedCourses))
	for _, oc := range detail.OfferedCourses {
		ocView := OfferedCourseDTO{
			Semester:     oc.Semester,
			Language:     oc.Language,
			TargetYears:  oc.TargetYears,
			Categories:   oc.Categories,
			TeacherGroup: make([]TeacherDTO, 0, len(oc.TeacherGroup)),
		}
		for i := range oc.TeacherGroup {
			ocView.TeacherGroup = append(ocView.TeacherGroup, newTeacherDTO(&oc.TeacherGroup[i]))
		}
		dto.OfferedCourses = append(dto.OfferedCourses, ocView)
	}

	sameCode, _, err := s.courseQuery.FindBy(ctx, course.CourseFilter{
		Code:      detail.Code,
		ExcludeID: courseID,
		OrderBy:   "rating_avg",
		OrderDir:  "desc",
	})
	if err != nil {
		return nil, err
	}
	dto.OtherTeachers = make([]CourseListItemDTO, len(sameCode))
	for i, c := range sameCode {
		dto.OtherTeachers[i] = newCourseListItemDTO(&c)
	}

	sameTeacher, _, err := s.courseQuery.FindBy(ctx, course.CourseFilter{
		TeacherID: detail.MainTeacherID,
		ExcludeID: courseID,
		OrderBy:   "rating_avg",
		OrderDir:  "desc",
	})
	if err != nil {
		return nil, err
	}
	dto.OtherCourses = make([]CourseListItemDTO, len(sameTeacher))
	for i, c := range sameTeacher {
		dto.OtherCourses[i] = newCourseListItemDTO(&c)
	}

	return dto, nil
}

func (s *CourseQueryService) ListTeacherCourses(ctx context.Context, teacherID int, f CourseListFilter) (*PaginatedResult[CourseListItemDTO], error) {
	filter := course.CourseFilter{
		TeacherID:   teacherID,
		Code:        f.Code,
		Department:  f.Department,
		Language:    f.Language,
		Categories:  f.Categories,
		TargetYears: f.TargetYears,
		Credit:      f.Credit,
		HasReview:   f.HasReview,
		OrderBy:     f.OrderBy,
		OrderDir:    f.OrderDir,
		Page:        f.Page,
		PageSize:    f.PageSize,
	}

	courses, total, err := s.courseQuery.FindBy(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]CourseListItemDTO, len(courses))
	for i, c := range courses {
		items[i] = newCourseListItemDTO(&c)
	}

	return &PaginatedResult[CourseListItemDTO]{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}
