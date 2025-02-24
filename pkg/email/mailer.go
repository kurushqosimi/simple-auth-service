package email

import (
	"bytes"
	"embed"
	"gopkg.in/gomail.v2"
	"html/template"
	"time"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *gomail.Dialer
	sender string
}

func NewMailer(host string, port int, username, password, sender string) *Mailer {
	d := gomail.NewDialer(host, port, username, password)
	return &Mailer{dialer: d, sender: sender}
}

func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subjectBuf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subjectBuf, "subject", data)
	if err != nil {
		return err
	}

	plainBodyBuf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBodyBuf, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBodyBuf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBodyBuf, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := gomail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subjectBuf.String())
	msg.SetBody("text/plain", plainBodyBuf.String())
	msg.AddAlternative("text/html", htmlBodyBuf.String())

	for i := 0; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}
