package telegram

import (
	"context"
	"database/sql"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
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
	instrumentedDefaultHandler := func(ctx context.Context, b *bot.Bot, update *models.Update) {
		txn := nrApp.StartTransaction("handler:default")
		defer txn.End()

		if update.Message != nil {
			txn.AddAttribute("chatID", update.Message.Chat.ID)
			txn.AddAttribute("userID", update.Message.From.ID)
		}

		ctxWithTxn := newrelic.NewContext(ctx, txn)

		defaultHandler(ctxWithTxn, b, update)
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(instrumentedDefaultHandler),
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

	prompts, err := LoadPrompts(&llmClient, config)
	if err != nil {
		log.Fatalf("failed to load prompts: %v", err)
	}

	summaryPrompt := prompts[0]
	problematicPrompt := prompts[1]
	valueAssessmentPrompt := prompts[2]
	sportsSchedulePrompt := prompts[3]

	// Register commands
	b.client.RegisterHandler(bot.HandlerTypeMessageText, "/tldr", bot.MatchTypePrefix, tldrHandler(b.Nrapp, llmClient, summaryPrompt))
	b.client.RegisterHandler(bot.HandlerTypeMessageText, "/problematic", bot.MatchTypePrefix, problematicSpeechHandler(b.Nrapp, llmClient, problematicPrompt))
	b.client.RegisterHandler(bot.HandlerTypeMessageText, "/valeapena", bot.MatchTypePrefix, valueAssessment(b.Nrapp, llmClient, valueAssessmentPrompt))
	b.client.RegisterHandler(bot.HandlerTypeMessageText, "/futebol", bot.MatchTypePrefix, sportsScheduleHandler(b.Nrapp, llmClient, sportsSchedulePrompt))

	b.client.Start(ctxWithTxn)
	return nil
}
