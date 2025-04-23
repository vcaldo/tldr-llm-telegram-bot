package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/constants"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/db"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/llm"
)

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var err error
	switch {
	case update.Message != nil && update.Message.Text != "":
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeText, update, update.Message.Text)
	case update.Message != nil && update.Message.Voice != nil:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeVoice, update, update.Message.Voice)
	case update.Message != nil && update.Message.Photo != nil:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypePhoto, update, update.Message.Photo)
	case update.Message != nil && update.Message.Animation != nil:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeAnimation, update, update.Message.Animation)
	case update.Message != nil && update.Message.Sticker != nil:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeSticker, update, update.Message.Sticker)
	case update.Message != nil && update.Message.Video != nil:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeVideo, update, update.Message.Video)
	case update.Message != nil && update.Message.VideoNote != nil:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeVideoNote, update, update.Message.VideoNote)
	default:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeGeneric, update, update.Message)
	}
	if err != nil {
		log.Printf("error logging message: %v", err)
	}
}

func tldrHandler(llmClient llm.LLMClient, summaryPrompt string) func(ctx context.Context, b *bot.Bot, update *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message.Text == "" || update.Message.ReplyToMessage == nil {
			return
		}

		firstMessageTimestamp, err := getMessageTimestamp(db.GetDB(), int64(update.Message.ReplyToMessage.ID), update.Message.Chat.ID)
		if err != nil {
			log.Printf("error getting message timestamp: %v", err)
			return
		}

		messages, err := db.FetchMessagesSince(ctx, db.GetDB(), update.Message.Chat.ID, int64(update.Message.ReplyToMessage.ID), firstMessageTimestamp, 720*time.Minute)
		if err != nil {
			log.Printf("error fetching messages: %v", err)
			return
		}

		if len(messages) == 0 {
			log.Printf("no messages found since %v", firstMessageTimestamp)
			return
		}

		formattedMessages := formatTextMessages(messages)
		prompt := fmt.Sprintf("%s %s", summaryPrompt, formattedMessages)

		summary, err := llmClient.AnalyzePrompt(prompt)
		if err != nil {
			log.Printf("error generating summary: %v", err)
			return
		}

		log.Printf("generated summary: %s", summary)

		SendLongMessage(ctx, b, update.Message.Chat.ID, summary)
	}
}

func problematicSpeechHandler(llmClient llm.LLMClient, problematicPrompt string) func(ctx context.Context, b *bot.Bot, update *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message.Text == "" {
			return
		}

		messages, err := db.FetchUnmoderatedMessages(ctx, db.GetDB(), update.Message.Chat.ID)
		if err != nil {
			log.Printf("error fetching messages: %v", err)
			return
		}

		if len(messages) == 0 {
			log.Printf("no unmoderated messages found")
			return
		}

		formattedMessages := formatTextMessages(messages)

		prompt := fmt.Sprintf("%s %s", problematicPrompt, formattedMessages)
		problematicContent, err := llmClient.AnalyzePrompt(prompt)

		if err != nil {
			log.Printf("error generating problematic content: %v", err)
			return
		}

		log.Printf("generated problematic content: %v", len(problematicContent))
		if len(problematicContent) > 4 {
			SendLongMessage(ctx, b, update.Message.Chat.ID, problematicContent)
		}

		// set the messages as moderated
		if err := db.SetMessagesModerated(ctx, db.GetDB(), messages); err != nil {
			log.Printf("error setting messages as moderated: %v", err)
		}
	}
}
func getMessageTimestamp(db *sql.DB, messageID int64, groupID int64) (*time.Time, error) {
	query := `SELECT timestamp FROM messages WHERE message_id = $1 AND chat_id = $2`
	row := db.QueryRow(query, messageID, groupID)
	var timestamp time.Time
	if err := row.Scan(&timestamp); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message with ID %d not found in chat %d", messageID, groupID)
		}
		return nil, err
	}
	return &timestamp, nil
}
