package email

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/OliPou/gommunication/internal/common"
	"github.com/OliPou/gommunication/internal/config"
	"github.com/OliPou/gommunication/internal/database"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sendgrid/sendgrid-go/helpers/eventwebhook"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
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

// AttachmentLimits defines the size limits for attachments
const (
	MaxAttachmentSize  = 10 * 1024 * 1024 // 10MB per file (reduced from 25MB)
	MaxTotalSize       = 20 * 1024 * 1024 // 20MB total (reduced from 30MB)
	MaxAttachmentCount = 5                // Maximum number of attachments
)

// validateAttachmentSize validates attachment size without full decoding
func validateAttachmentSize(base64Content string, filename string) error {
	// Calculate decoded size without actually decoding
	// Base64 encoding adds ~33% overhead, so we can estimate size
	base64Len := len(base64Content)

	// Remove padding characters for accurate calculation
	padding := 0
	if base64Len > 0 && base64Content[base64Len-1] == '=' {
		padding++
		if base64Len > 1 && base64Content[base64Len-2] == '=' {
			padding++
		}
	}

	// Calculate actual file size: (base64_length - padding) * 3 / 4
	estimatedSize := (base64Len - padding) * 3 / 4

	if estimatedSize > MaxAttachmentSize {
		return fmt.Errorf("attachment %s exceeds %dMB limit (estimated size: %dMB)",
			filename, MaxAttachmentSize/(1024*1024), estimatedSize/(1024*1024))
	}

	return nil
}

// validateAttachment validates an email attachment with memory optimization
func validateAttachment(attachment EmailAttachment) error {
	// Validate MIME type first (no memory allocation)
	if attachment.Type == "" {
		return fmt.Errorf("MIME type is required for attachment %s", attachment.Filename)
	}

	// Validate filename
	if attachment.Filename == "" {
		return fmt.Errorf("filename is required for attachment")
	}

	// Check size without full decoding
	if err := validateAttachmentSize(attachment.Content, attachment.Filename); err != nil {
		return err
	}

	// Only decode a small portion to validate base64 format (first 100 characters)
	testContent := attachment.Content
	if len(testContent) > 100 {
		testContent = testContent[:100]
	}

	if _, err := base64.StdEncoding.DecodeString(testContent); err != nil {
		return fmt.Errorf("invalid base64 content for attachment %s", attachment.Filename)
	}

	return nil
}

// validateAttachmentsCollection validates the entire collection efficiently
func validateAttachmentsCollection(attachments []EmailAttachment) error {
	if len(attachments) > MaxAttachmentCount {
		return fmt.Errorf("too many attachments: maximum %d allowed, got %d", MaxAttachmentCount, len(attachments))
	}

	totalEstimatedSize := 0

	for i, att := range attachments {
		// Validate individual attachment
		if err := validateAttachment(att); err != nil {
			return fmt.Errorf("attachment %d: %w", i+1, err)
		}

		// Calculate estimated size for total size check
		base64Len := len(att.Content)
		padding := 0
		if base64Len > 0 && att.Content[base64Len-1] == '=' {
			padding++
			if base64Len > 1 && att.Content[base64Len-2] == '=' {
				padding++
			}
		}
		estimatedSize := (base64Len - padding) * 3 / 4
		totalEstimatedSize += estimatedSize
	}

	if totalEstimatedSize > MaxTotalSize {
		return fmt.Errorf("total attachment size exceeds %dMB limit (estimated: %dMB)",
			MaxTotalSize/(1024*1024), totalEstimatedSize/(1024*1024))
	}

	return nil
}

