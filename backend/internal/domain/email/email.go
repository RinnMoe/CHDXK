package email

import "context"

type Email struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type Sender interface {
	SendEmail(ctx context.Context, email Email) error
}
