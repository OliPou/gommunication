package textmessage

type ApiConfig struct {
	DB                   DBInterface
	ApiKey               string
	ApiSecret            string
	VonageApiUrl         string
	VonageApiCallBackUrl string
}
