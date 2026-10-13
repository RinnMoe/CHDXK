package repository

import (
	"context"

	"jcourse/internal/domain/course"
)

// FindDiscovery interleaves departments without relying on community ratings.
// EXISTS avoids duplicate courses when a semester has multiple offerings.
func (r *CourseRepository) FindDiscovery(ctx context.Context, semester, date string, limit int) ([]course.CourseView, error) {
	if limit <= 0 {
		return []course.CourseView{}, nil
	}
	var ids []int
	if err := r.db.WithContext(ctx).Raw(`
		WITH candidates AS (
			SELECT c.id, c.department,
				ROW_NUMBER() OVER (PARTITION BY c.department ORDER BY MD5(? || ':' || c.id::text), c.id) AS position
			FROM courses c
			WHERE (? = '' OR EXISTS (
				SELECT 1 FROM offered_courses oc WHERE oc.course_id = c.id AND oc.semester = ?
			))
		)
		SELECT id FROM candidates
		ORDER BY position, MD5(? || ':' || department), department, id
		LIMIT ?`, date, semester, semester, date, limit).Scan(&ids).Error; err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []course.CourseView{}, nil
	}
	rows, _, err := r.FindBy(ctx, course.CourseFilter{CourseIDs: ids})
	if err != nil {
		return nil, err
	}
	byID := make(map[int]course.CourseView, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}
	result := make([]course.CourseView, 0, len(ids))
	for _, id := range ids {
		if row, ok := byID[id]; ok {
			result = append(result, row)
		}
	}
	return result, nil
}
