package email

import (
	"fmt"
	"net/http"

	"github.com/OliPou/gommunication/internal/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
	common.RespondWithJSON(c, http.StatusOK, sendEmail)

}

func (apiCfg *ApiConfig) HandlerGetEmails(c *gin.Context, consumer string) {
	emails, err := GetEmails(c, apiCfg, consumer)
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Error getting emails: %v", err))
		return
	}
	common.RespondWithJSON(c, http.StatusOK, emails)
}
