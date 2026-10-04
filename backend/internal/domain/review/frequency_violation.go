package review

import (
	"context"
	_ "embed"
	"strconv"
	"time"

	"jcourse/internal/domain/course"
	"jcourse/internal/domain/email"
	"jcourse/pkg/sitecopy"
)

const (
	SpamSuspensionEmailType = "account_banned"
)

var spamSuspensionEmailSubject = sitecopy.Value(context.Background(), "copy.email.subject")

//go:embed templates/review_frequency_suspension.html
var spamSuspensionEmailTemplate string

type FrequencyViolation struct {
	Reason          error
	Review          *Review
	Course          *course.CourseView
	SuspendDuration time.Duration
	BannedUntil     time.Time
}

func (e *FrequencyViolation) Error() string {
	return e.Reason.Error()
}

func (e *FrequencyViolation) Unwrap() error {
	return e.Reason
}

func (e *FrequencyViolation) NewSpamSuspensionEmails(ctx context.Context, recipients []string, localSuspension bool) ([]email.Email, error) {
	body, err := e.renderSpamSuspensionEmailBody(ctx, localSuspension)
	if err != nil {
		return nil, err
	}
	mails := make([]email.Email, 0, len(recipients))
	for _, to := range recipients {
		mails = append(mails, email.Email{To: to, Subject: sitecopy.Value(ctx, "copy.email.subject"), Body: body})
	}
	return mails, nil
}

func (e *FrequencyViolation) renderSpamSuspensionEmailBody(ctx context.Context, localSuspension bool) (string, error) {
	values := map[string]string{
		"UserID":          strconv.Itoa(e.Review.UserID),
		"CourseID":        strconv.Itoa(e.courseID()),
		"CourseCode":      e.courseCode(),
		"CourseName":      e.courseName(),
		"MainTeacherName": e.mainTeacherName(),
		"ReviewContent":   e.Review.Content,
		"Reason":          sitecopy.Text(ctx, e.Reason.Error()),
	}
	if localSuspension {
		values["BannedUntil"] = e.bannedUntilText()
	}
	for _, key := range []string{"greeting", "notice", "user_id", "course", "course_id", "course_code", "course_name", "teacher", "banned_until", "reason", "content", "footer"} {
		values[key] = sitecopy.Value(ctx, "copy.email."+key)
	}
	values["brand"] = sitecopy.Value(ctx, "copy.user.config.brand.743e9916a4b2")
	body, err := email.RenderTemplate(spamSuspensionEmailTemplate, values)
	if err != nil {
		return "", err
	}
	return body, nil
}

func (e *FrequencyViolation) bannedUntilText() string {
	if e.BannedUntil.IsZero() {
		return ""
	}
	return e.BannedUntil.Format(time.DateTime)
}

func (e *FrequencyViolation) courseID() int {
	if e.Course != nil {
		return e.Course.ID
	}
	return e.Review.CourseID
}

func (e *FrequencyViolation) courseCode() string {
	if e.Course == nil {
		return ""
	}
	return e.Course.Code
}

func (e *FrequencyViolation) courseName() string {
	if e.Course == nil {
		return ""
	}
	return e.Course.Name
}

func (e *FrequencyViolation) mainTeacherName() string {
	if e.Course == nil || e.Course.MainTeacher == nil {
		return ""
	}
	return e.Course.MainTeacher.Name
}
