package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/mnako/letters"

	"github.com/emersion/go-smtp"
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type user struct {
	Email  string `json:"email"`
	ChatId int64  `json:"chatId"`
}

type backend struct {
	bot   *botapi.BotAPI
	spam  *spam
	users []user
}

type session struct {
	client  net.Addr
	backend *backend
	from    string
	to      []int64
	email   letters.Email
}

func (bkd *backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	client := c.Conn().RemoteAddr()

	ip := getIpFromAddr(client)
	err := bkd.spam.AllowedAddress(ip)
	if err != nil {
		log.Printf("Blocked address %s", client)
		return nil, err
	}

	return &session{
		client:  client,
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
	isSpam := s.backend.spam.IsSpamContent(s.email.HTML) || s.backend.spam.IsSpamContent(s.email.Text)
	if isSpam {
		log.Printf("Discarding email, spam detected %s %s", s.from, s.client)
		return nil
	}
	for _, chatId := range s.to {

		content := textContent(s)

		msg := botapi.NewMessage(chatId, content)

		s.backend.bot.Send(msg)
		log.Printf("Sent email to %d", chatId)

		hasSent = true

	}
	if !hasSent {
		log.Printf("Discarding email, no recipient, from: %s (%s)", s.from, s.client)
	}
	return nil
}

type commandHandler struct {
	spam   *spam
	config *Config
	bot    *botapi.BotAPI
}

func (cmd *commandHandler) OnMessage(msg *botapi.Message) error {
	if msg.IsCommand() {
		command := msg.Command()
		if command == "ips" {
			m := botapi.NewMessage(msg.Chat.ID, "Updating blocked ips...")
			if _, err := cmd.bot.Send(m); err != nil {
				return err
			}
			if err := cmd.spam.UpdateBlockedIpsFromUrl(cmd.config.BlockedIpUrl); err != nil {
				return err
			}
			//m = botapi.NewMessage(msg.Chat.ID, "Updated blocked ips")
			m.Text = "Updated blocked ips"
			if _, err := cmd.bot.Send(m); err != nil {
				return err
			}
		} else if command == "words" {
			m := botapi.NewMessage(msg.Chat.ID, "Updating warning words...")
			if _, err := cmd.bot.Send(m); err != nil {
				return err
			}
			if err := cmd.spam.UpdateWarningWordsFromUrl(cmd.config.WarningWordsUrl); err != nil {
				return err
			}
			m.Text = "Updated warning words"
			//m = botapi.NewMessage(msg.Chat.ID, "Updated warning words")
			if _, err := cmd.bot.Send(m); err != nil {
				return err
			}
		} else if command == "start" {
			m := botapi.NewMessage(msg.Chat.ID, "Hello! I got your message, id logged on the server")
			log.Printf("[%s %s] %d", msg.From.FirstName, msg.From.LastName, msg.Chat.ID)
			if _, err := cmd.bot.Send(m); err != nil {
				return err
			}
		} else {
			log.Printf("Unknown command %s", command)
		}
	} else {
		log.Printf("Unknown message %s", msg.Text)
	}
	return nil
}

func main() {
	config := GetConfig()

	bot, err := botapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic(err)
	}

	spm := &spam{
		spamWords:    config.StopWords,
		warningWords: config.WarningWords,
		blockedIps:   config.BlockedIps,
	}
	if config.BlockedIpUrl != "" {
		err := spm.UpdateBlockedIpsFromUrl(config.BlockedIpUrl)
		if err != nil {
			log.Fatal("Error updating blocked ips", err)
		}
	}
	if config.WarningWordsUrl != "" {
		err := spm.UpdateWarningWordsFromUrl(config.WarningWordsUrl)
		if err != nil {
			log.Fatal("Error updating warning words", err)
		}
	}

	s := smtp.NewServer(&backend{
		bot:   bot,
		users: config.Users,
		spam:  spm,
	})

	log.Printf("Authorized on account %s", bot.Self.UserName)

	s.Addr = config.Listen
	s.Domain = config.Domain
	s.AllowInsecureAuth = true

	if os.Getenv("DEBUG") == "true" {
		bot.Debug = true
		s.Debug = os.Stdout
	}

	go ListenForMessages(bot, &commandHandler{
		spam:   spm,
		config: &config,
		bot:    bot,
	})

	log.Println("Starting SMTP server at", s.Addr)
	log.Fatal(s.ListenAndServe())
}
