package application

import (
	"context"

	"jcourse/internal/domain/teacher"
)

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

	views := make([]TeacherDTO, len(teachers))
	for i, t := range teachers {
		views[i] = newTeacherDTO(&t)
	}

	return &PaginatedResult[TeacherDTO]{
		Items:    views,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}
