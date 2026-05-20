package application

import "jcourse/internal/domain/teacher"

type TeacherDTO struct {
	ID         int    `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Title      string `json:"title,omitempty"`
}

func newTeacherDTO(t *teacher.TeacherForQuery) TeacherDTO {
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
	Pinyin     string `form:"pinyin"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}
