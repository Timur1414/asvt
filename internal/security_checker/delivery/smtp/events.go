package smtp

import (
	"os"

	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

type EventsNotifier struct {
}

func NewEventsNotifier() *EventsNotifier {
	return &EventsNotifier{}
}

func (obj *EventsNotifier) Notify() {
	log := logger.GetLogger()
	m := mail.NewMsg()
	if err := m.From(os.Getenv("MAIL_FROM")); err != nil {
		log.Fatal("failed to set 'from' email", zap.Error(err))
		return
	}
	if err := m.To(os.Getenv("MAIL_TO")); err != nil {
		log.Fatal("failed to set 'to' email", zap.Error(err))
	}
	m.Subject("Events Notifier")
	m.SetBodyString(mail.TypeTextPlain, "Привет! Ваша почтовая программа не поддерживает HTML.")
	m.AddAlternativeString(mail.TypeTextHTML, "<h1>Привет!</h1><p>Это красивое <b>HTML</b>-письмо.</p>")
	client, err := mail.NewClient("smtp.mail.ru",
		mail.WithPort(587),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(os.Getenv("MAIL_FROM")),
		mail.WithPassword(os.Getenv("MAIL_PASSWORD")),
		mail.WithTLSPolicy(mail.TLSMandatory), // или mail.TLSOpportunistic
	)
	// Mail:
	// smtp.mail.ru
	//
	// Google:
	// smtp.gmail.com
	//
	// Mailpit https://github.com/axllent/mailpit
	// (docker run -d -p 8025:8025 -p 1025:1025 axllent/mailpit)
	// localhost:1025
	if err != nil {
		log.Fatal("failed to create SMTP client", zap.Error(err))
		return
	}
	defer client.Close()
	if err = client.DialAndSend(m); err != nil {
		log.Fatal("failed to send email", zap.Error(err))
		return
	}
	log.Info("email sent")
}
