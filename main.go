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
	client  net.Addr
	backend *backend
	from    string
	to      []int64
	email   letters.Email
}

var addr map[string]int

func (bkd *backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	client := c.Conn().RemoteAddr()

	ip := getIpFromAddr(client)
	err := AllowedAddress(ip)
	if err != nil {
		log.Printf("Blocked address %s", client)
		return nil, err
	}
	if addr == nil && ip != "" {
		addr[ip]++
		log.Print(addr)
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
	isSpam := IsSpamContent(s.email.HTML) || IsSpamContent(s.email.Text)
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
		log.Printf("Discarding email, no recipients found %s %s", s.from, s.client)
	}
	return nil
}

func main() {
	config := GetConfig()

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

	go ScanIds(bot)

	log.Println("Starting SMTP server at", s.Addr)
	log.Fatal(s.ListenAndServe())
}
