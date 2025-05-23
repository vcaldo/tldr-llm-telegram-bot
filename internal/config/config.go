package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TelegramBotToken   string `envconfig:"TELEGRAM_BOT_TOKEN" required:"true"`
	Language           string `envconfig:"LANGUAGE" default:"en"`
	DatabaseURL        string `envconfig:"DATABASE_URL" required:"true"`
	LLMProvider        string `envconfig:"LLM_PROVIDER" required:"true"`
	OllamaApiUrl       string `envconfig:"OLLAMA_API_URL"`
	OllamaModel        string `envconfig:"OLLAMA_MODEL"`
	GeminiApiUrl       string `envconfig:"GEMINI_API_URL"`
	GeminiModel        string `envconfig:"GEMINI_MODEL"`
	GeminiApiKey       string `envconfig:"GEMINI_API_KEY"`
	PromptsPath        string `envconfig:"PROMPTS_PATH"`
	NewRelicLicenseKey string `envconfig:"NEW_RELIC_LICENSE_KEY"`
	NewRelicAppName    string `envconfig:"NEW_RELIC_APP_NAME"`
}

func LoadConfig() (*Config, error) {

	err := godotenv.Load(os.Getenv("ENV_FILE_PATH"))
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
