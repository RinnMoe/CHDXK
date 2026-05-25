package email

import "context"

type Email struct {
	To      string
	Subject string
	Body    string
}

type Sender interface {
	SendEmail(ctx context.Context, email Email) error
}
