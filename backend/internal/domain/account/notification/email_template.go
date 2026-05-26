package notification

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
	"time"

	"jcourse/internal/domain/email"
)

//go:embed templates/*.txt
var emailTemplateFS embed.FS

type EmailTemplateName string

const (
	EmailTemplateVerificationCode EmailTemplateName = "verification_code"
	EmailTemplateAccountBanned    EmailTemplateName = "account_banned"

	VerificationCodeEmailSubject = "选课社区验证码"
	AccountBannedEmailSubject    = "选课社区账号封禁通知"
)

type verificationCodeEmailData struct {
	Code      string
	ExpiresIn string
}

type AccountBannedEmailData struct {
	Username    string
	Reason      string
	BannedUntil string
}

func NewVerificationCodeEmail(to string, code string, ttl time.Duration) (email.Email, error) {
	body, err := renderEmailTemplate(EmailTemplateVerificationCode, verificationCodeEmailData{
		Code:      code,
		ExpiresIn: formatEmailDuration(ttl),
	})
	if err != nil {
		return email.Email{}, err
	}
	return email.Email{To: to, Subject: VerificationCodeEmailSubject, Body: body}, nil
}

func NewAccountBannedEmail(to string, data AccountBannedEmailData) (email.Email, error) {
	body, err := renderEmailTemplate(EmailTemplateAccountBanned, data)
	if err != nil {
		return email.Email{}, err
	}
	return email.Email{To: to, Subject: AccountBannedEmailSubject, Body: body}, nil
}

func renderEmailTemplate(name EmailTemplateName, data any) (string, error) {
	tmpl, err := template.ParseFS(emailTemplateFS, "templates/"+string(name)+".txt")
	if err != nil {
		return "", fmt.Errorf("parse email template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render email template %s: %w", name, err)
	}
	return buf.String(), nil
}

func formatEmailDuration(d time.Duration) string {
	if d <= 0 {
		return "一段时间"
	}

	d = d.Round(time.Minute)
	if d < time.Minute {
		return "1 分钟"
	}
	if d%time.Hour == 0 {
		return fmt.Sprintf("%d 小时", int(d/time.Hour))
	}
	return fmt.Sprintf("%d 分钟", int(d/time.Minute))
}
