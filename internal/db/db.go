package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/vcaldo/tldr-llm-telegram-bot/internal/config"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var db *sql.DB

func InitDB(ctx context.Context, config *config.Config) {
	var err error
	db, err = sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
		return
	}

	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("error pinging database: %v", err)
		return
	}

	ensureTablesExist(ctx)

	log.Println("database connection established successfully.")
}

func GetDB() *sql.DB {
	return db
}

func ensureTablesExist(ctx context.Context) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			message_id BIGINT NOT NULL,
			message_type TEXT NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			chat_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			reply_to_message_id BIGINT,
			first_name TEXT,
			last_name TEXT,
			username TEXT,
			content JSONB NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_message_id ON messages (message_id);`,
		`CREATE INDEX IF NOT EXISTS idx_timestamp ON messages (timestamp);`,
		`CREATE INDEX IF NOT EXISTS idx_message_type ON messages (message_type);`,
		`CREATE INDEX IF NOT EXISTS idx_chat_id ON messages (chat_id);`,
	}

	for _, query := range queries {
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			log.Fatalf("Failed to ensure table exists: %v", err)
		}
	}

	log.Println("All tables ensured to exist")
}
