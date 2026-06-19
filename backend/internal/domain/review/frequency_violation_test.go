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
		BannedUntil:     time.Date(2026, 6, 19, 12, 20, 30, 0, time.Local),
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
	if strings.Contains(mail.Body, spamSuspensionEmailSubject) {
		t.Fatalf("mail.Body = %q, should not contain subject %q", mail.Body, spamSuspensionEmailSubject)
	}
	for _, want := range []string{"<!doctype html>", "用户 ID", "10", "课程 ID：1", "课程代码：CS101", "课程名称：Intro CS", "主讲教师：张三", "封禁到", "2026-06-19 12:20:30", "spam content", "same course spam"} {
		if !strings.Contains(mail.Body, want) {
			t.Fatalf("mail.Body = %q, want to contain %q", mail.Body, want)
		}
	}
	if strings.Contains(mail.Body, "封禁时长") || strings.Contains(mail.Body, "2h0m0s") {
		t.Fatalf("mail.Body = %q, should contain banned-until time instead of duration", mail.Body)
	}
	for _, obsolete := range []string{"cellpadding=", "cellspacing="} {
		if strings.Contains(mail.Body, obsolete) {
			t.Fatalf("mail.Body = %q, should not contain obsolete attribute %q", mail.Body, obsolete)
		}
	}
}

func TestFrequencyViolationEmailEscapesHTML(t *testing.T) {
	violation := &FrequencyViolation{
		Reason:          errors.New(`<script>alert("reason")</script>`),
		Review:          &Review{UserID: 10, Content: `<script>alert("review")</script>`},
		SuspendDuration: time.Hour,
	}

	mails, err := violation.NewSpamSuspensionEmails([]string{"admin@example.edu"})
	if err != nil {
		t.Fatalf("NewSpamSuspensionEmails: %v", err)
	}
	if strings.Contains(mails[0].Body, `<script>alert("review")</script>`) || strings.Contains(mails[0].Body, `<script>alert("reason")</script>`) {
		t.Fatalf("mail.Body = %q, should escape raw HTML", mails[0].Body)
	}
	if !strings.Contains(mails[0].Body, `&lt;script&gt;`) {
		t.Fatalf("mail.Body = %q, want escaped script tag", mails[0].Body)
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
