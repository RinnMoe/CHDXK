package email

import "encoding/json"

const TaskTypeSendEmail = "email:send"

type SendEmailPayload struct {
	EmailType string            `json:"email_type"`
	To        string            `json:"to"`
	Subject   string            `json:"subject"`
	Template  string            `json:"template"`
	Params    map[string]string `json:"params"`
}

type SendEmailTask struct {
	payload SendEmailPayload
}

func NewSendEmailTask(emailType string, to string, subject string, template string, params map[string]string) SendEmailTask {
	return SendEmailTask{payload: SendEmailPayload{
		EmailType: emailType,
		To:        to,
		Subject:   subject,
		Template:  template,
		Params:    params,
	}}
}

func (t SendEmailTask) Type() string {
	return TaskTypeSendEmail
}

func (t SendEmailTask) Payload() []byte {
	b, _ := json.Marshal(t.payload)
	return b
}
