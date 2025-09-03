package textmessage

import (
	"time"

	"github.com/OliPou/gommunication/internal/database"
	"github.com/google/uuid"
)

type TextMessageParams struct {
	UserName  string `json:"userName" binding:"required"`
	Sender    string `json:"sender"    binding:"required,sendername"`
	Recipient string `json:"recipient" binding:"required,e164"`
	Text      string `json:"text" binding:"required"`
}

type TextMessageWebHookResponse struct {
	ApiKey           string `form:"api-key"`
	ClientRef        string `form:"client-ref"`
	ErrCode          string `form:"err-code"`
	MessageTimestamp string `form:"message-timestamp"`
	MessageID        string `form:"messageId"`
	Msisdn           string `form:"msisdn"`
	NetworkCode      string `form:"network-code"`
	Price            string `form:"price"`
	Scts             string `form:"scts"`
	Status           string `form:"status"`
	Timestamp        string `form:"timestamp"`
	To               string `form:"to"`
}

// VonageResponse represents the structure of the Vonage API response
type VonageResponse struct {
	Messages []struct {
		Status       string `json:"status"`
		ErrorText    string `json:"error-text,omitempty"`
		MessageID    string `json:"message-id"`
		To           string `json:"to"`
		MessagePrice string `json:"message-price"`
	} `json:"messages"`
	MessageCount string `json:"message-count"`
}

type TextMessageInterface interface {
	GetMessageID() uuid.UUID
	GetConsumer() string
	GetUserName() string
	GetSender() string
	GetRecipient() string
	GetStatus() string
	GetCreatedAt() time.Time
}
type TextMessageAdapter struct {
	dbTextMessage database.TextMessage
}

func (a TextMessageAdapter) GetMessageID() uuid.UUID {
	return a.dbTextMessage.MessageID
}

func (a TextMessageAdapter) GetConsumer() string {
	return a.dbTextMessage.Consumer
}

func (a TextMessageAdapter) GetUserName() string {
	return a.dbTextMessage.UserName
}

func (a TextMessageAdapter) GetSender() string {
	return a.dbTextMessage.Sender
}

func (a TextMessageAdapter) GetRecipient() string {
	return a.dbTextMessage.Recipient
}

func (a TextMessageAdapter) GetStatus() string {
	return a.dbTextMessage.Status
}

func (a TextMessageAdapter) GetCreatedAt() time.Time {
	return a.dbTextMessage.CreatedAt
}

type GetTextMessageRowAdapter struct {
	dbRow database.GetTextMessageRow
}

func (a GetTextMessageRowAdapter) GetMessageID() uuid.UUID {
	return a.dbRow.MessageID
}

func (a GetTextMessageRowAdapter) GetConsumer() string {
	return a.dbRow.Consumer
}

func (a GetTextMessageRowAdapter) GetUserName() string {
	return a.dbRow.UserName
}

func (a GetTextMessageRowAdapter) GetSender() string {
	return a.dbRow.Sender
}

func (a GetTextMessageRowAdapter) GetRecipient() string {
	return a.dbRow.Recipient
}

func (a GetTextMessageRowAdapter) GetStatus() string {
	return a.dbRow.Status
}

func (a GetTextMessageRowAdapter) GetCreatedAt() time.Time {
	return a.dbRow.CreatedAt
}

type GetTextMessagesRowAdapter struct {
	dbRow database.GetTextMessagesRow
}

func (a GetTextMessagesRowAdapter) GetMessageID() uuid.UUID {
	return a.dbRow.MessageID
}

func (a GetTextMessagesRowAdapter) GetConsumer() string {
	return a.dbRow.Consumer
}

func (a GetTextMessagesRowAdapter) GetUserName() string {
	return a.dbRow.UserName
}

func (a GetTextMessagesRowAdapter) GetSender() string {
	return a.dbRow.Sender
}

func (a GetTextMessagesRowAdapter) GetRecipient() string {
	return a.dbRow.Recipient
}

func (a GetTextMessagesRowAdapter) GetStatus() string {
	return a.dbRow.Status
}

func (a GetTextMessagesRowAdapter) GetCreatedAt() time.Time {
	return a.dbRow.CreatedAt
}

type TextMessage struct {
	MessageID uuid.UUID
	Consumer  string
	UserName  string
	Sender    string
	Recipient string
	Status    string
	CreatedAt time.Time
}

func FromInterface(tm TextMessageInterface) TextMessage {
	return TextMessage{
		MessageID: tm.GetMessageID(),
		Consumer:  tm.GetConsumer(),
		UserName:  tm.GetUserName(),
		Sender:    tm.GetSender(),
		Recipient: tm.GetRecipient(),
		Status:    tm.GetStatus(),
		CreatedAt: tm.GetCreatedAt(),
	}
}
