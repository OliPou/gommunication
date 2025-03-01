package textmessage

import (
	"fmt"
	"net/http"

	"github.com/OliPou/gommunication/internal/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

//	func (apiCfg *ApiConfig) HandlerGetEmails(c *gin.Context, consumer string) {
//		emails, err := GetEmails(c, apiCfg, consumer)
//		if err != nil {
//			common.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Error getting emails: %v", err))
//			return
//		}
//		common.RespondWithJSON(c, http.StatusOK, emails)
//	}
func (apiCfg *ApiConfig) HandlerGetTextMessages(c *gin.Context, consumer string) {
	textMessages, err := GetTextMessages(c, apiCfg, consumer)
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Error getting text messages: %v", err))
		return
	}
	common.RespondWithJSON(c, http.StatusOK, textMessages)
}

func (apiCfg *ApiConfig) HandlerGetTextMessage(c *gin.Context, consumer string) {
	textMessages, errorCode, err := GetTextMessage(c, apiCfg, consumer)
	if err != nil {
		common.RespondError(c, errorCode, fmt.Sprintf("Error getting text messages: %v", err))
		return
	}
	common.RespondWithJSON(c, http.StatusOK, textMessages)
}
