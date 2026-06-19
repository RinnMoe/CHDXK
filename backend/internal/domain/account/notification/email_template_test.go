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
	if strings.Contains(email.Body, VerificationCodeEmailSubject) {
		t.Fatalf("email.Body = %q, should not contain subject %q", email.Body, VerificationCodeEmailSubject)
	}
	for _, want := range []string{"<!doctype html>", "123456", "10 分钟", "请勿向他人泄露"} {
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
	if strings.Contains(email.Body, AccountBannedEmailSubject) {
		t.Fatalf("email.Body = %q, should not contain subject %q", email.Body, AccountBannedEmailSubject)
	}
	for _, want := range []string{"<table", "alice", "发布违规内容", "2026-06-01 12:00"} {
		if !strings.Contains(email.Body, want) {
			t.Fatalf("email.Body = %q, want to contain %q", email.Body, want)
		}
	}
	for _, obsolete := range []string{"cellpadding=", "cellspacing="} {
		if strings.Contains(email.Body, obsolete) {
			t.Fatalf("email.Body = %q, should not contain obsolete attribute %q", email.Body, obsolete)
		}
	}
}

func TestNewAccountBannedEmailEscapesHTML(t *testing.T) {
	email, err := NewAccountBannedEmail("alice@example.edu", AccountBannedEmailData{
		Reason: `<script>alert("x")</script>`,
	})
	if err != nil {
		t.Fatalf("NewAccountBannedEmail: %v", err)
	}
	if strings.Contains(email.Body, `<script>alert("x")</script>`) {
		t.Fatalf("email.Body = %q, should escape raw HTML", email.Body)
	}
	if !strings.Contains(email.Body, `&lt;script&gt;`) {
		t.Fatalf("email.Body = %q, want escaped script tag", email.Body)
	}
}

func TestNewAccountBannedEmailDefaultsToPermanentBan(t *testing.T) {
	email, err := NewAccountBannedEmail("alice@example.edu", AccountBannedEmailData{})
	if err != nil {
		t.Fatalf("NewAccountBannedEmail: %v", err)
	}
	if !strings.Contains(email.Body, "封禁期限") || !strings.Contains(email.Body, "永久") {
		t.Fatalf("email.Body = %q, want permanent ban text", email.Body)
	}
}
