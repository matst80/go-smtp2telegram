package server

import (
	"log"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandHandler interface {
	OnMessage(msg *botapi.Message) error
}

func ListenForMessages(bot *botapi.BotAPI, cmd CommandHandler) {
	u := botapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Recovered from panic in OnMessage: %v", r)
					}
				}()
				err := cmd.OnMessage(update.Message)
				if err != nil {
					log.Printf("Error: %v", err)
				}
			}()
		}
	}
}
