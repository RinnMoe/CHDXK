package notification

import (
	"strings"
	"testing"
	"time"
)

func TestNewVerificationCodeEmailRendersEmbeddedTemplate(t *testing.T) {
	email, err := NewVerificationCodeEmail("alice@example.edu", "123456", 10*time.Minute)
	if err != nil {
		t.Fatalf("NewVerificationCodeEmail: %v", err)
	}
	if email.To != "alice@example.edu" {
		t.Fatalf("email.To = %q, want alice@example.edu", email.To)
	}
	if email.Subject != VerificationCodeEmailSubject {
		t.Fatalf("email.Subject = %q, want %q", email.Subject, VerificationCodeEmailSubject)
	}
	for _, want := range []string{"选课社区验证码", "123456", "10 分钟", "请勿向他人泄露"} {
		if !strings.Contains(email.Body, want) {
			t.Fatalf("email.Body = %q, want to contain %q", email.Body, want)
		}
	}
}

func TestNewAccountBannedEmailRendersEmbeddedTemplate(t *testing.T) {
	email, err := NewAccountBannedEmail("alice@example.edu", AccountBannedEmailData{
		Username:    "alice",
		Reason:      "发布违规内容",
		BannedUntil: "2026-06-01 12:00",
	})
	if err != nil {
		t.Fatalf("NewAccountBannedEmail: %v", err)
	}
	if email.Subject != AccountBannedEmailSubject {
		t.Fatalf("email.Subject = %q, want %q", email.Subject, AccountBannedEmailSubject)
	}
	for _, want := range []string{"选课社区账号封禁通知", "alice", "发布违规内容", "2026-06-01 12:00"} {
		if !strings.Contains(email.Body, want) {
			t.Fatalf("email.Body = %q, want to contain %q", email.Body, want)
		}
	}
}

func TestNewAccountBannedEmailDefaultsToPermanentBan(t *testing.T) {
	email, err := NewAccountBannedEmail("alice@example.edu", AccountBannedEmailData{})
	if err != nil {
		t.Fatalf("NewAccountBannedEmail: %v", err)
	}
	if !strings.Contains(email.Body, "封禁期限：永久") {
		t.Fatalf("email.Body = %q, want permanent ban text", email.Body)
	}
}
