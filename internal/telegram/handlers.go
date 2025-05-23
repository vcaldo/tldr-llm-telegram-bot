package telegram

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/constants"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/db"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/llm"
)

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	txn := newrelic.FromContext(ctx)

	segDB := txn.StartSegment("db:log-message")
	var err error
	switch {
	case update.Message != nil && update.Message.Text != "":
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeText, update, sanitizeHTMLContent(update.Message.Text))
	case update.Message != nil && update.Message.Voice != nil:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeVoice, update, update.Message.Voice)
	case update.Message != nil && update.Message.Photo != nil:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypePhoto, update, update.Message.Photo)
	case update.Message != nil && update.Message.Animation != nil:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeAnimation, update, update.Message.Animation)
	case update.Message != nil && update.Message.Sticker != nil:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeSticker, update, update.Message.Sticker)
	case update.Message != nil && update.Message.Video != nil:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeVideo, update, update.Message.Video)
	case update.Message != nil && update.Message.VideoNote != nil:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeVideoNote, update, update.Message.VideoNote)
	default:
		err = db.LogMessage(ctx, db.GetDB(), constants.MessageTypeGeneric, update, update.Message)
	}
	segDB.End()

	if err != nil {
		txn.NoticeError(err)
		log.Printf("error logging message: %v", err)
	}

}

func tldrHandler(nrApp *newrelic.Application, llmClient llm.LLMClient, summaryPrompt string) func(ctx context.Context, b *bot.Bot, update *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message.Text == "" || update.Message.ReplyToMessage == nil {
			return
		}
		txn := nrApp.StartTransaction("handler:tldr")
		defer txn.End()

		txn.AddAttribute("chatID", update.Message.Chat.ID)
		txn.AddAttribute("userID", update.Message.From.ID)

		ctxWithTxn := newrelic.NewContext(ctx, txn)

		firstMessageTimestamp, err := getMessageTimestamp(db.GetDB(), int64(update.Message.ReplyToMessage.ID), update.Message.Chat.ID)
		if err != nil {
			log.Printf("error getting message timestamp: %v", err)
			return
		}

		messages, err := db.FetchMessagesSince(ctxWithTxn, db.GetDB(), update.Message.Chat.ID, int64(update.Message.ReplyToMessage.ID), firstMessageTimestamp, 720*time.Minute)
		if err != nil {
			log.Printf("error fetching messages: %v", err)
			return
		}

		if len(messages) == 0 {
			log.Printf("no messages found since %v", firstMessageTimestamp)
			return
		}

		txn.AddAttribute("messagesCount", len(messages))

		formattedMessages := formatTextMessages(messages)
		prompt := fmt.Sprintf("%s %s", summaryPrompt, formattedMessages)

		txn.AddAttribute("promptLen", len(prompt))

		summary, err := llmClient.AnalyzePrompt(nrApp, prompt)
		if err != nil {
			log.Printf("error generating summary: %v", err)
			return
		}

		log.Printf("generated summary: %s", summary)

		SendLongMessage(ctxWithTxn, nrApp, b, update.Message.Chat.ID, summary)
	}
}

func problematicSpeechHandler(nrApp *newrelic.Application, llmClient llm.LLMClient, problematicPrompt string) func(ctx context.Context, b *bot.Bot, update *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message.Text == "" {
			return
		}

		txn := nrApp.StartTransaction("handler:problematic-speech-moderation")
		defer txn.End()

		txn.AddAttribute("chatID", update.Message.Chat.ID)
		txn.AddAttribute("userID", update.Message.From.ID)

		ctxWithTxn := newrelic.NewContext(ctx, txn)

		messages, err := db.FetchUnmoderatedMessages(ctxWithTxn, db.GetDB(), update.Message.Chat.ID)
		if err != nil {
			log.Printf("error fetching messages: %v", err)
			txn.NoticeError(err)
			return
		}

		if len(messages) == 0 {
			log.Printf("no unmoderated messages found")
			return
		}

		txn.AddAttribute("messagesCount", len(messages))

		formattedMessages := formatTextMessages(messages)

		prompt := fmt.Sprintf("%s %s", problematicPrompt, formattedMessages)

		txn.AddAttribute("promptLen", len(prompt))

		problematicContent, err := llmClient.AnalyzePrompt(nrApp, prompt)
		if err != nil {
			log.Printf("error generating problematic content: %v", err)
			txn.NoticeError(err)
			return
		}

		if len(problematicContent) > 4 {
			SendLongMessage(ctxWithTxn, nrApp, b, update.Message.Chat.ID, problematicContent)
		}

		if err := db.SetMessagesModerated(ctxWithTxn, db.GetDB(), messages); err != nil {
			log.Printf("error setting messages as moderated: %v", err)
			txn.NoticeError(err)
		}
	}
}

func valueAssessment(nrApp *newrelic.Application, llmClient llm.LLMClient, valueAssessmentPrompt string) func(ctx context.Context, b *bot.Bot, update *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		txn := nrApp.StartTransaction("handler:value-assessment")
		defer txn.End()

		txn.AddAttribute("chatID", update.Message.Chat.ID)
		txn.AddAttribute("userID", update.Message.From.ID)

		ctxWithTxn := newrelic.NewContext(ctx, txn)

		err := db.LogMessage(ctxWithTxn, db.GetDB(), constants.MessageTypeText, update, update.Message.Text)
		if err != nil {
			txn.NoticeError(err)
			log.Printf("error logging message: %v", err)
		}

		if update.Message.Text == "" {
			return
		}

		log.Printf("Received message: %s", update.Message.Text)

		prompt := fmt.Sprintf("%s %s", valueAssessmentPrompt, update.Message.Text)

		txn.AddAttribute("promptLen", len(prompt))

		valueAssessmentContent, err := llmClient.AnalyzePrompt(nrApp, prompt)
		if err != nil {
			txn.NoticeError(err)
			log.Printf("error generating value assessment content: %v", err)
			return
		}

		SendLongMessage(ctxWithTxn, nrApp, b, update.Message.Chat.ID, valueAssessmentContent)
	}
}

func sportsScheduleHandler(nrApp *newrelic.Application, llmClient llm.LLMClient, sportsSchedulePrompt string) func(ctx context.Context, b *bot.Bot, update *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		txn := nrApp.StartTransaction("handler:sports-schedule")
		defer txn.End()

		txn.AddAttribute("chatID", update.Message.Chat.ID)
		txn.AddAttribute("userID", update.Message.From.ID)

		ctxWithTxn := newrelic.NewContext(ctx, txn)

		err := db.LogMessage(ctxWithTxn, db.GetDB(), constants.MessageTypeText, update, update.Message.Text)
		if err != nil {
			txn.NoticeError(err)
			log.Printf("error logging message: %v", err)
		}

		if update.Message.Text == "" {
			return
		}

		log.Printf("Received message: %s", update.Message.Text)

		prompt := fmt.Sprintf("%s %s", sportsSchedulePrompt, update.Message.Text)

		txn.AddAttribute("promptLen", len(prompt))

		sportsScheduleContent, err := llmClient.AnalyzePrompt(nrApp, prompt)
		if err != nil {
			txn.NoticeError(err)
			log.Printf("error generating sports schedule content: %v", err)
			return
		}

		SendLongMessage(ctxWithTxn, nrApp, b, update.Message.Chat.ID, sportsScheduleContent)
	}
}
