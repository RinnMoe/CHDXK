package application

import (
	"context"

	"jcourse/internal/domain/auth"
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
	Ascend      bool     `form:"ascend"`
	Page        int      `form:"page"`
	PageSize    int      `form:"page_size"`
}

type CourseQueryService struct {
	courseQuery      course.CourseQuery
	notificationRepo course.CourseNotificationRepository
}

func NewCourseQueryService(courseQuery course.CourseQuery, notificationRepo course.CourseNotificationRepository) *CourseQueryService {
	return &CourseQueryService{
		courseQuery:      courseQuery,
		notificationRepo: notificationRepo,
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
		Ascend:      f.Ascend,
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

func (s *CourseQueryService) GetCourseDetail(ctx context.Context, user *auth.User, courseID int) (*CourseDetailDTO, error) {
	detail, err := s.courseQuery.GetDetail(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, ErrCourseNotFound
	}

	if user != nil {
		level, err := s.notificationRepo.GetLevel(ctx, user.ID, courseID)
		if err != nil {
			return nil, err
		}
		detail.NotificationLevel = level
	}

	dto := &CourseDetailDTO{
		ID:                detail.ID,
		Code:              detail.Code,
		Name:              detail.Name,
		Credit:            detail.Credit,
		Department:        detail.Department,
		Language:          detail.Language,
		TargetYears:       detail.TargetYears,
		Categories:        detail.Categories,
		Rating:            newRatingInfoDTO(detail.Rating),
		MainTeacher:       TeacherDTO{ID: detail.MainTeacherID},
		NotificationLevel: int(detail.NotificationLevel),
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

	sameCodeCourses, _, err := s.courseQuery.FindBy(ctx, course.CourseFilter{
		Code:      detail.Code,
		ExcludeID: courseID,
		OrderBy:   "rating_avg",
	})
	if err != nil {
		return nil, err
	}
	dto.SameCodeCourses = make([]CourseListItemDTO, len(sameCodeCourses))
	for i, c := range sameCodeCourses {
		dto.SameCodeCourses[i] = newCourseListItemDTO(&c)
	}

	sameTeacherCourses, _, err := s.courseQuery.FindBy(ctx, course.CourseFilter{
		TeacherID: detail.MainTeacherID,
		ExcludeID: courseID,
		OrderBy:   "rating_avg",
	})
	if err != nil {
		return nil, err
	}
	dto.SameTeacherCourses = make([]CourseListItemDTO, len(sameTeacherCourses))
	for i, c := range sameTeacherCourses {
		dto.SameTeacherCourses[i] = newCourseListItemDTO(&c)
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
		Ascend:      f.Ascend,
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

func (s *CourseQueryService) ListCoursesByNotificationLevel(ctx context.Context, userID int, level course.NotificationLevel, f CourseListFilter) (*PaginatedResult[CourseListItemDTO], error) {
	courseIDs, err := s.notificationRepo.GetCoursesByLevel(ctx, userID, level)
	if err != nil {
		return nil, err
	}
	if len(courseIDs) == 0 {
		return &PaginatedResult[CourseListItemDTO]{
			Items:    []CourseListItemDTO{},
			Total:    0,
			Page:     f.Page,
			PageSize: f.PageSize,
		}, nil
	}

	filter := course.CourseFilter{
		CourseIDs: courseIDs,
		OrderBy:   f.OrderBy,
		Ascend:    f.Ascend,
		Page:      f.Page,
		PageSize:  f.PageSize,
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
