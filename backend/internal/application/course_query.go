package application

import "jcourse/internal/domain/course"

type TeacherDTO struct {
	ID         int    `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Department string `json:"department"`
}

func newTeacherDTO(t *course.TeacherForQuery) TeacherDTO {
	return TeacherDTO{
		ID:         t.ID,
		Code:       t.Code,
		Name:       t.Name,
		Department: t.Department,
	}
}

type CourseDTO struct {
	ID          int        `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	MainTeacher TeacherDTO `json:"main_teacher"`
}

func newCourseDTO(c *course.CourseForQuery) CourseDTO {
	dto := CourseDTO{
		ID:          c.ID,
		Code:        c.Code,
		Name:        c.Name,
		MainTeacher: TeacherDTO{ID: c.MainTeacherID},
	}
	if c.MainTeacher != nil {
		dto.MainTeacher = newTeacherDTO(c.MainTeacher)
	}
	return dto
}
