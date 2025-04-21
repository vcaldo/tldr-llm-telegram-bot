package db

import (
	"fmt"

	"github.com/go-telegram/bot/models"
)

func getDisplayName(update *models.Update) string {
	switch {
	case update.Message.From.FirstName != "" && update.Message.From.LastName != "":
		return fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName)
	case update.Message.From.FirstName != "":
		return update.Message.From.FirstName
	case update.Message.From.LastName != "":
		return update.Message.From.LastName
	default:
		return update.Message.From.Username
	}
}
