package repository

import (
	"context"
	"jcourse/internal/domain/teacher"
)

func (r *GormCourseHotRepository) TopTeachers(ctx context.Context, period, periodKey string, limit int) ([]teacher.HotTeacherView, error) {
	type row struct {
		ID                            int
		Code, Name, Department, Title string
		Score                         int64
		CourseCount                   int
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		WITH course_teachers AS (
			SELECT DISTINCT oc.course_id, unnest(oc.teacher_ids) AS teacher_id
			FROM offered_courses oc
			WHERE oc.semester = (SELECT MAX(latest.semester) FROM offered_courses latest WHERE latest.course_id = oc.course_id)
			UNION
			SELECT c.id, c.main_teacher_id FROM courses c
			WHERE c.main_teacher_id > 0 AND NOT EXISTS (
				SELECT 1 FROM offered_courses oc WHERE oc.course_id = c.id AND cardinality(oc.teacher_ids) > 0)
		)
		SELECT t.id, t.code, t.name, t.department, t.title,
			SUM(h.score) AS score, COUNT(DISTINCT h.course_id) AS course_count
		FROM course_hot_scores h
		JOIN course_teachers ct ON ct.course_id = h.course_id
		JOIN teachers t ON t.id = ct.teacher_id
		WHERE h.period = ? AND h.period_key = ? AND h.score > 0
		GROUP BY t.id, t.code, t.name, t.department, t.title
		ORDER BY score DESC, course_count DESC, t.id ASC LIMIT ?`, period, periodKey, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]teacher.HotTeacherView, len(rows))
	for i, item := range rows {
		items[i] = teacher.HotTeacherView{Teacher: teacher.TeacherView{ID: item.ID, Code: item.Code, Name: item.Name, Department: item.Department, Title: item.Title}, Score: item.Score, CourseCount: item.CourseCount}
	}
	return items, nil
}

var _ teacher.HotTeacherQuery = (*GormCourseHotRepository)(nil)
