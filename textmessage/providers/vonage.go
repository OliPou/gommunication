package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/OliPou/gommunication/internal/database"
	"github.com/OliPou/gommunication/textmessage"
	"github.com/google/uuid"
)

type Client struct {
	ApiKey               string
	ApiSecret            string
	VonageApiUrl         string
	VonageApiCallBackUrl string
}

type VonageTextMessageSender struct {
	Client *Client
}

func (s *VonageTextMessageSender) SendTextMessage(consumer string, params textmessage.TextMessageParams, smsEncoding string) ([]database.CreateTextMessageParams, error) {

	// Prepare the form data for the API request
	formData := url.Values{
		"api_key":    {s.Client.ApiKey},
		"api_secret": {s.Client.ApiSecret},
		"to":         {params.Recipient},
		"from":       {params.Sender},
		"text":       {params.Text},
		"type":       {smsEncoding},
		"callback":   {s.Client.VonageApiCallBackUrl},
	}

	// Make the POST request to Nexmo
	resp, err := http.PostForm(fmt.Sprintf("%s/sms/json", s.Client.VonageApiUrl), formData)
	if err != nil {
		slog.Error("error sending text message", "error", err)
		return nil, fmt.Errorf("error sending text message: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Log the response body for debugging
	fmt.Printf("Response body: %s\n", string(body))

	// Parse the JSON response
	var vonageResp textmessage.VonageResponse
	if err := json.Unmarshal(body, &vonageResp); err != nil {
		return nil, fmt.Errorf("error parsing response JSON: %v", err)
	}

	fmt.Printf("Api response %s", vonageResp.MessageCount)

	textMessagesParams, err := s.MapResponseToTextMessageParams(consumer, params, vonageResp)
	if err != nil {
		return nil, err
	}
	if len(textMessagesParams) == 0 {
		return nil, fmt.Errorf("no message response received")
	}
	return textMessagesParams, nil
}

func (s *VonageTextMessageSender) MapResponseToTextMessageParams(consumer string, params textmessage.TextMessageParams, response textmessage.VonageResponse) ([]database.CreateTextMessageParams, error) {
	var textMessagesParams []database.CreateTextMessageParams
	for _, message := range response.Messages {
		status := "failed"
		if message.Status == "0" {
			status = "pending"
		}

		// Save to database
		textMessageParams := database.CreateTextMessageParams{
			MessageID: uuid.MustParse(message.MessageID),
			Consumer:  consumer,
			UserName:  params.UserName,
			Sender:    params.Sender,
			Recipient: params.Recipient,
			Status:    status,
			ApiKey:    s.GetApiKey(),
			Price:     message.MessagePrice,
			CreatedAt: time.Now(),
		}

		textMessagesParams = append(textMessagesParams, textMessageParams)
	}
	return textMessagesParams, nil
}

func (s *VonageTextMessageSender) GetApiKey() string {
	return s.Client.ApiKey
}

func (s *VonageTextMessageSender) GetApiSecret() string {
	return s.Client.ApiSecret
}

func (s *VonageTextMessageSender) GetApiUrl() string {
	return s.Client.VonageApiUrl
}

func (s *VonageTextMessageSender) GetApiCallBackUrl() string {
	return s.Client.VonageApiCallBackUrl
}

func NewVonageClient(apiKey, apiSecret, apiUrl, apiCallBackUrl string) *Client {
	return &Client{
		ApiKey:               apiKey,
		ApiSecret:            apiSecret,
		VonageApiUrl:         apiUrl,
		VonageApiCallBackUrl: apiCallBackUrl,
	}
}
