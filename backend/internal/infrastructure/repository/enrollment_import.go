package repository

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"jcourse/internal/domain/course"
	"jcourse/pkg/apperr"
	"strings"
	"time"
)

func (r *CourseEnrollmentRepository) ImportSemester(ctx context.Context, userID int, semester string, items []course.EnrollmentSourceCourse) (*course.EnrollmentImportResult, error) {
	if userID <= 0 {
		return nil, apperr.BadRequest("请先登录后导入课程")
	}
	result := &course.EnrollmentImportResult{Unmatched: []course.EnrollmentSourceCourse{}}
	courseIDs := []int{}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seen := make(map[string]bool)
		matched := make(map[int]bool)
		for _, item := range items {
			item.LessonNo = strings.TrimSpace(item.LessonNo)
			item.SourceLessonID = strings.TrimSpace(item.SourceLessonID)
			key := "lesson:" + item.LessonNo
			if item.LessonNo == "" {
				key = "id:" + item.SourceLessonID
			}
			if seen[key] {
				continue
			}
			seen[key] = true
			query := tx.Table("course_sections AS section").
				Joins("JOIN offered_courses AS offering ON offering.id = section.offered_course_id").
				Where("offering.semester = ?", semester)
			if item.LessonNo != "" {
				query = query.Where("section.lesson_no = ?", item.LessonNo)
			} else {
				query = query.Where("section.source_lesson_id = ? AND section.course_name = ?", item.SourceLessonID, item.Name)
			}
			var ids []int
			if err := query.Distinct("offering.course_id").Limit(2).Pluck("offering.course_id", &ids).Error; err != nil {
				return err
			}
			// Ambiguous sections must be reported, never assigned to an arbitrary teacher.
			if len(ids) != 1 {
				result.Unmatched = append(result.Unmatched, item)
				continue
			}
			id := ids[0]
			if matched[id] {
				continue
			}
			matched[id] = true
			result.Matched++
			entity := CourseEnrollmentEntity{UserID: userID, CourseID: id, Semester: semester, CreatedAt: time.Now()}
			created := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "course_id"}, {Name: "semester"}}, DoNothing: true}).Create(&entity)
			if created.Error != nil {
				return created.Error
			}
			result.Added += int(created.RowsAffected)
			courseIDs = append(courseIDs, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	r.deleteEnrollmentCache(ctx, userID, courseIDs...)
	result.UnmatchedCount = len(result.Unmatched)
	return result, nil
}

var _ course.EnrollmentImporter = (*CourseEnrollmentRepository)(nil)
