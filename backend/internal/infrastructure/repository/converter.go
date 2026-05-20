package repository

import (
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/teacher"
)

func newTeacherView(e *TeacherEntity) *teacher.TeacherView {
	return &teacher.TeacherView{
		ID:         e.ID,
		Code:       e.Code,
		Name:       e.Name,
		Department: e.Department,
		Title:      e.Title,
	}
}

func newCourseViewFromEntity(e *CourseEntity) *course.CourseView {
	cv := &course.CourseView{
		ID:            e.ID,
		Code:          e.Code,
		Name:          e.Name,
		Credit:        e.Credit,
		Department:    e.Department,
		MainTeacherID: e.MainTeacherID,
		Categories:    e.Categories,
		Language:      e.Language,
		TargetYears:   e.TargetYears,
		Rating: course.RatingInfo{
			Count: e.RatingCount,
			Avg:   e.RatingAvg,
		},
	}
	if e.MainTeacher != nil {
		cv.MainTeacher = newTeacherView(e.MainTeacher)
	}
	return cv
}

func newCourseDetailViewFromEntity(e *CourseEntity) *course.CourseDetailView {
	dv := &course.CourseDetailView{
		ID:            e.ID,
		Code:          e.Code,
		Name:          e.Name,
		Credit:        e.Credit,
		Department:    e.Department,
		MainTeacherID: e.MainTeacherID,
		Categories:    e.Categories,
		Language:      e.Language,
		TargetYears:   e.TargetYears,
		Rating: course.RatingInfo{
			Count: e.RatingCount,
			Avg:   e.RatingAvg,
		},
		OfferedCourses: make([]*course.OfferedCourseView, 0),
	}
	if e.MainTeacher != nil {
		dv.MainTeacher = newTeacherView(e.MainTeacher)
	}
	return dv
}
