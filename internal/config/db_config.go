package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

var DB *sql.DB

func InitDB() {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	fmt.Println("Database URL:", dbUrl)

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Check if the database is ready
	if err := checkDatabase(db); err != nil {
		log.Fatal("Database is not ready:", err)
	}
	fmt.Println("Database connection established successfully")
	DB = db
}

// checkDatabase tries to ping the database until it succeeds or times out
func checkDatabase(db *sql.DB) error {
	for i := 0; i < 10; i++ {
		err := db.Ping()
		if err == nil {
			return nil
		}
		fmt.Println("Waiting for database to be ready...")
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("database is not ready")
}
