package telegram

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/db"
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
		log.Printf("Error logging message: %v", err)
	}
}

func fooHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Caught command `/foo`",
		ParseMode: models.ParseModeMarkdown,
	})
}

func barHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Caught command `/bar`",
		ParseMode: models.ParseModeMarkdown,
	})
}
