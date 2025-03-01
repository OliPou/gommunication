package textmessage

import (
	"context"

	"github.com/OliPou/gommunication/internal/database"
)

// type DBInterface interface {
// 	CreateTextMessage(context.Context, database.CreateTextMessageParams) (database.TextMessage, error)
// 	GetTextMessages(context.Context, string) ([]database.GetTextMessagesRow, error)
// 	GetTextMessage(context.Context, string, string) (database.GetTextMessageRow, error)
// 	UpdateTextMessageStatus(context.Context, database.UpdateTextMessageStatusParams) (database.TextMessage, error)
// }

type DBInterface interface {
	CreateTextMessage(context.Context, database.CreateTextMessageParams) (database.TextMessage, error)
	GetTextMessages(context.Context, string) ([]database.GetTextMessagesRow, error)
	GetTextMessage(context.Context, database.GetTextMessageParams) (database.GetTextMessageRow, error)
	UpdateTextMessageStatus(context.Context, database.UpdateTextMessageStatusParams) (database.TextMessage, error)
}
