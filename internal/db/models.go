package db

import "time"

type Message struct {
	MessageID        int64     `json:"message_id"`
	MessageType      string    `json:"message_type"`
	Timestamp        time.Time `json:"timestamp"`
	ChatID           int64     `json:"chat_id"`
	UserID           int64     `json:"user_id"`
	ReplyToMessageID int64     `json:"reply_to_message_id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Username         string    `json:"username"`
	Content          string    `json:"content"`
}
