package telegram

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/db"
)

func ScheduledProblematic(ctx context.Context, b *bot.Bot) {
	// Get all problematic messages from the database
	problematicMessages, err := db.GetProblematicMessages(ctx, db.GetDB())
	if err != nil {
		log.Printf("error getting problematic messages: %v", err)
		return
	}

	for _, message := range problematicMessages {
		// Send a message to the chat with the problematic message details
		b.SendMessage(ctx, message.ChatID, "Problematic message detected: "+message.Content)
	}
}
