# tldr-llm-telegram-bot
## Overview

`tldr-llm-telegram-bot` is a Telegram bot written in Go that leverages large language models to provide the following features:

- **Summarize Chats:** Quickly generate concise summaries of chat conversations.
- **Moderate Problematic Content:** Automatically detect and flag inappropriate or harmful messages.
- **Answer Value Comparison Questions:** Respond to questions that require comparing values or making judgments.

## Customizable Prompts


## Using External Prompt Files with Docker

You can provide a custom prompts file from outside the container when running the bot with Docker. This allows you to override the embedded prompts without modifying the code.

To do this, mount your external prompts directory as a Docker volume and set the `PROMPTS_PATH` environment variable to the mounted path. For example:

```yaml
    volumes:
      - .prompts.toml:/app/prompts.toml
      - .env:/app/.env
```

This setup ensures the bot uses your external prompt definitions instead of the defaults bundled in the container.

## Model Support

The bot supports two types of language models:
- **Gemini API:** Use Google's Gemini API for cloud-based inference with configurable parameters (temperature, top_k, top_p, max tokens).
- **Ollama Local Models:** Run models locally using Ollama for privacy and control.

The model implementation includes timeout handling and error recovery for reliable operation.

## Configuration

All required settings must be specified in a `.env` file at the root of the project:

```
# Required settings
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
DATABASE_URL=your_database_url
LLM_PROVIDER=gemini|ollama  # Choose one provider

# For Ollama (required if LLM_PROVIDER=ollama)
OLLAMA_API_URL=http://localhost:11434/api/generate
OLLAMA_MODEL=llama2

# For Gemini (required if LLM_PROVIDER=gemini)
GEMINI_API_URL=https://generativelanguage.googleapis.com
GEMINI_MODEL=gemini-pro
GEMINI_API_KEY=your_gemini_api_key

# Optional settings
LANGUAGE=en  # Default language
PROMPTS_PATH=path/to/prompts/directory
NEW_RELIC_LICENSE_KEY=your_new_relic_license_key
NEW_RELIC_APP_NAME=your_new_relic_app_name
```

## Database

The bot stores message data in a SQL database, which is configured using the `DATABASE_URL` environment variable. The database is used to maintain context between conversations and analyze message history.

## New Relic APM Integration

The codebase is fully instrumented with New Relic Application Performance Monitoring (APM) for observability. This provides:

- Transaction monitoring for each bot command
- Detailed metrics for model calls including prompt length, response time, and completion status
- Error tracking and alerting
- Customized attributes for better filtering and analysis

The New Relic integration is optional and can be enabled by providing license key and app name in the configuration.

## Usage Instructions

After starting the bot:

1. Find your bot on Telegram by its username
2. Start a conversation with `/start`
3. Available commands:
   - `/tldr` - Summarize recent messages in a chat
   - `/problematic` - Check if content violates community guidelines
   - `/valeapena - Ask a comparison question

## Running with Docker

To run the bot using Docker:

1. Ensure your `.env` file is properly configured.
2. Build and run the Docker container:
    ```sh
    docker compose build
    docker compose up -d
    ```

## Development Setup

To set up a local development environment:

1. Clone the repository
2. Install Go (1.24 or newer recommended)
3. Copy `.env.example` to `.env` and fill in required values
4. Run the application in development mode:
   ```sh
   go run ./cmd/tldr-llm-telegram-bot/main.go
   ```

## Requirements

- Go (for local development)
- Docker (for containerized deployment)
- PostgreSQL or compatible SQL database
- Access to either Gemini API or a local Ollama model

## License

See [LICENSE](./LICENSE) for details.
