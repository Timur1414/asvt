package smtp

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/Timur1414/asvt/internal/security_checker/domain"
	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

type EventsNotifier struct {
}

func NewEventsNotifier() *EventsNotifier {
	return &EventsNotifier{}
}

func (obj *EventsNotifier) Notify(objects []domain.DetectedObject) {
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
	plainText := "Оповещение от Дозора. Камера обнаружила подозрительные объекты:\n"
	htmlText := "<h1>Оповещение от Дозора</h1><p>Камера обнаружила подозрительные объекты:</p>"
	for _, object := range objects {
		plainText += object.ClassName + " с вероятностью " + strconv.FormatFloat(float64(object.Confidence), 'f', 3, 64) + " находится в поле зрения более 15 секунд\n"
		htmlText += "<p>" + object.ClassName + " с вероятностью " + strconv.FormatFloat(float64(object.Confidence), 'f', 3, 64) + " находится в поле зрения более 15 секунд</p>"
	}
	rtspURL := os.Getenv("CAMERA_STREAM_URL")
	frameBytes, err := captureFrame(rtspURL)
	if err != nil {
		log.Error("failed to capture frame", zap.Error(err))
	} else {
		if err := m.EmbedReader("snapshot.jpg", bytes.NewReader(frameBytes)); err != nil {
			log.Error("failed to embed snapshot into email", zap.Error(err))
		} else {
			htmlText += `<p><img src="cid:snapshot.jpg" alt="Кадр с камеры" style="max-width: 100%; height: auto;"/></p>`
		}
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

func captureFrame(rtspURL string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var stdout, stderr bytes.Buffer
	// ffmpeg аргументы:
	// -rtsp_transport tcp : стабильная передача через TCP
	// -i <url>            : адрес RTSP-потока в MediaMTX
	// -vframes 1          : захватить ровно 1 кадр
	// -f image2pipe       : передача изображения в пайп
	// -vcodec mjpeg       : кодирование в JPEG
	// pipe:1              : вывод в stdout
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-rtsp_transport", "tcp",
		"-i", rtspURL,
		"-vframes", "1",
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"pipe:1",
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg capture error: %w, stderr: %s", err, stderr.String())
	}
	return stdout.Bytes(), nil
}
