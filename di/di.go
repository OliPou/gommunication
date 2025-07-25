package di

import (
	"os"

	"github.com/OliPou/gommunication/email"
	"github.com/OliPou/gommunication/email/providers"
	"github.com/OliPou/gommunication/internal/config"
	"github.com/OliPou/gommunication/internal/database"
	"github.com/OliPou/gommunication/textmessage"
	"github.com/sendgrid/sendgrid-go"
)

type AppDependencies struct {
	APICfg            *email.ApiConfig
	APICfgTextMessage *textmessage.ApiConfig
}

func BuildDependencies() *AppDependencies {
	db := config.DB
	dbQueries := database.New(db)

	emailSender := &providers.SendGridEmailSender{
		Client: sendgrid.NewSendClient(os.Getenv("API_KEY")),
	}

	apiCfg := &email.ApiConfig{
		DB:          dbQueries,
		EmailSender: emailSender,
	}

	apiCfgTextMessage := &textmessage.ApiConfig{
		DB:                   dbQueries,
		ApiKey:               os.Getenv("VONAGE_API_KEY"),
		ApiSecret:            os.Getenv("VONAGE_API_SECRET"),
		VonageApiUrl:         os.Getenv("VONAGE_API_URL"),
		VonageApiCallBackUrl: os.Getenv("VONAGE_API_CALL_BACK_URL"),
	}

	return &AppDependencies{
		APICfg:            apiCfg,
		APICfgTextMessage: apiCfgTextMessage,
	}
}

func (d *AppDependencies) GetEmailConfig() *email.ApiConfig {
	return d.APICfg
}
func (d *AppDependencies) GetTextMessageConfig() *textmessage.ApiConfig {
	return d.APICfgTextMessage
}
