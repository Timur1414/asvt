package smtp

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/Timur1414/asvt/internal/security_checker/domain"
	"github.com/Timur1414/asvt/internal/security_checker/videobuffer"
	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

type EventsNotifier struct {
	videoExtractor videobuffer.VideoExtractor
}

func NewEventsNotifier(extractor videobuffer.VideoExtractor) *EventsNotifier {
	return &EventsNotifier{
		videoExtractor: extractor,
	}
}

func (obj *EventsNotifier) Notify(objects []domain.DetectedObject, startFragmentTime time.Time, endFragmentTime time.Time) {
	log := logger.GetLogger()
	m := mail.NewMsg()
	if err := m.From(os.Getenv("MAIL_FROM")); err != nil {
		log.Fatal("failed to set 'from' email", zap.Error(err))
		return
	}
	if err := m.To(os.Getenv("MAIL_TO")); err != nil {
		log.Fatal("failed to set 'to' email", zap.Error(err))
		return
	}
	m.Subject("Events Notifier")
	plainText := "Оповещение от Дозора. Камера обнаружила подозрительные объекты:\n"
	htmlText := "<h1>Оповещение от Дозора</h1><p>Камера обнаружила подозрительные объекты:</p>"
	for _, object := range objects {
		plainText += object.ClassName + " с вероятностью " + strconv.FormatFloat(float64(object.Confidence), 'f', 3, 64) + " находился в поле зрения более 15 секунд\n"
		htmlText += "<p>" + object.ClassName + " с вероятностью " + strconv.FormatFloat(float64(object.Confidence), 'f', 3, 64) + " находился в поле зрения более 15 секунд</p>"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	videoPath, err := obj.videoExtractor.ExtractClip(ctx, startFragmentTime, endFragmentTime)
	if err != nil {
		log.Error("failed to extract video clip from buffer", zap.Error(err))
	} else {
		defer func() {
			_ = os.Remove(videoPath)
		}()
		m.AttachFile(videoPath, mail.WithFileName("event_record.mp4"))
		log.Info("attached video clip to mail", zap.String("path", videoPath))
	}
	m.SetBodyString(mail.TypeTextPlain, plainText)
	m.AddAlternativeString(mail.TypeTextHTML, htmlText)
	client, err := mail.NewClient("smtp.mail.ru",
		mail.WithPort(587),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(os.Getenv("MAIL_FROM")),
		mail.WithPassword(os.Getenv("MAIL_PASSWORD")),
		mail.WithTLSPolicy(mail.TLSMandatory), // или mail.TLSOpportunistic
	)
	if err != nil {
		log.Error("failed to create SMTP client", zap.Error(err))
		return
	}
	defer client.Close()
	if err = client.DialAndSend(m); err != nil {
		log.Error("failed to send email", zap.Error(err))
		return
	}
	log.Info("email sent successfully")
}
