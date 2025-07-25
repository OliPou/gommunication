package providers

import (
	"fmt"

	"github.com/OliPou/gommunication/email"
	"github.com/OliPou/gommunication/internal/config"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

type SendGridEmailSender struct {
	Client *sendgrid.Client
}

func (s *SendGridEmailSender) Send(e email.Email) (email.SendResult, error) {
	from := mail.NewEmail(e.SenderName, e.SenderEmail)
	to := mail.NewEmail(e.RecipientName, e.RecipientEmail)
	message := mail.NewSingleEmail(from, e.EmailSubject, to, "Plain text fallback", e.Html)

	resp, err := s.Client.Send(message)
	if err != nil {
		return email.SendResult{}, err
	}

	if resp.StatusCode >= 400 {
		config.Log.Error("SendGrid returned an error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", resp.Body),
			zap.Any("headers", resp.Headers),
		)
		return email.SendResult{}, fmt.Errorf("sendgrid returned status %d: %s", resp.StatusCode, resp.Body)
	}

	return email.SendResult{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
		Headers:    resp.Headers,
	}, nil
}
