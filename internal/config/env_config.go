package config

import (
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		Log.Error("Error loading .env file", zap.Error(err))
		return err
	}

	return nil
}
