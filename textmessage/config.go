package textmessage

type ApiConfig struct {
	DB                DBInterface
	TextMessageSender TextMessageSender
}
