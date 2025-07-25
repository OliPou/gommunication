package common

import (
	"fmt"

	"github.com/OliPou/gommunication/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zapcore"
)

func LogClientRequest(c *gin.Context, serviceMethod string, message string, level zapcore.Level) {
	clientIP := getClientIP(c)
	requestID := c.GetHeader("X-Kong-Request-ID")

	logLine := fmt.Sprintf(
		"[%-15s] %s, %s - %s",
		clientIP,
		serviceMethod,
		requestID,
		message,
	)

	switch level {
	case zapcore.DebugLevel:
		config.Log.Debug(logLine)
	case zapcore.InfoLevel:
		config.Log.Info(logLine)
	case zapcore.WarnLevel:
		config.Log.Warn(logLine)
	case zapcore.ErrorLevel:
		config.Log.Error(logLine)
	default:
		config.Log.Info(logLine)
	}
}

func getClientIP(c *gin.Context) string {
	// Try to get the IP from the X-Forwarded-For header
	ip := c.GetHeader("X-Forwarded-For")
	if ip != "" {
		return ip
	}

	// Fallback to RemoteAddr if X-Forwarded-For is not set
	ip = c.Request.RemoteAddr
	if ip != "" {
		return ip
	}

	return "unknown"
}
