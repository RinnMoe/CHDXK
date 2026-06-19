package email

import (
	"bytes"
	"fmt"
	"html/template"
)

func RenderTemplate(tmpl string, data any) (string, error) {
	parsed, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse email template: %w", err)
	}
	var buf bytes.Buffer
	if err := parsed.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render email template: %w", err)
	}
	return buf.String(), nil
}
