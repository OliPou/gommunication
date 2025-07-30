package email

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/OliPou/gommunication/internal/config"
	"github.com/OliPou/gommunication/internal/database"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sendgrid/sendgrid-go/helpers/eventwebhook"
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

	description := fmt.Sprintf("Sending email | transaction_uuid: %s | consumer: %s | subject: %s | html: %s", transactionUUID.String(), consumer, emailSubject.String, html.String)
	config.LogClient(nil, description, zap.InfoLevel)

	// Create email params struct with single instances of repeated fields
	emailParams := database.CreateEmailParams{
		TransactionUuid:    transactionUUID,
		Consumer:           consumer,
		UserName:           params.UserName,
		EmailSubject:       emailSubject,
		Html:               html,
		SenderName:         params.SenderName,
		SenderEmail:        params.SenderEmail,
		RecipientsName:     params.RecipientName,
		RecipientsEmail:    params.RecipientEmail,
		Status:             "pending",
		CreatedAt:          time.Now(),
		EnableOpenTracking: params.EnableOpenTracking,
	}

	dbEmail, err := apiCfg.DB.CreateEmail(c, emailParams)
	if err != nil {
		config.LogClient(nil, "Error creating email entry in database: "+err.Error(), zap.ErrorLevel)
		return Email{}, fmt.Errorf("error creating email entry")
	}

	// Convert database email entry to Email struct
	email := DatabaseEmailToEmail(dbEmail)
	// Send the email using the configured email sender
	sendResult, err := apiCfg.EmailSender.Send(email)

	// If sending fails, log the error and update the email status in the database
	if err != nil {
		config.LogClient(nil, "Error sending email: "+err.Error(), zap.ErrorLevel)
		if err := failedSendEmail(c, apiCfg, transactionUUID, &email); err != nil {
			config.LogClient(nil, "Error updating email status to failed in database: "+err.Error(), zap.ErrorLevel)
			return email, fmt.Errorf("error updating email status to failed")
		}
		return email, fmt.Errorf("error sending email: %w", err)
	}

	// Log the successful send result
	config.LogClient(nil, "Email sent successfully: "+fmt.Sprintf("status %d: %s", sendResult.StatusCode, sendResult.Body), zap.InfoLevel)

	// Update the email status to 'sent' in the database
	if err := successSendEmail(c, apiCfg, transactionUUID, &email); err != nil {
		config.LogClient(nil, "Error updating email status to sent in database: "+err.Error(), zap.ErrorLevel)
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
		config.LogClient(nil, "Error getting email entries from database: "+err.Error(), zap.ErrorLevel)
		return []Email{}, fmt.Errorf("error getting email entry")
	}
	emails := make([]Email, 0, len(dbEmails))
	for _, dbEmail := range dbEmails {
		emails = append(emails, DatabaseEmailToEmail(dbEmail))
	}
	return emails, nil
}

func failedSendEmail(c *gin.Context, apiCfg *ApiConfig, transactionUUID uuid.UUID, email *Email) error {
	config.LogClient(nil, "Failed to send email", zap.ErrorLevel)
	status := "failed"
	email.Status = status
	if err := apiCfg.DB.UpdateEmailStatus(c, database.UpdateEmailStatusParams{
		TransactionUuid: transactionUUID,
		Status:          status,
	}); err != nil {
		config.LogClient(nil, "Error updating email status to failed in database: "+err.Error(), zap.ErrorLevel)
		return fmt.Errorf("error updating email status to failed")
	}
	return nil
}

func successSendEmail(c *gin.Context, apiCfg *ApiConfig, transactionUUID uuid.UUID, email *Email) error {
	config.LogClient(nil, "Email sent successfully", zap.InfoLevel)
	status := "sent"
	email.Status = status
	if err := apiCfg.DB.UpdateEmailStatus(c, database.UpdateEmailStatusParams{
		TransactionUuid: transactionUUID,
		Status:          status,
	}); err != nil {
		config.LogClient(nil, "Error updating email status to sent in database: "+err.Error(), zap.ErrorLevel)
		return fmt.Errorf("error updating email status to sent")
	}

	return nil
}

// HtmlToPlainText converts an HTML string to plain text by parsing the HTML and extracting its textual content.
// If an error occurs during parsing, it logs the error and returns an empty string.
//
// Parameters:
//   - html: The HTML string to be converted.
//
// Returns:
//   - A plain text representation of the HTML content, or an empty string if parsing fails.
func HtmlToPlainText(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		config.LogClient(nil, "Error parsing HTML to plain text: "+err.Error(), zap.ErrorLevel)
		return ""
	}
	return strings.TrimSpace(doc.Text())
}

// UpdateEmailOpened updates the 'opened' status of an email in the database for a given transaction UUID.
// It logs an error if the update fails and returns an error indicating the failure.
//
// Parameters:
//   - c: the Gin context for the current HTTP request.
//   - apiCfg: the API configuration containing the database connection.
//   - transactionUUID: the UUID of the email transaction to update.
//   - opened: a boolean indicating whether the email has been opened.
//
// Returns:
//   - error: an error if the update fails, otherwise nil.
func UpdateEmailOpened(c *gin.Context, apiCfg *ApiConfig, transactionUUID uuid.UUID, opened bool) error {
	if err := apiCfg.DB.UpdateEmailOpened(c, database.UpdateEmailOpenedParams{
		TransactionUuid: transactionUUID,
		Opened:          opened,
	}); err != nil {
		config.LogClient(nil, "Error updating email opened status in database: "+err.Error(), zap.ErrorLevel)
		return fmt.Errorf("error updating email opened status")
	}
	return nil
}

// VerifiedSignatureSendGrid verifies the SendGrid webhook signature from the incoming HTTP request.
// It checks for the required SendGrid headers, reads and restores the request body,
// converts the SendGrid public key from base64 to ECDSA format, and verifies the signature.
// Logs relevant information and errors using the configured logger.
// Returns true if the signature is valid, false otherwise.
func VerifiedSignatureSendGrid(c *gin.Context) bool {
	signature := c.GetHeader(eventwebhook.VerificationHTTPHeader)
	timestamp := c.GetHeader(eventwebhook.TimestampHTTPHeader)

	sendgridPublicKeyBase64 := os.Getenv("SENDGRID_VERIFICATION_KEY")

	if signature == "" || timestamp == "" {
		config.LogClient(c, "Missing SendGrid headers", zap.ErrorLevel)
		return false
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		config.LogClient(c, "Failed to read request body", zap.ErrorLevel)
		return false
	}

	// Recharge le body pour BindJSON ensuite
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	pubKey, err := eventwebhook.ConvertPublicKeyBase64ToECDSA(sendgridPublicKeyBase64)
	if err != nil {
		config.LogClient(c, "Invalid public key format: "+err.Error(), zap.ErrorLevel)
		return false
	}

	ok, err := eventwebhook.VerifySignature(pubKey, bodyBytes, signature, timestamp)
	if err != nil {
		config.LogClient(c, "Signature verification failed: "+err.Error(), zap.ErrorLevel)
		return false
	}

	if !ok {
		config.LogClient(c, "SendGrid signature invalid", zap.WarnLevel)
		return false
	}

	config.LogClient(c, "SendGrid signature verified", zap.InfoLevel)
	return true
}
