package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// ConnectWithRetry opens a PostgreSQL connection and retries up to maxRetries times
// using exponential backoff (1s → 2s → 4s → 8s → 16s …).
// It returns the *sql.DB on success or an error after all attempts are exhausted.
func ConnectWithRetry(dsn string, maxRetries int) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("🔄 Database connection attempt %d/%d...", attempt, maxRetries)

		db, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("⚠️  Failed to open database: %v. Retrying in %ds...", err, 1<<uint(attempt-1))
		} else if pingErr := db.Ping(); pingErr != nil {
			err = pingErr
			db.Close()
			db = nil
			waitSeconds := 1 << uint(attempt-1)
			if attempt < maxRetries {
				log.Printf("⚠️  Connection failed: %v. Retrying in %ds...", err, waitSeconds)
				time.Sleep(time.Duration(waitSeconds) * time.Second)
			}
		} else {
			log.Println("✅ Database connected successfully!")
			return db, nil
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}
