package config

import (
	"os"

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
		TimeKey:        "timestamp",
		LevelKey:       "level",
		MessageKey:     "message",
		CallerKey:      "caller",
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

	Log = zap.New(core, zap.AddCaller())
}
