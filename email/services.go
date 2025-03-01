package email

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/OliPou/gommunication/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
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
		fmt.Printf("Error creating entry in db: %v", err)
		return Email{}, fmt.Errorf("error creating email entry")
	}

	// Send email using SendGrid
	from := mail.NewEmail(params.SenderName, params.SenderEmail)
	subject := emailSubject.String
	to := mail.NewEmail(params.RecipientName, params.RecipientEmail)
	plainTextContent := "This is a plain text version of the email."
	htmlContent := html.String
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(apiCfg.ApiKey)
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
		return Email{}, fmt.Errorf("error sending email: %v", err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}

	return DatabaseEmailToEmail(dbEmail), nil
}

func GetEmails(c *gin.Context, apiCfg *ApiConfig, consumer string) ([]Email, error) {

	dbEmails, err := apiCfg.DB.GetEmailConsumer(c, consumer)
	if err != nil {
		fmt.Printf("Error getting db entry %v", err)
		return []Email{}, fmt.Errorf("error getting email entry")
	}
	emails := make([]Email, 0, len(dbEmails))
	for _, dbEmail := range dbEmails {
		emails = append(emails, DatabaseEmailToEmail(dbEmail))
	}
	return emails, nil
}
