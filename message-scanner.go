package main

import (
	"log"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ScanIds(bot *botapi.BotAPI) {
	u := botapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %d", update.Message.From.UserName, update.Message.Chat.ID)

			msg := botapi.NewMessage(update.Message.Chat.ID, "Mail bot got your message, thanks")
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}
