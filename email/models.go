package email

import (
	"time"

	"github.com/OliPou/gommunication/internal/database"
	"github.com/google/uuid"
)

// Email represents the email information
// @Description Email information and status
type Email struct {
	// Unique identifier for the email transaction
	TransactionUuid uuid.UUID `json:"transactionUuid"`
	// Consumer identifier
	Consumer string `json:"consumer"`
	// Name of the user sending the email
	UserName string `json:"userName"`
	// Subject of the email
	EmailSubject string `json:"emailSubject"`
	// Plain text content of the email
	EmailText string `json:"emailText"`
	// HTML content of the email
	Html string `json:"html"`
	// Name of the sender
	SenderName string `json:"senderName"`
	// Email address of the sender
	SenderEmail string `json:"senderEmail"`
	// Name of the recipient
	RecipientName string `json:"recipientName"`
	// Email address of the recipient
	RecipientEmail string `json:"recipientEmail"`
	// Current status of the email
	Status string `json:"status"`
	// Timestamp when the email was created
	CreatedAt time.Time `json:"createdAt"`
	// EnableOpenTracking indicates if open tracking is enabled for the email
	EnableOpenTracking bool `json:"enableOpenTracking"`
	// Opened indicates if the email has been opened
	Opened bool `json:"opened"`
}

func DatabaseEmailToEmail(dbEmail database.Email) Email {
	return Email{
		TransactionUuid:    dbEmail.TransactionUuid,
		Consumer:           dbEmail.Consumer,
		UserName:           dbEmail.UserName,
		EmailSubject:       dbEmail.EmailSubject.String,
		EmailText:          dbEmail.EmailText.String,
		Html:               dbEmail.Html.String,
		SenderName:         dbEmail.SenderName,
		SenderEmail:        dbEmail.SenderEmail,
		RecipientName:      dbEmail.RecipientsName,
		RecipientEmail:     dbEmail.RecipientsEmail,
		Status:             dbEmail.Status,
		CreatedAt:          dbEmail.CreatedAt,
		EnableOpenTracking: dbEmail.EnableOpenTracking,
		Opened:             dbEmail.Opened,
	}
}

// EmailParams represents the parameters for sending an email
// @Description Parameters required to send an email
type EmailParams struct {
	// Name of the user sending the email
	UserName string `json:"userName" binding:"required"`
	// Subject of the email
	EmailSubject *string `json:"emailSubject" binding:"required"`
	// Plain text content of the email
	EmailText *string `json:"emailText" binding:"required"`
	// HTML content of the email
	Html *string `json:"html" binding:"required"`
	// Name of the sender
	SenderName string `json:"senderName" binding:"required"`
	// Email address of the sender
	SenderEmail string `json:"senderEmail" binding:"required"`
	// Name of the recipient
	RecipientName string `json:"recipientName" binding:"required"`
	// Email address of the recipient
	RecipientEmail string `json:"recipientEmail" binding:"required"`
	// EnableOpenTracking indicates if open tracking is enabled for the email
	EnableOpenTracking bool `json:"enableOpenTracking" binding:"required"`
}

// SendGridEvent represents an event received from SendGrid's webhook.
// It contains information about the email event, such as the recipient's email address,
// event type, IP address, content type, event ID, machine open status, message ID,
// timestamp, transaction UUID, and user agent.
type SendGridEvent struct {
	// Email address of the recipient
	Email string `json:"email"`
	// Type of event (e.g., "open", "click")
	Event string `json:"event"`
	// IP address of the recipient
	IP string `json:"ip"`
	// Content type of the email
	SGContentType string `json:"sg_content_type"`
	// Event ID from SendGrid
	SGEventID string `json:"sg_event_id"`
	// Indicates if the email was opened by a machine
	SGMachineOpen bool `json:"sg_machine_open"`
	// Message ID from SendGrid
	SGMessageID string `json:"sg_message_id"`
	// Timestamp of the event
	Timestamp int64 `json:"timestamp"`
	// Transaction UUID associated with the email
	TransactionUUID uuid.UUID `json:"transaction_uuid"`
	// User agent of the recipient's device
	UserAgent string `json:"useragent"`
}

// SubdomainOwnership represents the ownership details of a subdomain,
// including its unique identifier, associated subdomain ID, API key,
// and the timestamp when the ownership record was created.
type SubdomainOwnership struct {
	// ID of the subdomain ownership record
	SubdomainOwnershipUUID uuid.UUID `json:"subdomain_ownership_uuid"`
	SubdomainID            uuid.UUID `json:"subdomain_id"`
	APIKey                 string    `json:"api_key"`
}
