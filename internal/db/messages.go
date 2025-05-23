package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-telegram/bot/models"
	"github.com/lib/pq"
)

func LogMessage(ctx context.Context, db *sql.DB, MessageType string, update *models.Update, content interface{}) error {
	contentJSON, err := json.Marshal(content)
	if err != nil {
		log.Printf("Failed to marshal content: %v", err)
		return err
	}

	var replyToMessageID *int64
	if update.Message.ReplyToMessage != nil {
		id := int64(update.Message.ReplyToMessage.ID)
		replyToMessageID = &id
	}

	displayName := getDisplayName(update)

	messageTime := time.Unix(int64(update.Message.Date), 0)

	query := `
		INSERT INTO messages (message_id, message_type, timestamp, chat_id, user_id, reply_to_message_id, first_name, last_name, username, display_name, content, moderated)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

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
		displayName,
		contentJSON,
		false,
	)

	return err
}

func FetchMessagesSince(ctx context.Context, db *sql.DB, chatID, messageID int64, since *time.Time, interval time.Duration) ([]Message, error) {
	endTime := since.Add(interval)

	query := `
		SELECT message_id, message_type, timestamp, chat_id, user_id, reply_to_message_id, first_name, last_name, username, display_name, content, moderated
		FROM messages
		WHERE chat_id = $1
		  AND timestamp BETWEEN $2 AND $3
		  AND message_id >= $4
		ORDER BY timestamp ASC
		LIMIT 2048`

	rows, err := db.QueryContext(ctx, query, chatID, since, endTime, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(
			&msg.MessageID,
			&msg.MessageType,
			&msg.Timestamp,
			&msg.ChatID,
			&msg.UserID,
			&msg.ReplyToMessageID,
			&msg.FirstName,
			&msg.LastName,
			&msg.Username,
			&msg.DisplayName,
			&msg.Content,
			&msg.Moderated,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

func FetchUnmoderatedMessages(ctx context.Context, db *sql.DB, chatID int64) ([]Message, error) {
	query := `
		SELECT message_id, message_type, timestamp, chat_id, user_id, reply_to_message_id, first_name, last_name, username, display_name, content, moderated
		FROM messages
		WHERE chat_id = $1 AND moderated = false
		ORDER BY timestamp ASC
		LIMIT 2048`

	rows, err := db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(
			&msg.MessageID,
			&msg.MessageType,
			&msg.Timestamp,
			&msg.ChatID,
			&msg.UserID,
			&msg.ReplyToMessageID,
			&msg.FirstName,
			&msg.LastName,
			&msg.Username,
			&msg.DisplayName,
			&msg.Content,
			&msg.Moderated,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

func SetMessagesModerated(ctx context.Context, db *sql.DB, messages []Message) error {
	if len(messages) == 0 {
		return nil
	}

	var messageIDs []int64
	for _, msg := range messages {
		messageIDs = append(messageIDs, msg.MessageID)
	}

	query := `UPDATE messages SET moderated = TRUE WHERE message_id = ANY($1)`
	_, err := db.ExecContext(ctx, query, pq.Array(messageIDs))
	if err != nil {
		return fmt.Errorf("failed to update messages as moderated: %w", err)
	}

	return nil
}
