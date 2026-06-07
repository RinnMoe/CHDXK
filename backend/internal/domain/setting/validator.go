package setting

import (
	"context"
	"strings"

	"jcourse/internal/domain/course"
)

type ValueValidator interface {
	Validate(ctx context.Context, value string) error
}

type ValueValidatorFunc func(ctx context.Context, value string) error

func (f ValueValidatorFunc) Validate(ctx context.Context, value string) error {
	return f(ctx, value)
}

func NewCurrentSemesterValidator(courseRepo course.CourseRepository) ValueValidator {
	return currentSemesterValidator{courseRepo: courseRepo}
}

type currentSemesterValidator struct {
	courseRepo course.CourseRepository
}

func (v currentSemesterValidator) Validate(ctx context.Context, value string) error {
	value = strings.TrimSpace(value)
	if value == "" || v.courseRepo == nil {
		return ErrInvalidCurrentSemester
	}
	allowed, err := v.courseRepo.OfferedSemesterExists(ctx, value)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrInvalidCurrentSemester
	}
	return nil
}
