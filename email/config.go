package email

type ApiConfig struct {
	DB          DBInterface
	EmailSender EmailSender
}

type SendResult struct {
	StatusCode int
	Body       string
	Headers    map[string][]string
}
