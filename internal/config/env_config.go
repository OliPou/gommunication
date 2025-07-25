package config

import (
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// LoadEnv loads environment variables from a .env file using the godotenv package.
// If the .env file cannot be loaded, it logs the error and returns it.
// Returns nil if the environment variables are loaded successfully.
func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		Log.Error("Error loading .env file", zap.Error(err))
		return err
	}

	return nil
}
