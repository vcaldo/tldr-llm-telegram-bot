package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/go-telegram/bot/models"
)

func LogMessage(ctx context.Context, db *sql.DB, MessageType string, update *models.Update, content interface{}) error {
	contentJSON, err := json.Marshal(content)
	if err != nil {
		log.Printf("Failed to marshal content: %v", err)
		return err
	}

	// Check if the update is a reply to another message
	var replyToMessageID *int64
	if update.Message.ReplyToMessage != nil {
		id := int64(update.Message.ReplyToMessage.ID)
		replyToMessageID = &id
	}

	// Convert Unix timestamp to time.Time
	messageTime := time.Unix(int64(update.Message.Date), 0)

	query := `
		INSERT INTO messages (message_id, message_type, timestamp, chat_id, user_id, reply_to_message_id, first_name, last_name, username, content)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = db.ExecContext(ctx, query,
		update.Message.ID,
		MessageType,
		messageTime,
		update.Message.Chat.ID,
		update.Message.From.ID,
		replyToMessageID,
		update.Message.From.FirstName,
		update.Message.From.LastName,
		update.Message.From.Username,
		contentJSON,
	)

	return err
}
