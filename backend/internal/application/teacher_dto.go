package application

import "jcourse/internal/domain/teacher"

// Read model: teacher query result
type TeacherDTO struct {
	ID         int    `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Title      string `json:"title,omitempty"`
}

func newTeacherDTO(t *teacher.TeacherView) TeacherDTO {
	return TeacherDTO{
		ID:         t.ID,
		Code:       t.Code,
		Name:       t.Name,
		Department: t.Department,
		Title:      t.Title,
	}
}

type TeacherListFilter struct {
	Department string `form:"department"`
	Title      string `form:"title"`
	Q          string `form:"q"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}
