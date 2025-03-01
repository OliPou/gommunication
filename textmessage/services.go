package textmessage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/OliPou/gommunication/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UUIDGenerator func() uuid.UUID

// GSM 7-bit default alphabet
var gsm7Chars string = "@£$¥èéùìòÇ\nØø\rÅåΔ_ΦΓΛΩΠΨΣΘΞ\x1BÆæßÉ !\"#¤%&'()*+,-./0123456789:;<=>?¡ABCDEFGHIJKLMNOPQRSTUVWXYZÄÖÑÜ§¿abcdefghijklmnopqrstuvwxyzäöñüà"

// Extended characters (require escape character)
var gsm7ExtChars string = "^{}\\[~]|€"

func isGSM7(text string) bool {
	for _, r := range text {
		if !strings.ContainsRune(gsm7Chars, r) && !strings.ContainsRune(gsm7ExtChars, r) {
			return false
		}
	}
	return true
}

// dereferenceString safely dereferences a string pointer and returns the string value and a boolean indicating validity
func dereferenceString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{
			String: *s,
			Valid:  true,
		}
	}
	return sql.NullString{
		String: "",
		Valid:  false,
	}
}

// func SendTextMessage(c *gin.Context, params TextMessageParams, consumer string, apiCfg *ApiConfig, generateUUID UUIDGenerator) ([]TextMessage, error) {
// 	// Generate a unique transaction UUID
// 	fmt.Printf("CallBackUrl: %s", apiCfg.VonageApiCallBackUrl)
// 	var smsEncoding string

// 	if isGSM7(params.Text) {
// 		smsEncoding = "text"
// 		slog.Info("GSM 7-bit SMS encoding detected")
// 	} else {
// 		smsEncoding = "unicode"
// 		slog.Info("Unicode SMS encoding detected")
// 	}

// 	// First, try to send the SMS via Nexmo API
// 	formData := url.Values{
// 		"api_key":    {apiCfg.ApiKey},
// 		"api_secret": {apiCfg.ApiSecret},
// 		"to":         {params.Recipient},
// 		"from":       {params.Sender},
// 		"text":       {params.Text},
// 		"type":       {smsEncoding},
// 		"callback":   {apiCfg.VonageApiCallBackUrl},
// 	}

// 	// Make the POST request to Nexmo
// 	resp, err := http.PostForm(fmt.Sprintf("%s/sms/json", apiCfg.VonageApiUrl), formData)
// 	if err != nil {
// 		slog.Error("error sending text message", "error", err)
// 		return []TextMessage{}, fmt.Errorf("error sending text message: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	// Read the response body
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return []TextMessage{}, fmt.Errorf("error reading response body: %v", err)
// 	}

// 	// Log the response body for debugging
// 	fmt.Printf("Response body: %s\n", string(body))

// 	// Parse the JSON response
// 	var vonageResp VonageResponse
// 	if err := json.Unmarshal(body, &vonageResp); err != nil {
// 		return []TextMessage{}, fmt.Errorf("error parsing response JSON: %v", err)
// 	}

// 	fmt.Printf("Api response %s", vonageResp.MessageCount)

// 	var textMessages []TextMessage
// 	for _, message := range vonageResp.Messages {
// 		status := "failed"
// 		if message.Status == "0" {
// 			status = "pending"
// 		}

// 		// Save to database
// 		textMessageParams := database.CreateTextMessageParams{
// 			MessageID: uuid.MustParse(message.MessageID),
// 			Consumer:  consumer,
// 			UserName:  params.UserName,
// 			Sender:    params.Sender,
// 			Recipient: params.Recipient,
// 			Status:    status,
// 			ApiKey:    apiCfg.ApiKey,
// 			Price:     message.MessagePrice,
// 			CreatedAt: time.Now(),
// 		}

// 		dbTextMessage, err := apiCfg.DB.CreateTextMessage(c, textMessageParams)
// 		if err != nil {
// 			return []TextMessage{}, fmt.Errorf("error creating text message entry: %v", err)
// 		}

// 		textMessages = append(textMessages, DatabaseTextMessageToTextMessage(dbTextMessage))
// 	}

// 	if len(textMessages) == 0 {
// 		return []TextMessage{}, fmt.Errorf("failed to send SMS")
// 	}

// 	return textMessages, nil
// }

