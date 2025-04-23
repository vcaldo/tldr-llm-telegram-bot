package telegram

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/constants"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/db"
)

const telegramMaxMessageLength = 4096

func formatTextMessages(messages []db.Message) string {
	var formattedMessages string

	for _, msg := range messages {
		if msg.MessageType == constants.MessageTypeText {
			if msg.ReplyToMessageID != nil {
				formattedMessages += fmt.Sprintf(
					"%s %d %s %s replied to %d: %s\n",
					msg.Timestamp.Format("2006-01-02 15:04:05"),
					msg.MessageID,
					msg.DisplayName,
					msg.Username,
					*msg.ReplyToMessageID,
					msg.Content,
				)
			} else {
				formattedMessages += fmt.Sprintf(
					"%s %d %s %s said: %s\n",
					msg.Timestamp.Format("2006-01-02 15:04:05"),
					msg.MessageID,
					msg.DisplayName,
					msg.Username,
					msg.Content,
				)
			}
		}
	}

	return formattedMessages
}

func SendLongMessage(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	if len(text) <= telegramMaxMessageLength {
		// Envia a mensagem inteira se for menor ou igual ao limite
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		}); err != nil {
			log.Printf("error sending message %v", err)
		}
		return
	}

	runes := []rune(text)
	for i := 0; i < len(runes); i += telegramMaxMessageLength {
		end := i + telegramMaxMessageLength
		if end > len(runes) {
			end = len(runes)
		}
		messageChunk := string(runes[i:end])

		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      messageChunk,
			ParseMode: models.ParseModeHTML,
		}); err != nil {
			log.Printf("error sending message chunk: %v", err)
			return
		}
	}
}
