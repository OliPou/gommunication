package email

import (
	"context"

	"github.com/OliPou/gommunication/internal/database"
	"github.com/google/uuid"
)

type DBInterface interface {
	CreateEmail(context.Context, database.CreateEmailParams) (database.Email, error)
	UpdateEmailStatus(context.Context, database.UpdateEmailStatusParams) error
	UpdateEmailOpened(context.Context, database.UpdateEmailOpenedParams) error
	GetEmailConsumer(context.Context, string) ([]database.Email, error)
	CreateAvailableSubdomain(context.Context, string) (database.AvailableSubdomain, error)
	CreateSubdomainOwnership(context.Context, database.CreateSubdomainOwnershipParams) (database.SubdomainOwnership, error)
	GetAvailableSubdomainByName(context.Context, string) (database.AvailableSubdomain, error)
	GetSubdomainOwnershipBySubdomain(context.Context, uuid.UUID) (database.SubdomainOwnership, error)
	IsSubdomainAllowedForConsumer(context.Context, database.IsSubdomainAllowedForConsumerParams) (bool, error)
}

type EmailSender interface {
	Send(email Email) (SendResult, error)
}
