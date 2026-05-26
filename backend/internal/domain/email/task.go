package email

import "encoding/json"

const TaskTypeSendEmail = "email:send"

type SendEmailPayload struct {
	EmailType string `json:"email_type"`
	Email     Email  `json:"email"`
}

type SendEmailTask struct {
	payload SendEmailPayload
}

func NewSendEmailTask(emailType string, mail Email) SendEmailTask {
	return SendEmailTask{payload: SendEmailPayload{
		EmailType: emailType,
		Email:     mail,
	}}
}

func (t SendEmailTask) Type() string {
	return TaskTypeSendEmail
}

func (t SendEmailTask) Payload() []byte {
	b, _ := json.Marshal(t.payload)
	return b
}
