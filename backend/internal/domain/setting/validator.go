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

type ValueValidatorFactory struct {
	validators map[string]ValueValidator
}

func NewValueValidatorFactory(validators map[string]ValueValidator) *ValueValidatorFactory {
	copy := make(map[string]ValueValidator, len(validators))
	for key, validator := range validators {
		if key != "" && validator != nil {
			copy[key] = validator
		}
	}
	return &ValueValidatorFactory{validators: copy}
}

func (f *ValueValidatorFactory) Validate(ctx context.Context, key, value string) error {
	if f == nil {
		return nil
	}
	validator := f.validators[key]
	if validator == nil {
		return nil
	}
	return validator.Validate(ctx, value)
}

func NewSystemSettingValueValidatorFactory(courseRepo course.CourseRepository) *ValueValidatorFactory {
	return NewValueValidatorFactory(map[string]ValueValidator{
		SystemSettingKeyCurrentSemester: currentSemesterValidator{courseRepo: courseRepo},
	})
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
