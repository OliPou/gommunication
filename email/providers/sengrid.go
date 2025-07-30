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

	// Plain text fallback
	plainText := email.HtmlToPlainText(e.Html)

	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.Subject = e.EmailSubject

	p := mail.NewPersonalization()
	p.AddTos(to)

	addCustomArgs(p, map[string]string{
		"transaction_uuid": e.TransactionUuid.String(),
	})
	message.AddPersonalizations(p)

	message.AddContent(mail.NewContent("text/plain", plainText))
	message.AddContent(mail.NewContent("text/html", e.Html))

	// message := mail.NewSingleEmail(from, e.EmailSubject, to, email.HtmlToPlainText(e.Html), e.Html)

	if e.EnableOpenTracking {
		enableOpenTracking(message)
	}

	resp, err := s.Client.Send(message)
	if err != nil {
		return email.SendResult{}, err
	}

	if resp.StatusCode >= 400 {
		config.LogClient(nil, "SendGrid returned an error: "+fmt.Sprintf("status %d: %s", resp.StatusCode, resp.Body), zap.ErrorLevel)
		return email.SendResult{}, fmt.Errorf("sendgrid returned status %d: %s", resp.StatusCode, resp.Body)
	}

	return email.SendResult{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
		Headers:    resp.Headers,
	}, nil
}

func enableOpenTracking(message *mail.SGMailV3) {
	trackingSettings := mail.NewTrackingSettings()
	openTracking := mail.NewOpenTrackingSetting()
	openTracking.SetEnable(true)
	trackingSettings.SetOpenTracking(openTracking)
	message.SetTrackingSettings(trackingSettings)
}

func addCustomArgs(p *mail.Personalization, args map[string]string) {
	if p.CustomArgs == nil {
		p.CustomArgs = make(map[string]string)
	}

	for k, v := range args {
		p.CustomArgs[k] = v
	}
}
