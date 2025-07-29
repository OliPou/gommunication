package config

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger initializes the global logger instance with a console encoder configuration.
// The logger outputs logs to standard output with ISO8601 timestamps, capitalized log levels,
// short caller information, and string-encoded durations. The log level is set to Info.
// This function assigns the configured logger to the global Log variable.
func InitLogger() {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:     "timestamp",
		LevelKey:    "level",
		MessageKey:  "message",
		NameKey:     "logger",
		FunctionKey: "func",

		EncodeName:     zapcore.FullNameEncoder,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	Log = zap.New(core, zap.AddCaller(), zap.WithCaller(true))
}

func LogClient(c *gin.Context, message string, level zapcore.Level) {
	clientIP := getClientIP(c)
	requestID := extractRequestID(c)
	tenantID := "no-tenant-id"
	organizationID := "no-organization-id"

	logLine := fmt.Sprintf("[%s] [%s] [%s]-[%s] %s", clientIP, requestID, tenantID, organizationID, message)

	logger := Log.WithOptions(zap.AddCallerSkip(1))

	switch level {
	case zapcore.DebugLevel:
		logger.Debug(logLine)
	case zapcore.InfoLevel:
		logger.Info(logLine)
	case zapcore.WarnLevel:
		logger.Warn(logLine)
	case zapcore.ErrorLevel:
		logger.Error(logLine)
	default:
		logger.Info(logLine)
	}
}

func getClientIP(c *gin.Context) string {
	if c == nil {
		return "no-client-ip"
	}
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

func extractRequestID(c *gin.Context) string {
	if c == nil {
		return "no-request-id"
	}

	if id := c.GetHeader("X-Kong-Request-ID"); id != "" {
		return id
	}
	return "no-request-id"
}
