package email

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
)

type Service struct {
	sender Sender
}

func NewService(sender Sender) *Service {
	return &Service{sender: sender}
}

func (s *Service) SendTemplatedEmail(ctx context.Context, payload SendEmailPayload) error {
	body, err := renderTemplate(payload.Template, payload.Params)
	if err != nil {
		return err
	}
	mail := Email{To: payload.To, Subject: payload.Subject, Body: body}
	return s.sender.SendEmail(ctx, mail)
}

func renderTemplate(tmpl string, params map[string]string) (string, error) {
	parsed, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse email template: %w", err)
	}
	var buf bytes.Buffer
	if err := parsed.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("render email template: %w", err)
	}
	return buf.String(), nil
}
