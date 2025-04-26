package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/config"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/constants"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/db"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/llm"
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

func SendLongMessage(ctx context.Context, nrApp *newrelic.Application, b *bot.Bot, chatID int64, text string) {
	txn := nrApp.StartTransaction("telegram:send-long-message")
	defer txn.End()

	txn.AddAttribute("chatID", chatID)
	txn.AddAttribute("messageLength", len(text))

	txnCtx := newrelic.NewContext(ctx, txn)

	if len(text) <= telegramMaxMessageLength {
		segment := txn.StartSegment("telegram:send-message")
		if _, err := b.SendMessage(txnCtx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		}); err != nil {
			txn.NoticeError(err)
			log.Printf("error sending message %v", err)
		}
		segment.End()
		return
	}

	runes := []rune(text)
	for i := 0; i < len(runes); i += telegramMaxMessageLength {
		segment := txn.StartSegment("telegram:send-message-chunk")

		end := i + telegramMaxMessageLength
		if end > len(runes) {
			end = len(runes)
		}
		messageChunk := string(runes[i:end])

		if _, err := b.SendMessage(txnCtx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      messageChunk,
			ParseMode: models.ParseModeHTML,
		}); err != nil {
			txn.NoticeError(err)
			log.Printf("error sending message chunk: %v", err)
			segment.End()
			return
		}
		segment.End()
	}
}

func LoadPrompts(llmClient *llm.LLMClient, config *config.Config) ([]string, error) {
	err := llm.LoadPrompts(config.PromptsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompts: %w", err)
	}

	var prompts []string

	summaryPrompt, err := llm.GetPrompt("summary", config.Language)
	if err != nil {
		log.Fatalf("failed to get prompt: %v", err)
	}
	prompts = append(prompts, summaryPrompt)

	problematicPrompt, err := llm.GetPrompt("problematic", config.Language)
	if err != nil {
		log.Fatalf("failed to get prompt: %v", err)
	}
	prompts = append(prompts, problematicPrompt)

	valueAssessmentPrompt, err := llm.GetPrompt("value_assessment", config.Language)
	if err != nil {
		log.Fatalf("failed to get prompt: %v", err)
	}
	prompts = append(prompts, valueAssessmentPrompt)

	sportsSchedulePrompt, err := llm.GetPrompt("sports_schedule", config.Language)
	if err != nil {
		log.Fatalf("failed to get prompt: %v", err)
	}
	prompts = append(prompts, sportsSchedulePrompt)

	return prompts, nil
}
