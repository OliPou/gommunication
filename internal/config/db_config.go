package config

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

var DB *sql.DB

const (
	maxAttempts = 10
	retryDelay  = 2 * time.Second
)

func InitDB() error {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		Log.Error("DB_URL environment variable is not set")
		return fmt.Errorf("DB_URL environment variable is not set")
	}

	Log.Info("Connecting to database...", zap.String("db_url", dbUrl))

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		Log.Error("Failed to open DB connection", zap.Error(err))
		return err
	}

	// Check if DB is ready (ping)
	if err := checkDatabase(db); err != nil {
		Log.Error("Database is not ready", zap.Error(err))
		return err
	}

	Log.Info("Database connection established successfully")
	DB = db
	return nil
}

// checkDatabase tries to ping the database until it succeeds or times out
func checkDatabase(db *sql.DB) error {
	for i := range maxAttempts {
		err := db.Ping()
		if err == nil {
			return nil
		}

		// If the error is fatal, log it and return
		if isFatalDatabaseError(err) {
			Log.Error("Fatal database error — aborting retries", zap.Error(err))
			return err
		}

		// Log non-fatal errors and retry
		Log.Warn("Database is not ready, retrying...",
			zap.Int("attempt", i+1),
			zap.Int("max_attempts", maxAttempts),
			zap.Error(err),
		)

		time.Sleep(retryDelay)
	}

	return fmt.Errorf("database is not ready after %d attempts", maxAttempts)
}

func isFatalDatabaseError(err error) bool {
	lowered := strings.ToLower(err.Error())

	fatalErrors := []string{
		"password authentication failed",
		"does not exist",
		"connection refused",
		"could not connect to server",
		"invalid connection string",
		"ssl connection error",
	}
	for _, substr := range fatalErrors {
		if strings.Contains(lowered, substr) {
			return true
		}
	}
	return false
}
