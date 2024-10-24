package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"knatofs.se/crapmail/pkg/client"
	"knatofs.se/crapmail/pkg/server"
)

type commandHandler struct {
	smtpClient *client.SmtpClient
	spam       *server.Spam
	config     *server.Config
	bot        *botapi.BotAPI
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

func getValidEmailAddresses(input string) []string {
	emails := strings.Split(input, " ")
	var validEmails []string
	for _, email := range emails {
		if strings.Contains(email, "@") {
			validEmails = append(validEmails, email)
		}
	}
	return validEmails
}

func (cmd *commandHandler) findUser(chatId int64) *server.User {
	for _, user := range cmd.config.Users {
		if user.ChatId == chatId {
			return &user
		}
	}
	return nil
}

func (cmd *commandHandler) OnMessage(msg *botapi.Message) error {
	if msg.IsCommand() {
		switch command := msg.Command(); command {
		case "send":
			log.Println(msg.CommandArguments())
			user := cmd.findUser(msg.Chat.ID)
			if user == nil {
				return fmt.Errorf("User not found")
			}
			messge, err := client.ParseMessage(msg.CommandArguments(), user.Email, "Chat reply")
			if err != nil {
				return err
			}
			err = cmd.smtpClient.Send(*messge)

			return err

		case "config":
			sendConfig(cmd.bot, msg.Chat.ID, cmd.config)
		case "users":
			sendConfig(cmd.bot, msg.Chat.ID, cmd.config.Users)
		case "add":
			log.Println(msg.CommandArguments())
			p := getValidEmailAddresses(msg.CommandArguments())
			for _, email := range p {
				log.Printf("Adding %s", email)
				cmd.config.Users = append(cmd.config.Users, server.User{ChatId: msg.Chat.ID, Email: email})
			}
			sendConfig(cmd.bot, msg.Chat.ID, cmd.config.Users)
		case "block":
			ip := msg.CommandArguments()
			if ip != "" {
				cmd.spam.BlockedIps = append(cmd.spam.BlockedIps, ip)
				m := botapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Blocked ip %s", ip))
				_, err := cmd.bot.Send(m)
				return err
			}
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
