package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/OliPou/gommunication/email"
	"github.com/OliPou/gommunication/internal/common"
	"github.com/OliPou/gommunication/internal/database"
	"github.com/OliPou/gommunication/middleware"
	"github.com/OliPou/gommunication/textmessage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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

	var ginRouterGroupName string = os.Getenv("GIN_ROUTER_GROUP_NAME")

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
	}
	config.AllowCredentials = true
	config.ExposeHeaders = []string{"Content-Length"}
	config.MaxAge = 12 * 60 * 60 // 12 hours

	// Add CORS middleware
	router.Use(cors.New(config))

	v1Router := router.Group(fmt.Sprintf("/%s", ginRouterGroupName))
	v1Router.GET("/healthz", handlerHealthz)
	v1Router.POST("/send-email", middleware.Auth(apiCfg.HandlerSendEmail))
	v1Router.GET("/get-send-email", middleware.Auth(apiCfg.HandlerGetEmails))
	v1Router.POST("/send-text-message", middleware.Auth(apiCfgTextMessage.HandlerSendTextMessage))
	v1Router.GET("/text-messages-sent", middleware.Auth(apiCfgTextMessage.HandlerGetTextMessages))
	v1Router.GET("/webhooks/delivery-receipt", apiCfgTextMessage.HandlerTextMessageWebHook)
	v1Router.GET("/text-message-sent/:messageId", middleware.Auth(apiCfgTextMessage.HandlerGetTextMessage))

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
