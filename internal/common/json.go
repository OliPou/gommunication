package common

import (
	"log"

	"github.com/gin-gonic/gin"
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
	if status > 499 {
		log.Printf("Responding with 5xx error: %s", message)
	}
	RespondWithJSON(c, status, map[string]string{"error": message})
}
