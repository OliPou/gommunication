package middleware

import (
	"fmt"
	"net/http"

	"github.com/OliPou/gommunication/internal/common"
	"github.com/OliPou/gommunication/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	// MaxRequestSize limits the entire HTTP request size (including JSON payload)
	MaxRequestSize = 50 * 1024 * 1024 // 50MB total request size
)

// RequestSizeLimitMiddleware limits the size of incoming HTTP requests
// This prevents large payloads from consuming too much memory before validation
func RequestSizeLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Limit request body size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxRequestSize)

		// Check if request exceeds limit
		if c.Request.ContentLength > MaxRequestSize {
			config.LogClient(nil, fmt.Sprintf("Request size too large: %d bytes (max: %d)",
				c.Request.ContentLength, MaxRequestSize), zap.WarnLevel)

			c.JSON(http.StatusRequestEntityTooLarge, common.ErrorResponse{
				Message: fmt.Sprintf("Request too large. Maximum size allowed: %dMB", MaxRequestSize/(1024*1024)),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AttachmentValidationMiddleware can be added for early attachment validation
func AttachmentValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware can be extended to do early attachment validation
		// before the request reaches the handler
		c.Next()
	}
}
