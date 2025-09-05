package email

import (
	"fmt"
	"net/http"

	"github.com/OliPou/gommunication/internal/common"
	"github.com/OliPou/gommunication/internal/config"
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
// @Param bu path string  true  "Business Unit"
// @Param email body EmailParams true "Email parameters"
// @Success 200 {object} Email
// @Failure 500 {object} common.ErrorResponse
// @Router /{bu}/email/send [post]
func (apiCfg *ApiConfig) HandlerSendEmail(c *gin.Context, consumer string) {
	var params EmailParams
	if err := common.ValidateRequest(c, &params); err != nil {
		common.RespondError(c, http.StatusBadRequest, fmt.Sprintf("Error validating body: %v", err))
		return
	}
	bu := c.GetString("bu")
	sendEmail, err := SendEmail(c, params, consumer, bu, apiCfg, uuid.New)
	if err != nil {
		if reqErr, ok := err.(*common.RequestError); ok {
			common.RespondError(c, reqErr.StatusCode, reqErr.Message)
		} else {
			common.RespondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	config.LogClient(c, fmt.Sprintf("Email sent with ID: %s", sendEmail.TransactionUuid), zapcore.InfoLevel)
	common.RespondWithJSON(c, http.StatusOK, sendEmail)

}

// HandlerGetEmails godoc
// @Summary Get all emails
// @Description Retrieves all emails for a specific consumer
// @Tags Email
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param bu path string  true  "Business Unit"
// @Success 200 {array} Email
// @Failure 500 {object} common.ErrorResponse
// @Router /{bu}/emails [get]
func (apiCfg *ApiConfig) HandlerGetEmails(c *gin.Context, consumer string) {

	emails, err := GetEmails(c, apiCfg, consumer)
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Error getting emails: %v", err))
		return
	}
	config.LogClient(c, fmt.Sprintf("Retrieved %d emails for consumer: %s with Business Unit: %s", len(emails), consumer, c.GetString("bu")), zapcore.InfoLevel)
	common.RespondWithJSON(c, http.StatusOK, emails)
}

// CreateSubdomainOwnership godoc
// @Summary Create a subdomain ownership
// @Description Creates a subdomain ownership record for a consumer
// @Tags Email
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param bu path string  true  "Business Unit"
// @Param subdomainOwnership body SubdomainOwnership true "Subdomain ownership parameters"
// @Success 200 {object} SubdomainOwnership
// @Failure 500 {object} common.ErrorResponse
// @Router /{bu}/email/subdomain-ownerships [post]
func (apiCfg *ApiConfig) CreateSubdomainOwnership(c *gin.Context, consumer string) {
	var params SubdomainOwnership
	if err := common.ValidateRequest(c, &params); err != nil {
		common.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Error validating body: %v", err))
		return
	}
	subdomainOwnership, err := CreateSubdomainOwnership(c, params, consumer, c.GetString("bu"), apiCfg)
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	config.LogClient(c, fmt.Sprintf("Subdomain ownership created with ID: %d", subdomainOwnership.SubdomainOwnershipUUID), zapcore.InfoLevel)
	common.RespondWithJSON(c, http.StatusOK, subdomainOwnership)
}

// HandlerSendGridWebhook godoc
// @Summary Handle SendGrid webhook events
// @Description Processes events from SendGrid webhook, such as email opens
// @Tags Email
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} common.ErrorResponse
// @Router /sendgrid/webhooks/event [post]
// @Param events body []SendGridEvent true "SendGrid events"
// @Security ApiKeyAuth
func (apiCfg *ApiConfig) HandlerSendGridWebhook(c *gin.Context) {
	var events []SendGridEvent

	if !VerifiedSignatureSendGrid(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
		return
	}

	if err := c.BindJSON(&events); err != nil {
		config.LogClient(c, "Error binding JSON for SendGrid webhook: "+err.Error(), zapcore.ErrorLevel)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	for _, event := range events {
		if event.Event == "open" {
			config.LogClient(c, fmt.Sprintf("Processing open event for email: %s", event.Email), zapcore.InfoLevel)
			if err := UpdateEmailOpened(c, apiCfg, event.TransactionUUID, true); err != nil {
				config.LogClient(c, "Error updating email opened status: "+err.Error(), zapcore.ErrorLevel)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update email opened status"})
				return
			}
			config.LogClient(c, fmt.Sprintf("Email opened for transaction UUID: %s", event.TransactionUUID), zapcore.InfoLevel)
		} else {
			config.LogClient(c, fmt.Sprintf("Received non-open event: %s for email: %s", event.Event, event.Email), zapcore.InfoLevel)
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})

}
