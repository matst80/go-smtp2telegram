package main

import (
	"encoding/json"
	"fmt"
	"log"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type commandHandler struct {
	spam   *Spam
	config *Config
	bot    *botapi.BotAPI
}

func getMessageJson(data interface{}) string {
	configString, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "<i>Error serializing data</i>"
	}
	return fmt.Sprintf("<code><pre>%s</pre></code>", configString)
}

func sendConfig(bot *botapi.BotAPI, chatId int64, data interface{}) error {
	m := botapi.NewMessage(chatId, getMessageJson(data))
	m.ParseMode = "HTML"
	_, err := bot.Send(m)
	return err
}

func (cmd *commandHandler) OnMessage(msg *botapi.Message) error {
	if msg.IsCommand() {
		switch command := msg.Command(); command {
		case "config":
			sendConfig(cmd.bot, msg.Chat.ID, cmd.config)
		case "users":
			sendConfig(cmd.bot, msg.Chat.ID, cmd.config.Users)
		case "add":
			log.Println(msg.CommandArguments())
			p := getValidEmailAddresses(msg.CommandArguments())
			for _, email := range p {
				log.Printf("Adding %s", email)
				cmd.config.Users = append(cmd.config.Users, User{ChatId: msg.Chat.ID, Email: email})
			}
			sendConfig(cmd.bot, msg.Chat.ID, cmd.config.Users)

		case "ips":
			m := botapi.NewMessage(msg.Chat.ID, "Updating blocked ips...")
			cmd.bot.Send(m)
			if err := cmd.spam.UpdateBlockedIpsFromUrl(cmd.config.BlockedIpUrl); err != nil {
				return err
			}
			m.Text = "Updated blocked ips"
			_, err := cmd.bot.Send(m)
			return err
		case "words":
			m := botapi.NewMessage(msg.Chat.ID, "Updating warning words...")
			cmd.bot.Send(m)

			if err := cmd.spam.UpdateWarningWordsFromUrl(cmd.config.WarningWordsUrl); err != nil {
				return err
			}

			m.Text = "Updated warning words"
			_, err := cmd.bot.Send(m)
			return err
		case "start":
			m := botapi.NewMessage(msg.Chat.ID, "Hello! I got your message, id logged on the server")
			log.Printf("[%s %s] %d", msg.From.FirstName, msg.From.LastName, msg.Chat.ID)
			_, err := cmd.bot.Send(m)
			return err
		default:
			log.Printf("Unknown command %s", command)
		}
	} else {
		log.Printf("Unknown message %s", msg.Text)
	}
	return nil
}
