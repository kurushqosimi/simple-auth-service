package adapters

import "fullstack-simple-app/pkg/email"

type EmailAdapter struct {
	mailer *email.Mailer
}

func NewEmailAdapter(mailer *email.Mailer) *EmailAdapter {
	return &EmailAdapter{
		mailer: mailer,
	}
}

func (a *EmailAdapter) SendMail(recipient string, templateFile string, data interface{}) error {
	return a.mailer.Send(recipient, templateFile, data)
}
