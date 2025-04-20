package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/vcaldo/tldr-llm-telegram-bot/internal/config"
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

	bot, err := telegram.NewBot(ctx, config)
	if err != nil {
		log.Fatalf("error initializing bot: %v", err)
		return
	}

	log.Println("bot started")
	bot.Start(ctx)
	defer log.Println("bot stopped")
}
