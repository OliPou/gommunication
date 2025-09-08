package email

import "fmt"

type ApiConfig struct {
	DB           DBInterface
	EmailSenders map[AccessLevel]EmailSender
}

type AccessLevel string

const (
	AccessLevelBasic   AccessLevel = "basic"
	AccessLevelPremium AccessLevel = "premium"
)

type SendResult struct {
	StatusCode int
	Body       string
	Headers    map[string][]string
}

// ResolveSender returns the EmailSender associated with the specified AccessLevel.
// If the provided level is empty, it defaults to AccessLevelBasic.
// If an EmailSender is configured for the given access level, it is returned.
// Otherwise, an error is returned indicating that no sender is configured for the specified level.
func (a *ApiConfig) ResolveSender(level AccessLevel) (EmailSender, error) {
	if level == "" {
		level = AccessLevelBasic
	}
	if a.EmailSenders != nil {
		if s, ok := a.EmailSenders[level]; ok && s != nil {
			return s, nil
		}
	}
	return nil, fmt.Errorf("no email sender configured for access level %q", level)
}
