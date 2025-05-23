package db

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/newrelic/go-agent/v3/integrations/nrpq"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/config"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB(ctx context.Context, config *config.Config, nrApp *newrelic.Application) {
	var err error
	db, err = sql.Open("nrpostgres", config.DatabaseURL)
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

func CloseDB() {
	if err := db.Close(); err != nil {
		log.Fatalf("error closing the database: %v", err)
	}
	log.Println("database connection closed.")
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
			display_name TEXT,
			content JSONB NOT NULL,
			moderated BOOLEAN DEFAULT FALSE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_timestamp ON messages (timestamp);`,
		`CREATE INDEX IF NOT EXISTS idx_message_id ON messages (message_id);`,
		`CREATE INDEX IF NOT EXISTS idx_chat_id ON messages (chat_id);`,
		`CREATE INDEX IF NOT EXISTS idx_message_type ON messages (message_type);`,
		`CREATE INDEX IF NOT EXISTS idx_moderated ON messages (moderated);`,
	}

	for _, query := range queries {
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			log.Fatalf("failed to ensure table exists: %v", err)
		}
	}

	log.Println("all tables and indexes ensured to exist")
}
