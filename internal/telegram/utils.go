package telegram

import (
	"fmt"

	"github.com/vcaldo/tldr-llm-telegram-bot/internal/constants"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/db"
)

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
