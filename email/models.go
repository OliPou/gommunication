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
}

func DatabaseEmailToEmail(dbEmail database.Email) Email {
	return Email{
		TransactionUuid: dbEmail.TransactionUuid,
		Consumer:        dbEmail.Consumer,
		UserName:        dbEmail.UserName,
		EmailSubject:    dbEmail.EmailSubject.String,
		EmailText:       dbEmail.EmailText.String,
		Html:            dbEmail.Html.String,
		SenderName:      dbEmail.SenderName,
		SenderEmail:     dbEmail.SenderEmail,
		RecipientName:   dbEmail.RecipientsName,
		RecipientEmail:  dbEmail.RecipientsEmail,
		Status:          dbEmail.Status,
		CreatedAt:       dbEmail.CreatedAt,
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
}
