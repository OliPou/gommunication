package main

import (
	"fmt"
	"os"

	"github.com/OliPou/gommunication/di"
	_ "github.com/OliPou/gommunication/docs"
	"github.com/OliPou/gommunication/internal/config"
	"github.com/OliPou/gommunication/internal/routers"
	"github.com/gin-contrib/cors"
	_ "github.com/lib/pq"
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
	// Load env
	config.LoadEnv()
	// Init DB connection
	config.InitDB()
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

	fmt.Printf("Server starting on port: %s\n", portString)

	// Start the server
	if err := router.Run(":" + portString); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
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
