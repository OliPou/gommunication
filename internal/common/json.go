package common

import (
	"fmt"

	"github.com/OliPou/gommunication/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zapcore"
)

func RespondWithJSON(c *gin.Context, status int, payload interface{}) {
	// Set HTTP status code
	c.Status(status)

	// If payload is nil, just return with status code
	if payload == nil {
		return
	}

	// Set JSON content type and send response
	c.Header("Content-Type", "application/json")

	// Gin's JSON method handles marshalling and error checking internally
	c.JSON(status, payload)
}

func RespondError(c *gin.Context, status int, message string) {

	logMsg := fmt.Sprintf("Error response: %s", message)

	var level zapcore.Level
	switch {
	case status >= 500:
		level = zapcore.ErrorLevel
	case status >= 400:
		level = zapcore.WarnLevel
	default:
		level = zapcore.InfoLevel
	}

	config.LogClient(c, logMsg, level)

	RespondWithJSON(c, status, map[string]string{"error": message})
}

// ErrorResponse represents the structure of an error response.
type ErrorResponse struct {
	Message string `json:"message"`
}
