package email

import (
	"time"

	"github.com/OliPou/gommunication/internal/database"
	"github.com/google/uuid"
)

type Email struct {
	TransactionUuid uuid.UUID
	Consumer        string
	UserName        string
	EmailSubject    string
	EmailText       string
	Html            string
	SenderName      string
	SenderEmail     string
	RecipientName   string
	RecipientEmail  string
	Status          string
	CreatedAt       time.Time
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

type EmailParams struct {
	UserName       string  `json:"userName" binding:"required"`
	EmailSubject   *string `json:"emailSubject,required"`
	EmailText      *string `json:"emailText,required"`
	Html           *string `json:"html,required"`
	SenderName     string  `json:"senderName" binding:"required"`
	SenderEmail    string  `json:"senderEmail" binding:"required"`
	RecipientName  string  `json:"recipientName" binding:"required"`
	RecipientEmail string  `json:"recipientEmail" binding:"required"`
}
