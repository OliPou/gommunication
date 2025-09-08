package di

import (
	"os"

	"github.com/OliPou/gommunication/email"
	"github.com/OliPou/gommunication/email/providers"
	"github.com/OliPou/gommunication/internal/config"
	"github.com/OliPou/gommunication/internal/database"
	"github.com/OliPou/gommunication/textmessage"
	textMsgProviders "github.com/OliPou/gommunication/textmessage/providers"
	"github.com/sendgrid/sendgrid-go"
)

type AppDependencies struct {
	APICfg            *email.ApiConfig
	APICfgTextMessage *textmessage.ApiConfig
}

// BuildDependencies initializes and returns an instance of AppDependencies,
// setting up the required database queries, email sender, and API configurations
// for both email and text message services. It retrieves necessary configuration
// values from environment variables and application config, and wires up the
// dependencies for use throughout the application.
func BuildDependencies() *AppDependencies {
	db := config.DB
	dbQueries := database.New(db)

	sendGridSender := &providers.SendGridEmailSender{
		Client: sendgrid.NewSendClient(os.Getenv("API_KEY")),
	}

	apiCfg := &email.ApiConfig{
		DB: dbQueries,
		EmailSenders: map[email.AccessLevel]email.EmailSender{
			email.AccessLevelPremium: sendGridSender,
		},
	}

	textMsgSender := &textMsgProviders.VonageTextMessageSender{
		Client: textMsgProviders.NewVonageClient(
			os.Getenv("VONAGE_API_KEY"),
			os.Getenv("VONAGE_API_SECRET"),
			os.Getenv("VONAGE_API_URL"),
			os.Getenv("VONAGE_API_CALL_BACK_URL"),
		),
	}

	apiCfgTextMessage := &textmessage.ApiConfig{
		DB:                dbQueries,
		TextMessageSender: textMsgSender,
	}

	return &AppDependencies{
		APICfg:            apiCfg,
		APICfgTextMessage: apiCfgTextMessage,
	}
}

// GetEmailConfig returns the email API configuration associated with the application dependencies.
// It provides access to the email.ApiConfig instance used for email-related operations.
func (d *AppDependencies) GetEmailConfig() *email.ApiConfig {
	return d.APICfg
}

// GetTextMessageConfig returns the API configuration for text message services.
// It provides access to the textmessage.ApiConfig instance managed by AppDependencies.
func (d *AppDependencies) GetTextMessageConfig() *textmessage.ApiConfig {
	return d.APICfgTextMessage
}
