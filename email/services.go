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

// SendEmail handles the process of sending an email within the application.
// It generates a unique transaction UUID, logs the email sending attempt, creates a database entry for the email,
// and sends the email using the configured email sender. The function returns the sent Email struct or an error
// if any step fails.
//
// Parameters:
//   - c: the Gin context for the current HTTP request.
//   - params: the parameters required to compose the email (subject, HTML content, sender/recipient info, etc.).
//   - consumer: a string identifying the consumer of the email service.
//   - apiCfg: the API configuration containing database and email sender instances.
//   - generateUUID: a function to generate a new UUID for the transaction.
//
// Returns:
//   - Email: the sent email as an Email struct.
//   - error: an error if the email could not be created or sent.
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

	// Convert database email entry to Email struct
	email := DatabaseEmailToEmail(dbEmail)

	// Send the email using the configured email sender
	sendResult, err := apiCfg.EmailSender.Send(email)

	// If sending fails, log the error and update the email status in the database
	if err != nil {
		config.Log.Error("Error sending email", zap.Error(err), zap.String("transaction_uuid", transactionUUID.String()))
		if err := failedSendEmail(c, apiCfg, transactionUUID, &email); err != nil {
			config.Log.Error("Error updating email status to failed in database", zap.Error(err))
			return email, fmt.Errorf("error updating email status to failed")
		}
		return email, fmt.Errorf("error sending email: %w", err)
	}

	config.Log.Info("Email sent successfully",
		zap.Int("status_code", sendResult.StatusCode),
		zap.String("body", sendResult.Body),
		zap.Any("headers", sendResult.Headers),
	)

	// Update the email status to 'sent' in the database
	if err := successSendEmail(c, apiCfg, transactionUUID, &email); err != nil {
		config.Log.Error("Error updating email status to sent in database", zap.Error(err))
		return email, fmt.Errorf("error updating email status to sent")
	}

	return email, nil
}

// GetEmails retrieves a list of emails associated with the specified consumer from the database.
// It takes a Gin context, an API configuration, and the consumer identifier as parameters.
// Returns a slice of Email objects and an error if the retrieval fails.
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

func failedSendEmail(c *gin.Context, apiCfg *ApiConfig, transactionUUID uuid.UUID, email *Email) error {
	config.Log.Error("Failed to send email", zap.String("transaction_uuid", transactionUUID.String()))
	status := "failed"
	email.Status = status
	if err := apiCfg.DB.UpdateEmailStatus(c, database.UpdateEmailStatusParams{
		TransactionUuid: transactionUUID,
		Status:          status,
	}); err != nil {
		config.Log.Error("Error updating email status to failed in database", zap.Error(err))
		return fmt.Errorf("error updating email status to failed")
	}
	return nil
}

func successSendEmail(c *gin.Context, apiCfg *ApiConfig, transactionUUID uuid.UUID, email *Email) error {
	config.Log.Info("Email sent successfully", zap.String("transaction_uuid", transactionUUID.String()))
	status := "sent"
	email.Status = status
	if err := apiCfg.DB.UpdateEmailStatus(c, database.UpdateEmailStatusParams{
		TransactionUuid: transactionUUID,
		Status:          status,
	}); err != nil {
		config.Log.Error("Error updating email status to sent in database", zap.Error(err))
		return fmt.Errorf("error updating email status to sent")
	}

	return nil
}
