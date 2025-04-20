package telegram

import (
	"context"
	"database/sql"
	"log"

	"github.com/go-telegram/bot"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/config"
)

type Bot struct {
	client *bot.Bot
	db     *sql.DB
}

func NewBot(ctx context.Context, config *config.Config, db *sql.DB) (Bot, error) {
	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}

	client, err := bot.New(config.TelegramBotToken, opts...)
	if err != nil {
		log.Fatalf("error creating bot: %v", err)
		return Bot{}, err
	}

	return Bot{
		client: client,
		db:     db,
	}, nil
}

func (b Bot) Start(ctx context.Context) error {
	b.client.RegisterHandler(bot.HandlerTypeMessageText, "foo", bot.MatchTypeCommand, fooHandler)
	b.client.RegisterHandler(bot.HandlerTypeMessageText, "bar", bot.MatchTypeCommandStartOnly, barHandler)

	b.client.Start(ctx)
	return nil
}
