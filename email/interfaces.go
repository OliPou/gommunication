package email

import (
	"context"

	"github.com/OliPou/gommunication/internal/database"
)

type DBInterface interface {
	CreateEmail(context.Context, database.CreateEmailParams) (database.Email, error)
	UpdateEmailStatus(context.Context, database.UpdateEmailStatusParams) error
	UpdateEmailOpened(context.Context, database.UpdateEmailOpenedParams) error
	GetEmailConsumer(context.Context, string) ([]database.Email, error)
}

type EmailSender interface {
	Send(email Email) (SendResult, error)
}
