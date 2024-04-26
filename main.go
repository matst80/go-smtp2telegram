package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mnako/letters"

	"github.com/emersion/go-smtp"
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	Token  string `json:"token"`
	Domain string `json:"domain"`
	Listen string `json:"listen"`
	Users  []user `json:"users"`
}

type user struct {
	Email  string `json:"email"`
	ChatId int64  `json:"chatId"`
}

type backend struct {
	bot   *botapi.BotAPI
	users []user
}

type session struct {
	backend *backend
	from    string
	to      []int64
	email   letters.Email
}

func (bkd *backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &session{
		backend: bkd,
		to:      []int64{},
		email:   letters.Email{},
	}, nil
}

func (s *session) AuthPlain(username, password string) error {
	return nil
}

func (s *session) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *session) Rcpt(to string, opts *smtp.RcptOptions) error {
	for _, u := range s.backend.users {
		if to == u.Email {
			s.to = append(s.to, u.ChatId)
		}
	}

	return nil
}

func saveHtml(userId int64, email letters.Email) error {
	// Save email to a file
	err := os.MkdirAll(fmt.Sprintf("mail/%d", userId), 0755)
	if err != nil {
		return err
	}
	file, err := os.Create(fmt.Sprintf("mail/%d/%s.html", userId, email.Headers.MessageID))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(email.HTML)
	if err != nil {
		return err
	}
	for i, attachment := range email.AttachedFiles {
		file, err := os.Create(fmt.Sprintf("mail/%d/%s-%d", userId, email.Headers.MessageID, i))
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.Write(attachment.Data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *session) Data(r io.Reader) error {
	if len(s.to) > 0 {
		email, err := letters.ParseEmail(r)
		if err != nil {
			return err
		}
		s.email = email
		for _, userId := range s.to {
			go saveHtml(userId, s.email)
		}
	}
	return nil
}

func (s *session) Reset() {}

func textContent(s *session) string {
	return fmt.Sprintf("From: %s\nSubject: %s\n\n%s", s.from, s.email.Headers.Subject, s.email.Text)
}

func (s *session) Logout() error {
	hasSent := false

	for _, chatId := range s.to {

		content := textContent(s)

		// fmt.Sprintf("From: %s\nSubject: %s\n\n%s", s.from, s.email.Headers.Subject, s.email.Text)
		msg := botapi.NewMessage(chatId, content)

		s.backend.bot.Send(msg)
		log.Printf("Sent email to %d", chatId)

		hasSent = true

	}
	if !hasSent {
		log.Printf("Discarding email, no recipients found %s", s.from)
	}
	return nil
}

func getConfig() Config {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal("Error opening config.json")
	}
	defer configFile.Close()
	bytes, err := io.ReadAll(configFile)
	if err != nil {
		log.Fatal("Error reading config.json")
	}
	var config Config
	json.Unmarshal([]byte(bytes), &config)
	return config
}

func scanIds(bot *botapi.BotAPI) {
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

func main() {
	config := getConfig()

	bot, err := botapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic(err)
	}

	s := smtp.NewServer(&backend{
		bot:   bot,
		users: config.Users,
	})

	log.Printf("Authorized on account %s", bot.Self.UserName)

	s.Addr = config.Listen
	s.Domain = config.Domain
	s.AllowInsecureAuth = true

	if os.Getenv("DEBUG") == "true" {
		bot.Debug = true
		s.Debug = os.Stdout
	}

	go scanIds(bot)

	log.Println("Starting SMTP server at", s.Addr)
	log.Fatal(s.ListenAndServe())
}
