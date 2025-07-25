package email

import (
	"fmt"
	"net/http"

	"github.com/OliPou/gommunication/internal/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap/zapcore"
)

// HandlerSendEmail godoc
// @Summary Send an email
// @Description Sends an email with the provided parameters
// @Tags Email
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param email body EmailParams true "Email parameters"
// @Success 200 {object} Email
// @Failure 500 {object} common.ErrorResponse
// @Router /email/send [post]
func (apiCfg *ApiConfig) HandlerSendEmail(c *gin.Context, consumer string) {
	var params EmailParams
	if err := common.ValidateRequest(c, &params); err != nil {
		common.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Error validating body: %v", err))
		return
	}
	sendEmail, err := SendEmail(c, params, consumer, apiCfg, uuid.New)
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Error logging database: %v", err))
		return
	}

	common.LogClientRequest(c, "email.SendEmail", fmt.Sprintf("Email sent with ID: %s", sendEmail.TransactionUuid), zapcore.InfoLevel)
	common.RespondWithJSON(c, http.StatusOK, sendEmail)

}

// HandlerGetEmails godoc
// @Summary Get all emails
// @Description Retrieves all emails for a specific consumer
// @Tags Email
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} Email
// @Failure 500 {object} common.ErrorResponse
// @Router /email [get]
func (apiCfg *ApiConfig) HandlerGetEmails(c *gin.Context, consumer string) {

	emails, err := GetEmails(c, apiCfg, consumer)
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Error getting emails: %v", err))
		return
	}
	common.LogClientRequest(c, "email.GetEmails", fmt.Sprintf("Retrieved %d emails for consumer: %s", len(emails), consumer), zapcore.InfoLevel)
	common.RespondWithJSON(c, http.StatusOK, emails)
}
