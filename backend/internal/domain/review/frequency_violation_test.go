package review

import (
	"errors"
	"strings"
	"testing"
	"time"

	"jcourse/internal/domain/course"
	"jcourse/internal/domain/teacher"
)

func TestFrequencyViolationNewSpamSuspensionEmails(t *testing.T) {
	violation := &FrequencyViolation{
		Reason:          errors.New("same course spam"),
		Review:          &Review{UserID: 10, Content: "spam content"},
		Course:          &course.CourseView{ID: 1, Code: "CS101", Name: "Intro CS", MainTeacher: &teacher.TeacherView{Name: "张三"}},
		SuspendDuration: 2 * time.Hour,
	}

	mails, err := violation.NewSpamSuspensionEmails([]string{"admin@example.edu", "owner@example.edu"})
	if err != nil {
		t.Fatalf("NewSpamSuspensionEmails: %v", err)
	}
	if len(mails) != 2 {
		t.Fatalf("mails = %d, want 2", len(mails))
	}
	mail := mails[0]
	if mail.To != "admin@example.edu" {
		t.Fatalf("mail.To = %q, want admin@example.edu", mail.To)
	}
	if mails[1].To != "owner@example.edu" || mails[1].Body != mail.Body {
		t.Fatalf("second mail = %+v, want owner with shared body", mails[1])
	}
	if mail.Subject != spamSuspensionEmailSubject {
		t.Fatalf("mail.Subject = %q, want %q", mail.Subject, spamSuspensionEmailSubject)
	}
	for _, want := range []string{"选课社区用户封禁通知", "10", "课程：1 CS101 Intro CS 张三", "spam content", "2h0m0s", "same course spam"} {
		if !strings.Contains(mail.Body, want) {
			t.Fatalf("mail.Body = %q, want to contain %q", mail.Body, want)
		}
	}
}

func TestFrequencyViolationNewSpamSuspensionEmailsWithoutCourse(t *testing.T) {
	violation := &FrequencyViolation{
		Reason:          errors.New("similar content"),
		Review:          &Review{UserID: 10, CourseID: 1, Content: "spam content"},
		SuspendDuration: time.Hour,
	}

	if _, err := violation.NewSpamSuspensionEmails([]string{"admin@example.edu"}); err != nil {
		t.Fatalf("NewSpamSuspensionEmails without course: %v", err)
	}
}
