package routers

import (
	"net/http"

	"github.com/OliPou/gommunication/di"
	"github.com/OliPou/gommunication/internal/common"
	"github.com/OliPou/gommunication/middleware"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(deps *di.AppDependencies) *gin.Engine {
	common.RegisterValidators()
	router := gin.Default()
	apiCfg := deps.GetEmailConfig()
	apiCfgTextMessage := deps.GetTextMessageConfig()

	// Set up Health check endpoint
	router.GET("/health", handlerHealthz)

	// Set up Swagger
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Set up Business Unit routes
	bu := router.Group("/:bu",
		middleware.BUValidator(),
	)
	// Set up Email and Text Message routes
	emailsRouter := bu.Group("/emails")
	emailsRouter.GET("/", middleware.Auth(apiCfg.HandlerGetEmails))

	emailRouter := bu.Group("/email")
	emailRouter.POST("/send", middleware.Auth(apiCfg.HandlerSendEmail))
	emailRouter.POST("/subdomain-ownerships", middleware.Auth(apiCfg.CreateSubdomainOwnership))

	sendgridRouter := router.Group("/sendgrid")
	sendgridRouter.POST("/webhooks/event", apiCfg.HandlerSendGridWebhook)

	textMessagesRouter := router.Group("/text-messages")
	textMessagesRouter.GET("/", middleware.Auth(apiCfgTextMessage.HandlerGetTextMessages))

	textMessageRouter := router.Group("/text-message")
	textMessageRouter.POST("/send", middleware.Auth(apiCfgTextMessage.HandlerSendTextMessage))
	textMessageRouter.GET("/webhooks/delivery-receipt", apiCfgTextMessage.HandlerTextMessageWebHook)
	textMessageRouter.GET("/:messageId", middleware.Auth(apiCfgTextMessage.HandlerGetTextMessage))

	return router
}

// @Summary Health check
// @Description Check if the server is running
// @Tags Health
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func handlerHealthz(c *gin.Context) {
	status := struct {
		Status string `json:"status"`
		Ready  bool   `json:"ready"`
	}{
		Status: "ok",
		Ready:  true,
	}
	common.RespondWithJSON(c, http.StatusOK, status)
}
