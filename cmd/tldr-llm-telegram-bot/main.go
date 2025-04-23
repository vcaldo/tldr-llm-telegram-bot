package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/config"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/db"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/telegram"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
		return
	}

	nrApp, err := newrelic.NewApplication(
		newrelic.ConfigAppName(config.NewRelicAppName),
		newrelic.ConfigLicense(config.NewRelicLicenseKey),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		log.Fatalf("warning: failed initializing New Relic: %v", err)
	} else {
		defer nrApp.Shutdown(10 * time.Second)
		log.Println("New Relic agent initialized")
	}
	db.InitDB(ctx, config, nrApp)
	defer db.CloseDB()
	log.Println("database initialized")

	bot, err := telegram.NewBot(ctx, config, db.GetDB(), nrApp)
	if err != nil {
		log.Fatalf("error initializing bot: %v", err)
		return
	}

	log.Println("bot started")
	bot.Start(ctx)
	defer log.Println("bot stopped")
}
