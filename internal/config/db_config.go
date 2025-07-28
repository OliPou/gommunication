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

// InitDB initializes the database connection using the DB_URL environment variable.
// It attempts to open a connection to a PostgreSQL database and checks if the database is ready.
// If successful, it assigns the database connection to the global DB variable.
// Returns an error if the environment variable is not set, the connection cannot be opened,
// or the database is not ready.
func InitDB() error {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		LogClient(nil, "DB_URL environment variable is not set", zap.ErrorLevel)
		return fmt.Errorf("DB_URL environment variable is not set")
	}

	LogClient(nil, "Connecting to database...", zap.ErrorLevel)

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		LogClient(nil, "Failed to open DB connection: "+err.Error(), zap.ErrorLevel)
		return err
	}

	// Check if DB is ready (ping)
	if err := checkDatabase(db); err != nil {
		LogClient(nil, "Database is not ready: "+err.Error(), zap.ErrorLevel)
		return err
	}

	LogClient(nil, "Database connection established successfully", zap.InfoLevel)
	DB = db
	return nil
}

// checkDatabase attempts to ping the provided database connection up to maxAttempts times,
// waiting retryDelay between each attempt. If a fatal database error is encountered,
// it logs the error and aborts further retries. Non-fatal errors are logged as warnings,
// and the function retries until the maximum number of attempts is reached. If the database
// is still not ready after all attempts, an error is returned.
func checkDatabase(db *sql.DB) error {
	for i := range maxAttempts {
		err := db.Ping()
		if err == nil {
			return nil
		}

		// If the error is fatal, log it and return
		if isFatalDatabaseError(err) {
			LogClient(nil, "Fatal database error: "+err.Error(), zap.ErrorLevel)
			LogClient(nil, "Fatal database error — aborting retries", zap.ErrorLevel)
			return err
		}

		// Log non-fatal errors and retry
		LogClient(nil, fmt.Sprintf("Attempt %d/%d: Database not ready, retrying...", i+1, maxAttempts), zap.WarnLevel)
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("database is not ready after %d attempts", maxAttempts)
}

// isFatalDatabaseError checks whether the provided error is considered fatal for database operations.
// It inspects the error message (case-insensitive) for known substrings that indicate fatal issues,
// such as authentication failures, missing databases, connection problems, or invalid connection strings.
// Returns true if the error matches any of these fatal conditions, otherwise returns false.
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
