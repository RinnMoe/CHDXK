package application

import (
	"context"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/teacher"
	"time"
)

type HotTeacherItemDTO struct {
	Teacher     TeacherDTO `json:"teacher"`
	Score       int64      `json:"score"`
	CourseCount int        `json:"course_count"`
}
type HotTeacherListDTO struct {
	Period    string              `json:"period"`
	PeriodKey string              `json:"period_key"`
	Items     []HotTeacherItemDTO `json:"items"`
}
type TeacherHotQueryService struct{ query teacher.HotTeacherQuery }

func NewTeacherHotQueryService(query teacher.HotTeacherQuery) *TeacherHotQueryService {
	return &TeacherHotQueryService{query: query}
}
func (s *TeacherHotQueryService) List(ctx context.Context, period string, limit int) (*HotTeacherListDTO, error) {
	if period == "" {
		period = "week"
	}
	loc, err := course.DefaultHotCourseLocation()
	if err != nil {
		return nil, err
	}
	p, err := course.NewHotCoursePeriod(course.HotCoursePeriodName(period), time.Now(), loc)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.query.TopTeachers(ctx, string(p.Period), p.PeriodKey, limit)
	if err != nil {
		return nil, err
	}
	result := &HotTeacherListDTO{Period: string(p.Period), PeriodKey: p.PeriodKey, Items: make([]HotTeacherItemDTO, len(rows))}
	for i, row := range rows {
		result.Items[i] = HotTeacherItemDTO{Teacher: newTeacherDTO(&row.Teacher), Score: row.Score, CourseCount: row.CourseCount}
	}
	return result, nil
}
