package setting

import "context"

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
