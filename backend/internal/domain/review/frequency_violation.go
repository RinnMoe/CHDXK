package review

import (
	"time"

	"jcourse/internal/domain/course"
)

type FrequencyViolation struct {
	Reason          error
	Review          *Review
	Course          *course.Course
	SuspendDuration time.Duration
}

func (e *FrequencyViolation) Error() string {
	return e.Reason.Error()
}

func (e *FrequencyViolation) Unwrap() error {
	return e.Reason
}
