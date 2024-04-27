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

type backend struct {
	hash   *hash
	bot    *botapi.BotAPI
	config *Config
	spam   *Spam
}

type session struct {
	client  net.Addr
	backend *backend
	from    string
	to      []int64
	email   letters.Email
	mailId  string
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
	for _, u := range s.backend.config.Users {
		if to == u.Email {
			s.to = append(s.to, u.ChatId)
		}
	}

	return nil
}

func saveHtml(emailId string, userId int64, email letters.Email) error {
	// Save email to a file
	err := os.MkdirAll(fmt.Sprintf("mail/%d", userId), 0755)
	if err != nil {
		return err
	}
	file, err := os.Create(fmt.Sprintf("mail/%d/%s.html", userId, emailId))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(email.HTML)
	if err != nil {
		return err
	}
	for i, attachment := range email.AttachedFiles {
		file, err := os.Create(fmt.Sprintf("mail/%d/%s-%d", userId, emailId, i))
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
		if email.HTML != "" {
			mailId := getEmailFileName(email.Headers)

			s.mailId = mailId
			for _, userId := range s.to {
				go saveHtml(mailId, userId, s.email)
			}
		}
	}
	return nil
}

func (s *session) Reset() {}

func textContent(s *session, chatId int64) string {
	extra := ""
	if s.mailId != "" && s.email.HTML != "" {
		hashQuery := s.backend.hash.createSimpleHash(fmt.Sprintf("%d%s", chatId, s.mailId))
		extra = fmt.Sprintf("\n\n%s/mail/%d/%s.html?hash=%s", s.backend.config.BaseUrl, chatId, s.mailId, hashQuery)
	}

	return fmt.Sprintf("From: %s\nSubject: %s\n\n%s%s", s.from, s.email.Headers.Subject, s.email.Text, extra)
}

func (s *session) Logout() error {
	hasSent := false
	isSpam := s.backend.spam.IsSpamHtml(s.email.HTML) || s.backend.spam.IsSpamContent(s.email.Text)
	if isSpam {
		ip := getIpFromAddr(s.client)
		s.backend.spam.LogSpamIp(ip)
		log.Printf("Spam detected (%s) [%s]", s.from, ip)
		return nil
	}
	for _, chatId := range s.to {

		content := textContent(s, chatId)

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

func main() {
	config, err := GetConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	bot, err := botapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic(err)
	}

	h := &hash{
		salt: config.HashSalt,
	}

	spm := &Spam{
		SpamWords:    config.StopWords,
		WarningWords: []string{},
		BlockedIps:   []string{},
		MaxSpamCount: 5,
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
		hash:   h,
		bot:    bot,
		config: &config,
		spam:   spm,
	})
	go WebServer(h)

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