func SendTextMessage(c *gin.Context, params TextMessageParams, consumer string, apiCfg *ApiConfig, generateUUID UUIDGenerator) ([]TextMessage, error) {
	// Generate a unique transaction UUID
	fmt.Printf("CallBackUrl: %s", apiCfg.VonageApiCallBackUrl)
	var smsEncoding string

	if isGSM7(params.Text) {
		smsEncoding = "text"
		slog.Info("GSM 7-bit SMS encoding detected")
	} else {
		smsEncoding = "unicode"
		slog.Info("Unicode SMS encoding detected")
	}

	// First, try to send the SMS via Nexmo API
	formData := url.Values{
		"api_key":    {apiCfg.ApiKey},
		"api_secret": {apiCfg.ApiSecret},
		"to":         {params.Recipient},
		"from":       {params.Sender},
		"text":       {params.Text},
		"type":       {smsEncoding},
		"callback":   {apiCfg.VonageApiCallBackUrl},
	}

	// Make the POST request to Nexmo
	resp, err := http.PostForm(fmt.Sprintf("%s/sms/json", apiCfg.VonageApiUrl), formData)
	if err != nil {
		slog.Error("error sending text message", "error", err)
		return []TextMessage{}, fmt.Errorf("error sending text message: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []TextMessage{}, fmt.Errorf("error reading response body: %v", err)
	}

	// Log the response body for debugging
	fmt.Printf("Response body: %s\n", string(body))

	// Parse the JSON response
	var vonageResp VonageResponse
	if err := json.Unmarshal(body, &vonageResp); err != nil {
		return []TextMessage{}, fmt.Errorf("error parsing response JSON: %v", err)
	}

	fmt.Printf("Api response %s", vonageResp.MessageCount)

	var textMessages []TextMessage
	for _, message := range vonageResp.Messages {
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
			ApiKey:    apiCfg.ApiKey,
			Price:     message.MessagePrice,
			CreatedAt: time.Now(),
		}

		dbTextMessage, err := apiCfg.DB.CreateTextMessage(c, textMessageParams)
		if err != nil {
			return []TextMessage{}, fmt.Errorf("error creating text message entry: %v", err)
		}

		adapter := TextMessageAdapter{dbTextMessage: dbTextMessage}
		textMessages = append(textMessages, FromInterface(adapter))
	}

	if len(textMessages) == 0 {
		return []TextMessage{}, fmt.Errorf("failed to send SMS")
	}

	return textMessages, nil
}

// func TextMessageWebHook(c *gin.Context, params TextMessageWebHookResponse, apiCfg *ApiConfig) (TextMessageWebHookResponse, error) {
// 	var dlr TextMessageWebHookResponse
// 	if err := c.ShouldBindQuery(&dlr); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return TextMessageWebHookResponse{}, err
// 	}
// 	_, err := apiCfg.DB.UpdateTextMessageStatus(c, database.UpdateTextMessageStatusParams{
// 		MessageID: uuid.MustParse(dlr.MessageID),
// 		ApiKey:    dlr.ApiKey,
// 		Status:    dlr.Status,
// 		Price:     dlr.Price,
// 		ErrCode: sql.NullString{
// 			String: dlr.ErrCode,
// 			Valid:  true,
// 		},
// 	})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return TextMessageWebHookResponse{}, err
// 	}

// 	return dlr, nil
// }

func TextMessageWebHook(c *gin.Context, params TextMessageWebHookResponse, apiCfg *ApiConfig) (TextMessageWebHookResponse, error) {
	var dlr TextMessageWebHookResponse
	if err := c.ShouldBindQuery(&dlr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return TextMessageWebHookResponse{}, err
	}
	_, err := apiCfg.DB.UpdateTextMessageStatus(c, database.UpdateTextMessageStatusParams{
		MessageID: uuid.MustParse(dlr.MessageID),
		ApiKey:    dlr.ApiKey,
		Status:    dlr.Status,
		Price:     dlr.Price,
		ErrCode: sql.NullString{
			String: dlr.ErrCode,
			Valid:  true,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return TextMessageWebHookResponse{}, err
	}

	return dlr, nil
}

// func GetTextMessages(c *gin.Context, apiCfg *ApiConfig, consumer string) ([]TextMessage, error) {
// 	textMessagesRows, err := apiCfg.DB.GetTextMessages(c, consumer)
// 	if err != nil {
// 		return []TextMessage{}, err
// 	}
// 	textMessages := ConvertGetTextMessagesRowsToTextMessages(textMessagesRows)

// 	return textMessages, nil
// }

func GetTextMessages(c *gin.Context, apiCfg *ApiConfig, consumer string) ([]TextMessage, error) {
	textMessagesRows, err := apiCfg.DB.GetTextMessages(c, consumer)
	if err != nil {
		return []TextMessage{}, err
	}

	var textMessages []TextMessage
	for _, row := range textMessagesRows {
		adapter := GetTextMessagesRowAdapter{dbRow: row}
		textMessages = append(textMessages, FromInterface(adapter))
	}

	return textMessages, nil
}

func GetTextMessage(c *gin.Context, apiCfg *ApiConfig, consumer string) (TextMessage, int, error) {
	messageId := c.Param("messageId")
	// if messageId == "" {
	// 	return TextMessage{}, fmt.Errorf("messageId is required")
	// }
	if _, err := uuid.Parse(messageId); err != nil {
		return TextMessage{}, http.StatusBadRequest, fmt.Errorf("invalid uuid")
	}
	// Assuming you have a function to get the text message from the database
	dbTextMessageRow, err := apiCfg.DB.GetTextMessage(c, database.GetTextMessageParams{
		Consumer:  consumer,
		MessageID: uuid.MustParse(messageId),
	})
	if err != nil {
		return TextMessage{}, 307, err
	}

	// Convert using the adapter
	adapter := GetTextMessageRowAdapter{dbRow: dbTextMessageRow}
	textMessage := FromInterface(adapter)
	return textMessage, 0, nil
}
