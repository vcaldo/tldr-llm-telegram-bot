package telegram

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/db"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/llm"
)

const (
	MessageTypeText      = "text"
	MessageTypeVoice     = "voice"
	MessageTypePhoto     = "photo"
	MessageTypeAnimation = "animation"
	MessageTypeSticker   = "sticker"
	MessageTypeVideo     = "video"
	MessageTypeVideoNote = "video_note"
	MessageTypeGeneric   = "generic"
)

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var err error
	switch {
	case update.Message != nil && update.Message.Text != "":
		err = db.LogMessage(ctx, db.GetDB(), MessageTypeText, update, update.Message.Text)
	case update.Message != nil && update.Message.Voice != nil:
		err = db.LogMessage(ctx, db.GetDB(), MessageTypeVoice, update, update.Message.Voice)
	case update.Message != nil && update.Message.Photo != nil:
		err = db.LogMessage(ctx, db.GetDB(), MessageTypePhoto, update, update.Message.Photo)
	case update.Message != nil && update.Message.Animation != nil:
		err = db.LogMessage(ctx, db.GetDB(), MessageTypeAnimation, update, update.Message.Animation)
	case update.Message != nil && update.Message.Sticker != nil:
		err = db.LogMessage(ctx, db.GetDB(), MessageTypeSticker, update, update.Message.Sticker)
	case update.Message != nil && update.Message.Video != nil:
		err = db.LogMessage(ctx, db.GetDB(), MessageTypeVideo, update, update.Message.Video)
	case update.Message != nil && update.Message.VideoNote != nil:
		err = db.LogMessage(ctx, db.GetDB(), MessageTypeVideoNote, update, update.Message.VideoNote)
	default:
		err = db.LogMessage(ctx, db.GetDB(), MessageTypeGeneric, update, update.Message)
	}
	if err != nil {
		log.Printf("error logging message: %v", err)
	}
}

func tldrHandler(llmClient llm.LLMClient) func(ctx context.Context, b *bot.Bot, update *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		// implement prompt + chat
		prompt := update.Message.Text
		summary, err := llmClient.AnalyzePrompt(prompt)
		if err != nil {
			log.Printf("error generating summary: %v", err)
			return
		}

		log.Printf("generated summary: %s", summary)
		// Send the summary back to the user
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      summary,
			ParseMode: "html",
		}); err != nil {
			log.Printf("error sending message: %v", err)
		}
	}
}

func problematicSpeechHandler(llmClient llm.LLMClient) func(ctx context.Context, b *bot.Bot, update *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		// implement prompt + chat
		prompt := update.Message.Text
		problematicContent, err := llmClient.AnalyzePrompt(prompt)

		if err != nil {
			log.Printf("error generating summary: %v", err)
			return
		}

		log.Printf("generated summary: %s", problematicContent)
		// Send the summary back to the user
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      problematicContent,
			ParseMode: "html",
		}); err != nil {
			log.Printf("error sending message: %v", err)
		}
	}
}
