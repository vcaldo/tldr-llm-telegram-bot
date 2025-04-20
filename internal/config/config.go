package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TelegramBotToken string `envconfig:"TELEGRAM_BOT_TOKEN" required:"true"`
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("error loading .env file: %v", err)
		return nil, err
	}

	var config Config
	err = envconfig.Process("", &config)
	if err != nil {
		log.Printf("error processing environment variables: %v", err)
		return nil, err
	}

	return &config, nil
}
