package textmessage

import (
	"fmt"
	"net/http"

	"github.com/OliPou/gommunication/internal/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HandlerSendTextMessage sends a text message
// @Summary Send a text message
// @Description Send a text message to a recipient
// @Tags TextMessage
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body TextMessageParams true "Text message parameters"
// @Success 200 {object} TextMessage
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /text-message/send [post]
func (apiCfg *ApiConfig) HandlerSendTextMessage(c *gin.Context, consumer string) {
	var params TextMessageParams
	if err := common.ValidateRequest(c, &params); err != nil {
		common.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Error validating body: %v", err))
		return
	}
	sendTextMessage, err := SendTextMessage(c, params, consumer, apiCfg, uuid.New)
	if err != nil {
		if err.Error() == "invalid uuid" {
			common.RespondError(c, http.StatusBadRequest, fmt.Sprintf("Error: %v", err))
			return
		}
		common.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Error: %v", err))
		return
	}
	common.RespondWithJSON(c, http.StatusOK, sendTextMessage)
}

// @Summary Receive text message webhook
// @Description Receive text message webhook
// @Tags TextMessage
// @Accept json
// @Produce json
// @Param body query TextMessageWebHookResponse true "Text message webhook response"
// @Success 200 {object} map[string]string
// @Failure 500 {object} common.ErrorResponse
// @Router /text-message/webhooks/delivery-receipt [get]
func (apiCfg *ApiConfig) HandlerTextMessageWebHook(c *gin.Context) {
	var params TextMessageWebHookResponse
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Log received DLR
	_, err := TextMessageWebHook(c, params, apiCfg)
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Error logging database: %v", err))
		return
	}
	fmt.Printf("DLR Received: %+v\n", params)
	c.JSON(http.StatusOK, gin.H{"message": "DLR received"})
}

// @Summary Get text messages
// @Description Get text messages for a consumer
// @Tags TextMessage
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} TextMessage
// @Failure 401 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /text-messages [get]
func (apiCfg *ApiConfig) HandlerGetTextMessages(c *gin.Context, consumer string) {
	textMessages, err := GetTextMessages(c, apiCfg, consumer)
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Error getting text messages: %v", err))
		return
	}
	common.RespondWithJSON(c, http.StatusOK, textMessages)
}

// @Summary Get a text message
// @Description Get a specific text message for a consumer
// @Tags TextMessage
// @Produce json
// @Param messageId path string true "MessageId"
// @Security ApiKeyAuth
// @Success 200 {object} TextMessage
// @Failure 500 {object} common.ErrorResponse
// @Router /text-message/{messageId} [get]
func (apiCfg *ApiConfig) HandlerGetTextMessage(c *gin.Context, consumer string) {
	textMessages, errorCode, err := GetTextMessage(c, apiCfg, consumer)
	if err != nil {
		common.RespondError(c, errorCode, fmt.Sprintf("Error getting text messages: %v", err))
		return
	}
	common.RespondWithJSON(c, http.StatusOK, textMessages)
}
