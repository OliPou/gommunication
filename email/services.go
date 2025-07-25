package email

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/OliPou/gommunication/internal/config"
	"github.com/OliPou/gommunication/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UUIDGenerator func() uuid.UUID

// dereferenceString safely dereferences a string pointer and returns the string value and a boolean indicating validity
func dereferenceString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{
			String: *s,
			Valid:  true,
		}
	}
	return sql.NullString{
		String: "",
		Valid:  false,
	}
}

func SendEmail(c *gin.Context, params EmailParams, consumer string, apiCfg *ApiConfig, generateUUID UUIDGenerator) (Email, error) {
	transactionUUID := generateUUID()
	emailSubject := dereferenceString(params.EmailSubject)
	html := dereferenceString(params.Html)

	config.Log.Debug("Sending email", zap.String("transaction_uuid", transactionUUID.String()), zap.String("consumer", consumer), zap.String("subject", emailSubject.String), zap.String("html", html.String))

	// Create email params struct with single instances of repeated fields
	emailParams := database.CreateEmailParams{
		TransactionUuid: transactionUUID,
		Consumer:        consumer,
		UserName:        params.UserName,
		EmailSubject:    emailSubject,
		Html:            html,
		SenderName:      params.SenderName,
		SenderEmail:     params.SenderEmail,
		RecipientsName:  params.RecipientName,
		RecipientsEmail: params.RecipientEmail,
		Status:          "pending",
		CreatedAt:       time.Now(),
	}

	dbEmail, err := apiCfg.DB.CreateEmail(c, emailParams)
	if err != nil {
		config.Log.Error("Error creating email entry in database", zap.Error(err))
		return Email{}, fmt.Errorf("error creating email entry")
	}

	email := DatabaseEmailToEmail(dbEmail)
	// Send the email using the configured email sender
	sendResult, err := apiCfg.EmailSender.Send(email)
	if err != nil {
		config.Log.Error("Error sending email", zap.Error(err))
		return Email{}, fmt.Errorf("error sending email: %v", err)
	}

	config.Log.Info("Email sent successfully",
		zap.Int("status_code", sendResult.StatusCode),
		zap.String("body", sendResult.Body),
		zap.Any("headers", sendResult.Headers),
	)

	return email, nil
}

func GetEmails(c *gin.Context, apiCfg *ApiConfig, consumer string) ([]Email, error) {

	dbEmails, err := apiCfg.DB.GetEmailConsumer(c, consumer)
	if err != nil {
		config.Log.Error("Error getting email entries from database", zap.Error(err))
		return []Email{}, fmt.Errorf("error getting email entry")
	}
	emails := make([]Email, 0, len(dbEmails))
	for _, dbEmail := range dbEmails {
		emails = append(emails, DatabaseEmailToEmail(dbEmail))
	}
	return emails, nil
}
