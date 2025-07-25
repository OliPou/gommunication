package di

import (
	"os"

	"github.com/OliPou/gommunication/email"
	"github.com/OliPou/gommunication/internal/config"
	"github.com/OliPou/gommunication/internal/database"
	"github.com/OliPou/gommunication/textmessage"
)

type AppDependencies struct {
	APICfg            *email.ApiConfig
	APICfgTextMessage *textmessage.ApiConfig
}

func BuildDependencies() *AppDependencies {
	db := config.DB
	dbQueries := database.New(db)

	apiCfg := &email.ApiConfig{
		DB:     dbQueries,
		ApiKey: os.Getenv("API_KEY"),
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
