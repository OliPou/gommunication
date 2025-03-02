package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/OliPou/gommunication/docs"
	"github.com/OliPou/gommunication/email"
	"github.com/OliPou/gommunication/internal/common"
	"github.com/OliPou/gommunication/internal/database"
	"github.com/OliPou/gommunication/middleware"
	"github.com/OliPou/gommunication/textmessage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Your Project API
// @version 1.0
// @description This is a sample server for a pet store.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name x-application-id
func main() {
	// Set Gin to release mode
	// gin.SetMode(gin.ReleaseMode)
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file:", err)
		// Continue execution as .env file might not exist in production
	}

	// Get PORT from environment variables with default fallback
	portString := os.Getenv("PORT")
	if portString == "" {
		portString = "8080"
	}

	// var ginRouterGroupName string = os.Getenv("GIN_ROUTER_GROUP_NAME")

	dbUrl := os.Getenv("DB_URL")
	sendGridAPIKey := os.Getenv("API_KEY")
	vonageApiKey := os.Getenv("VONAGE_API_KEY")
	vonageApiSecret := os.Getenv("VONAGE_API_SECRET")
	vonageApiUrl := os.Getenv("VONAGE_API_URL")
	vonageApiCallBackUrl := os.Getenv("VONAGE_API_CALL_BACK_URL")

	if dbUrl == "" {
		log.Fatal("DB_URL not found in environment variables")
	}
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Check if the database is ready
	if err := checkDatabase(db); err != nil {
		log.Fatal("Database is not ready:", err)
	}

	dbQueries := database.New(db)

	apiCfg := &email.ApiConfig{
		DB:     dbQueries,
		ApiKey: sendGridAPIKey,
	}

	apiCfgTextMessage := &textmessage.ApiConfig{
		DB:                   dbQueries,
		ApiKey:               vonageApiKey,
		ApiSecret:            vonageApiSecret,
		VonageApiUrl:         vonageApiUrl,
		VonageApiCallBackUrl: vonageApiCallBackUrl,
	}

	fmt.Printf("Server starting on port: %s\n", portString)

	// Initialize the router
	router := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()

	// Get allowed origins from environment variable or use default
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		// For development, you might want to allow all origins
		config.AllowAllOrigins = true
	} else {
		config.AllowOrigins = []string{allowedOrigins}
	}

	// Additional CORS configurations
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"x-application-id",
	}
	config.AllowCredentials = true
	config.ExposeHeaders = []string{"Content-Length"}
	config.MaxAge = 12 * 60 * 60 // 12 hours

	// Add CORS middleware
	router.Use(cors.New(config))

	emailRouter := router.Group(fmt.Sprintf("/email"))
	textMessageRouter := router.Group(fmt.Sprintf("/text-message"))
	router.GET("/healthz", handlerHealthz)
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	emailRouter.POST("/send", middleware.Auth(apiCfg.HandlerSendEmail))
	emailRouter.GET("/", middleware.Auth(apiCfg.HandlerGetEmails))
	textMessageRouter.POST("/send", middleware.Auth(apiCfgTextMessage.HandlerSendTextMessage))
	textMessageRouter.GET("/", middleware.Auth(apiCfgTextMessage.HandlerGetTextMessages))
	textMessageRouter.GET("/webhooks/delivery-receipt", apiCfgTextMessage.HandlerTextMessageWebHook)
	textMessageRouter.GET("/:messageId", middleware.Auth(apiCfgTextMessage.HandlerGetTextMessage))

	// Start the server
	if err := router.Run(":" + portString); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}

// checkDatabase tries to ping the database until it succeeds or times out
func checkDatabase(db *sql.DB) error {
	for i := 0; i < 10; i++ {
		err := db.Ping()
		if err == nil {
			return nil
		}
		fmt.Println("Waiting for database to be ready...")
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("database is not ready")
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
