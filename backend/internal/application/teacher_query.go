package application

import (
	"context"
	"errors"

	"jcourse/internal/domain/teacher"
)

var ErrTeacherNotFound = errors.New("teacher not found")

type TeacherQueryService struct {
	teacherQuery teacher.TeacherQuery
}

func NewTeacherQueryService(teacherQuery teacher.TeacherQuery) *TeacherQueryService {
	return &TeacherQueryService{teacherQuery: teacherQuery}
}

func (s *TeacherQueryService) GetTeacherFilters(ctx context.Context) (*teacher.TeacherFilters, error) {
	return s.teacherQuery.GetFilters(ctx)
}

func (s *TeacherQueryService) GetTeacher(ctx context.Context, teacherID int) (*TeacherDTO, error) {
	teachers, _, err := s.teacherQuery.FindBy(ctx, teacher.TeacherFilter{TeacherIDs: []int{teacherID}})
	if err != nil {
		return nil, err
	}
	if len(teachers) == 0 {
		return nil, ErrTeacherNotFound
	}

	dto := newTeacherDTO(&teachers[0])
	return &dto, nil
}

func (s *TeacherQueryService) ListTeachers(ctx context.Context, f TeacherListFilter) (*PaginatedResult[TeacherDTO], error) {
	filter := teacher.TeacherFilter{
		Department: f.Department,
		Title:      f.Title,
		Q:          f.Q,
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
