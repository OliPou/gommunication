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

// Send sends an email using the SendGrid API client.
// It constructs the email message from the provided email.Email struct,
// sends it via the SendGrid client, and returns the result or an error.
// If SendGrid returns an error status code (>= 400), the error is logged
// and an error is returned.
//
// Parameters:
//   - e: The email.Email struct containing sender, recipient, subject, and content.
//
// Returns:
//   - email.SendResult: The result of the send operation, including status code, body, and headers.
//   - error: An error if sending fails or SendGrid returns an error status code.
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
