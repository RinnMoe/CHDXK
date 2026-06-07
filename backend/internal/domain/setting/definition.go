package setting

import (
	"context"
	"strconv"
	"strings"
	"time"
)

type ValueType string

const (
	ValueTypeString   ValueType = "string"
	ValueTypeInt      ValueType = "int"
	ValueTypeBool     ValueType = "bool"
	ValueTypeFloat    ValueType = "float"
	ValueTypeDuration ValueType = "duration"
)

type Definition struct {
	Key             string
	Group           string
	Label           string
	Description     string
	Type            ValueType
	DefaultValue    string
	Public          bool
	Secret          bool
	RequiresRestart bool
	Validator       ValueValidator
}

func (d Definition) Validate(ctx context.Context, value string) error {
	value = strings.TrimSpace(value)
	switch d.Type {
	case ValueTypeInt:
		if _, err := strconv.Atoi(value); err != nil {
			return ErrInvalidSystemSettingValue
		}
	case ValueTypeBool:
		if _, err := strconv.ParseBool(value); err != nil {
			return ErrInvalidSystemSettingValue
		}
	case ValueTypeFloat:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return ErrInvalidSystemSettingValue
		}
	case ValueTypeDuration:
		if _, err := time.ParseDuration(value); err != nil {
			return ErrInvalidSystemSettingValue
		}
	}
	if d.Validator != nil {
		return d.Validator.Validate(ctx, value)
	}
	return nil
}

func (d Definition) Normalize(value string) string {
	value = strings.TrimSpace(value)
	if d.Type != ValueTypeBool {
		return value
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return value
	}
	return strconv.FormatBool(parsed)
}
