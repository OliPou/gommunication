package main

import (
	"os"

	"github.com/OliPou/gommunication/di"
	_ "github.com/OliPou/gommunication/docs"
	"github.com/OliPou/gommunication/internal/config"
	"github.com/OliPou/gommunication/internal/routers"
	"github.com/gin-contrib/cors"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
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

// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name x-application-id

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	// Initialize logger
	config.InitLogger()
	defer config.Log.Sync()

	// Load env
	if err := config.LoadEnv(); err != nil {
		config.LogClient(nil, "Failed to load environment variables: "+err.Error(), zap.ErrorLevel)
		return err
	}

	// Init DB connection
	if err := config.InitDB(); err != nil {
		config.LogClient(nil, "Failed to initialize database: "+err.Error(), zap.ErrorLevel)
		return err
	}
	defer func() {
		if err := config.DB.Close(); err != nil {
			config.LogClient(nil, "Failed to close DB: "+err.Error(), zap.ErrorLevel)
		}
	}()

	// Initialize dependencies
	deps := di.BuildDependencies()
	// Initialize the router
	router := routers.SetupRouter(deps)
	// Add CORS middleware
	router.Use(cors.New(createCorsConfig()))

	// Get PORT from environment variables with default fallback
	portString := os.Getenv("PORT")
	if portString == "" {
		portString = "8080"
	}

	config.Log.Info("Server starting", zap.String("port", portString))

	// Start the server
	return router.Run(":" + portString)
}

// config is the CORS configuration for the Gin router
func createCorsConfig() cors.Config {
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
	return config
}
