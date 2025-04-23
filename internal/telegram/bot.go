package telegram

import (
	"context"
	"database/sql"
	"log"

	"github.com/go-telegram/bot"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/config"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/llm"
)

type Bot struct {
	client *bot.Bot
	db     *sql.DB
	Nrapp  *newrelic.Application
}

func NewBot(ctx context.Context, config *config.Config, db *sql.DB, nrApp *newrelic.Application) (Bot, error) {
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
		Nrapp:  nrApp,
	}, nil
}

func (b Bot) Start(ctx context.Context) error {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	txn := b.Nrapp.StartTransaction("bot-startup")
	defer txn.End()

	ctxWithTxn := newrelic.NewContext(ctx, txn)

	llmClient, err := llm.NewLLMClient(config)
	if err != nil {
		log.Printf("error creating LLM client: %v", err)

		return err
	}

	llm.LoadPrompts(config.PromptsPath)

	summaryPrompt, err := llm.GetPrompt("summary", config.Language)
	if err != nil {
		log.Fatalf("Failed to get prompt: %v", err)
	}

	problematicPrompt, err := llm.GetPrompt("problematic", config.Language)
	if err != nil {
		log.Fatalf("Failed to get prompt: %v", err)
	}

	valueAssessmentPrompt, err := llm.GetPrompt("value_assessment", config.Language)
	if err != nil {
		log.Fatalf("Failed to get prompt: %v", err)
	}

	// Register commands
	b.client.RegisterHandler(bot.HandlerTypeMessageText, "/tldr", bot.MatchTypePrefix, tldrHandler(b.Nrapp, llmClient, summaryPrompt))
	b.client.RegisterHandler(bot.HandlerTypeMessageText, "/problematic", bot.MatchTypePrefix, problematicSpeechHandler(b.Nrapp, llmClient, problematicPrompt))
	b.client.RegisterHandler(bot.HandlerTypeMessageText, "/valeapena", bot.MatchTypePrefix, valueAssessment(b.Nrapp, llmClient, valueAssessmentPrompt))

	b.client.Start(ctxWithTxn)
	return nil
}