// convertToSendGridAttachments converts EmailAttachment slice to SendGrid Attachment slice
// with memory-efficient processing
func convertToSendGridAttachments(attachments []EmailAttachment) ([]*mail.Attachment, error) {
	// Pre-allocate slice to avoid memory reallocations
	sgAttachments := make([]*mail.Attachment, 0, len(attachments))

	for i, att := range attachments {
		// Create SendGrid attachment without additional validation
		// (validation already done in validateAttachmentsCollection)
		sgAttachment := mail.NewAttachment()
		sgAttachment.SetContent(att.Content)
		sgAttachment.SetType(att.Type)
		sgAttachment.SetFilename(att.Filename)

		if att.Name != "" && att.Name != att.Filename {
			// Use Name as display name if different from filename
			sgAttachment.SetFilename(att.Name)
		}

		disposition := att.Disposition
		if disposition == "" {
			disposition = "attachment"
		}
		sgAttachment.SetDisposition(disposition)

		if att.ContentID != "" {
			sgAttachment.SetContentID(att.ContentID)
		}

		sgAttachments = append(sgAttachments, sgAttachment)

		// Log memory-friendly message
		config.LogClient(nil, fmt.Sprintf("Processed attachment %d/%d: %s", i+1, len(attachments), att.Filename), zap.DebugLevel)
	}

	return sgAttachments, nil
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
func SendEmail(c *gin.Context, params EmailParams, consumer string, bu string, apiCfg *ApiConfig, generateUUID UUIDGenerator) (Email, error) {
	transactionUUID := generateUUID()
	emailSubject := dereferenceString(params.EmailSubject)
	html := dereferenceString(params.Html)
	replyTo := dereferenceString(params.ReplyTo)

	// Validate attachments if any
	if len(params.Attachments) > 0 {
		// Use optimized validation that doesn't decode all content
		if err := validateAttachmentsCollection(params.Attachments); err != nil {
			config.LogClient(nil, "Attachment validation failed: "+err.Error(), zap.ErrorLevel)
			return Email{}, common.NewRequestError(400, err.Error())
		}
	}

	description := fmt.Sprintf("Sending email | transaction_uuid: %s | consumer: %s | subject: %s | attachments: %d",
		transactionUUID.String(), consumer, emailSubject.String, len(params.Attachments))
	config.LogClient(nil, description, zap.InfoLevel)

	domain := extractDomainFromEmail(params.SenderEmail)

	allowed, err := apiCfg.DB.IsSubdomainAllowedForConsumer(c, database.IsSubdomainAllowedForConsumerParams{Name: domain, BusinessUnit: bu})
	if err != nil {
		config.LogClient(nil, "DB error checking subdomain permission: "+err.Error(), zap.ErrorLevel)
		return Email{}, fmt.Errorf("internal error while checking subdomain permission")
	}
	if !allowed {
		config.LogClient(nil, fmt.Sprintf("Unauthorized subdomain usage: %s for consumer %s with Business Unit: %s", domain, consumer, bu), zap.WarnLevel)
		return Email{}, common.NewForbiddenError(fmt.Sprintf("You are not allowed to use this subdomain: %s", domain))
	}

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
		ReplyTo:            replyTo,
	}

	dbEmail, err := apiCfg.DB.CreateEmail(c, emailParams)
	if err != nil {
		config.LogClient(nil, "Error creating email entry in database: "+err.Error(), zap.ErrorLevel)
		return Email{}, fmt.Errorf("error creating email entry")
	}

	// Convert database email entry to Email struct
	email := DatabaseEmailToEmail(dbEmail)

	// Add attachments to email struct for the sender
	email.Attachments = params.Attachments

	// Send the email using the configured email sender
	sender, err := apiCfg.ResolveSender(params.AccessLevel)
	if err != nil {
		config.LogClient(nil, err.Error(), zap.ErrorLevel)
		if err2 := failedSendEmail(c, apiCfg, transactionUUID, &email); err2 != nil {
			config.LogClient(nil, "Error updating email status to failed in database: "+err2.Error(), zap.ErrorLevel)
		}
		return email, err
	}
	sendResult, err := sender.Send(email)

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
func GetEmails(c *gin.Context, apiCfg *ApiConfig, consumer string, bu string) ([]Email, error) {
	dbEmails, err := apiCfg.DB.GetEmailsByConsumerAndBusinessUnit(c, database.GetEmailsByConsumerAndBusinessUnitParams{
		Consumer:     consumer,
		BusinessUnit: bu,
	})
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

// extractDomainFromEmail extracts the domain part from an email address.
// It splits the email address at the '@' character and returns the domain part.
// If the email address is invalid (does not contain '@'), it returns an empty string.
func extractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
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

// CreateSubdomainOwnership creates a new subdomain ownership record in the database.
// It takes a Gin context, the parameters for subdomain ownership, the consumer identifier,
// and the API configuration. The function logs the creation attempt and any errors encountered.
// On success, it returns the created SubdomainOwnership and a nil error; otherwise, it returns
// an empty SubdomainOwnership and an error.
func CreateSubdomainOwnership(c *gin.Context, params SubdomainOwnership, consumer string, bu string, apiCfg *ApiConfig) (SubdomainOwnership, error) {

	if params.BusinessUnit != bu {
		errMsg := fmt.Sprintf("BusinessUnit params %s does not match with Business Unit path %s", params.BusinessUnit, bu)
		config.LogClient(c, errMsg, zap.ErrorLevel)
		return SubdomainOwnership{}, fmt.Errorf("%s", errMsg)
	}

	subdomainOwnership := database.CreateSubdomainOwnershipParams{
		SubdomainOwnershipUuid: uuid.New(),
		SubdomainID:            params.SubdomainID,
		BusinessUnit:           params.BusinessUnit,
	}
	config.LogClient(c, fmt.Sprintf("Consumer %s creating subdomain ownership for subdomain ID: %s with Business Unit: %s", consumer, subdomainOwnership.SubdomainID, subdomainOwnership.BusinessUnit), zap.InfoLevel)

	sub, err := apiCfg.DB.CreateSubdomainOwnership(c, subdomainOwnership)
	if err != nil {
		config.LogClient(c, "Error creating subdomain ownership: "+err.Error(), zap.ErrorLevel)
		return SubdomainOwnership{}, fmt.Errorf("error creating subdomain ownership")
	}
	result := SubdomainOwnership{
		SubdomainOwnershipUUID: sub.SubdomainOwnershipUuid,
		SubdomainID:            sub.SubdomainID,
		BusinessUnit:           sub.BusinessUnit,
	}
	return result, nil
}
