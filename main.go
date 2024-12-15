package main

import (
	"log"
	"os"

	"github.com/emersion/go-smtp"
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"knatofs.se/crapmail/pkg/client"
	"knatofs.se/crapmail/pkg/server"
)

func main() {
	config, err := server.GetConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	bot, err := botapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic(err)
	}

	h := &server.SimpleHash{
		Salt: config.HashSalt,
	}

	dkimBytes, err := client.ReadFile("dkim.pem")
	if err != nil {
		log.Panic(err)
	}

	smtpClient, err := client.MakeSmtpClient(config.Domain, config.DkimSelector, dkimBytes)
	if err != nil {
		log.Fatal("Error creating smtp client", err)
	}
	spm := &server.Spam{
		SpamWords:    config.StopWords,
		WarningWords: []string{},
		BlockedIps:   []string{},
		MaxSpamCount: 5,
	}
	go func() {
		if config.BlockedIpUrl != "" {
			err := spm.UpdateBlockedIpsFromUrl(config.BlockedIpUrl)
			if err != nil {
				log.Println("Error updating blocked ips", err)
			}
		}
		if config.WarningWordsUrl != "" {
			err := spm.UpdateWarningWordsFromUrl(config.WarningWordsUrl)
			if err != nil {
				log.Println("Error updating warning words", err)
			}
		}
	}()

	var spam server.SpamClassification = nil
	if config.OpenAi.ApiKey != "" {
		spam = server.MakeAiClassifier(&config.OpenAi)
	}

	s := smtp.NewServer(&server.Backend{
		SpamClassifier: spam,
		HashGenerator:  h,
		Bot:            bot,
		Config:         config,
		SpamChecker:    spm,
	})
	go server.WebServer(h)

	log.Printf("Bot authorized [%s]", bot.Self.UserName)

	s.Addr = config.Listen
	s.Domain = config.Domain
	s.AllowInsecureAuth = true

	if os.Getenv("DEBUG") == "true" {
		bot.Debug = true
		spm.Debug = true
		s.Debug = os.Stdout
	}

	go server.ListenForMessages(bot, &commandHandler{
		smtpClient: smtpClient,
		spam:       spm,
		config:     config,
		bot:        bot,
	})

	log.Printf("Starting SMTP server at %s", s.Addr)
	log.Fatal(s.ListenAndServe())
}
