package application

import (
	"context"

	"jcourse/internal/domain/teacher"
)

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

type TeacherQueryService struct {
	teacherQuery teacher.TeacherQuery
}

func NewTeacherQueryService(teacherQuery teacher.TeacherQuery) *TeacherQueryService {
	return &TeacherQueryService{teacherQuery: teacherQuery}
}

func (s *TeacherQueryService) ListTeachers(ctx context.Context, f TeacherListFilter) (*PaginatedResult[TeacherDTO], error) {
	filter := teacher.TeacherFilter{
		Department: f.Department,
		Title:      f.Title,
		Pinyin:     f.Pinyin,
		Page:       f.Page,
		PageSize:   f.PageSize,
	}

	teachers, total, err := s.teacherQuery.FindBy(ctx, filter)
	if err != nil {
		return nil, err
	}

	dtos := make([]TeacherDTO, len(teachers))
	for i, t := range teachers {
		dtos[i] = newTeacherDTO(&t)
	}

	return &PaginatedResult[TeacherDTO]{
		Items:    dtos,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}
