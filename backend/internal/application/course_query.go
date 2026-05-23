package application

import (
	"context"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

type CourseListFilter struct {
	Q           string   `form:"q"`
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
	reviewQuery      review.ReviewQuery
	notificationRepo course.CourseNotificationRepository
	hotRepo          course.HotCourseRepository
}

func NewCourseQueryService(courseQuery course.CourseQuery, reviewQuery review.ReviewQuery, notificationRepo course.CourseNotificationRepository, hotRepo course.HotCourseRepository) *CourseQueryService {
	return &CourseQueryService{
		courseQuery:      courseQuery,
		reviewQuery:      reviewQuery,
		notificationRepo: notificationRepo,
		hotRepo:          hotRepo,
	}
}

func (s *CourseQueryService) GetCourseFilters(ctx context.Context) (*course.CourseFilters, error) {
	return s.courseQuery.GetFilters(ctx)
}

func (s *CourseQueryService) ListCourses(ctx context.Context, f CourseListFilter) (*PaginatedResult[CourseListItemDTO], error) {
	filter := course.CourseFilter{
		Q:           f.Q,
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

func (s *CourseQueryService) ListHotCourses(ctx context.Context, period string, limit int) (*HotCourseListDTO, error) {
	if period == "" {
		period = string(course.HotCoursePeriodWeek)
	}
	hotPeriod := course.HotCoursePeriod(period)
	if hotPeriod != course.HotCoursePeriodWeek && hotPeriod != course.HotCoursePeriodMonth {
		return nil, course.ErrInvalidHotCoursePeriod
	}
	if s.hotRepo == nil {
		return &HotCourseListDTO{Period: string(hotPeriod), Items: []HotCourseItemDTO{}}, nil
	}

	ranks, err := s.hotRepo.Top(ctx, hotPeriod, time.Now(), int64(limit))
	if err != nil {
		return nil, err
	}
	if len(ranks) == 0 {
		return &HotCourseListDTO{Period: string(hotPeriod), Items: []HotCourseItemDTO{}}, nil
	}

	ids := make([]int, 0, len(ranks))
	for _, rank := range ranks {
		ids = append(ids, rank.CourseID)
	}
	courses, _, err := s.courseQuery.FindBy(ctx, course.CourseFilter{CourseIDs: ids})
	if err != nil {
		return nil, err
	}
	courseMap := make(map[int]course.CourseView, len(courses))
	for _, c := range courses {
		courseMap[c.ID] = c
	}

	items := make([]HotCourseItemDTO, 0, len(ranks))
	for _, rank := range ranks {
		c, ok := courseMap[rank.CourseID]
		if !ok {
			continue
		}
		items = append(items, HotCourseItemDTO{
			Course: newCourseListItemDTO(&c),
			Score:  rank.Score,
		})
	}

	return &HotCourseListDTO{Period: string(hotPeriod), Items: items}, nil
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

	if user != nil {
		myReviews, _, err := s.reviewQuery.FindBy(ctx, review.ReviewFilter{
			CourseID: courseID,
			UserID:   user.ID,
			OrderBy:  "created_at",
			Page:     1,
			PageSize: 1,
		})
		if err != nil {
			return nil, err
		}
		if len(myReviews) > 0 {
			myReview := newReviewDTO(&myReviews[0], true)
			dto.MyReview = &myReview
		}
	}

	return dto, nil
}

func (s *CourseQueryService) ListTeacherCourses(ctx context.Context, teacherID int, f CourseListFilter) (*PaginatedResult[CourseListItemDTO], error) {
	filter := course.CourseFilter{
		TeacherID:   teacherID,
		Q:           f.Q,
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
