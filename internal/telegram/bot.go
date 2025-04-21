package telegram

import (
	"context"
	"database/sql"
	"log"

	"github.com/go-telegram/bot"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/config"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/llm"
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
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	llmClient, err := llm.NewLLMClient(config)
	if err != nil {
		log.Printf("error creating LLM client: %v", err)

		return err
	}

	llm.LoadPrompts(config.PromptsPath)
	if err != nil {
		log.Printf("error loading prompts: %v", err)
	}

	summaryPrompt, err := llm.GetPrompt("summary", config.Language)
	if err != nil {
		log.Fatalf("Failed to get prompt: %v", err)
	}

	problematicPrompt, err := llm.GetPrompt("problematic", config.Language)
	if err != nil {
		log.Fatalf("Failed to get prompt: %v", err)
	}

	// Register commands
	b.client.RegisterHandler(bot.HandlerTypeMessageText, "/tldr", bot.MatchTypePrefix, tldrHandler(llmClient, summaryPrompt))
	b.client.RegisterHandler(bot.HandlerTypeMessageText, "/problematic", bot.MatchTypePrefix, problematicSpeechHandler(llmClient, problematicPrompt))

	b.client.Start(ctx)
	return nil
}
