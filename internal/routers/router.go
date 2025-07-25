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
	router := gin.Default()
	apiCfg := deps.GetEmailConfig()
	apiCfgTextMessage := deps.GetTextMessageConfig()

	// Set up Health check endpoint
	router.GET("/healthz", handlerHealthz)

	// Set up Swagger
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Set up Email and Text Message routes
	emailRouter := router.Group("/email")
	emailRouter.POST("/send", middleware.Auth(apiCfg.HandlerSendEmail))
	emailRouter.GET("/", middleware.Auth(apiCfg.HandlerGetEmails))

	textMessageRouter := router.Group("/text-message")
	textMessageRouter.POST("/send", middleware.Auth(apiCfgTextMessage.HandlerSendTextMessage))
	textMessageRouter.GET("/", middleware.Auth(apiCfgTextMessage.HandlerGetTextMessages))
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
// @Router /healthz [get]
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
