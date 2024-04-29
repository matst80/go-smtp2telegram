package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/mnako/letters"

	"github.com/emersion/go-msgauth/dkim"
	"github.com/emersion/go-smtp"
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type backend struct {
	hash         *hash
	bot          *botapi.BotAPI
	aiClassifier *aiClassifier
	config       *Config
	spam         *Spam
}

type session struct {
	client       net.Addr
	backend      *backend
	hasValidDkim bool
	from         string
	to           []rcpt
	email        letters.Email
	mailId       string
}

type rcpt struct {
	extraInfo bool
	address   string
	chatId    int64
}

func (bkd *backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	client := c.Conn().RemoteAddr()

	ip := getIpFromAddr(client)
	err := bkd.spam.AllowedAddress(ip)
	if err != nil {
		log.Printf("Blocked address %s", client)
		return nil, &smtp.SMTPError{Code: 550, Message: "Blocked address"}
	}

	return &session{
		client:  client,
		backend: bkd,
		to:      []rcpt{},
		email:   letters.Email{},
	}, nil
}

func (s *session) AuthPlain(username, password string) error {
	return nil
}

func (s *session) Mail(from string, opts *smtp.MailOptions) error {
	if s.backend.config.CustomFromMessage != nil {
		for _, i := range s.backend.config.CustomFromMessage {
			if i.Email == from {
				log.Printf("Custom message for %s: %s", from, i.Message)
				return &smtp.SMTPError{Code: 550, Message: i.Message}
			}
		}
	}
	s.from = from
	return nil
}

func (s *session) Rcpt(to string, opts *smtp.RcptOptions) error {
	for _, u := range s.backend.config.Users {
		if to == u.Email {
			s.to = append(s.to, rcpt{chatId: u.ChatId, extraInfo: u.DebugInfo, address: to})
			return nil
		}
	}
	return &smtp.SMTPError{Code: 550, Message: "User not found"}
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

func hasValidDkim(r io.Reader, from string) bool {
	verifications, err := dkim.Verify(r)
	if err != nil {
		// log.Printf("Error verifying DKIM: %v", err)
		return false
	}
	for _, v := range verifications {
		if v.Err == nil {
			// log.Printf("Valid signature for: %s", v.Domain)
			return strings.Contains(from, v.Domain)
		}
	}
	return false
}

func (s *session) Data(r io.Reader) error {

	if len(s.to) > 0 {
		var buf bytes.Buffer
		tee := io.TeeReader(r, &buf)

		s.hasValidDkim = hasValidDkim(&buf, s.from)

		email, err := letters.ParseEmail(tee)
		if err != nil {
			return err
		}
		s.email = email
		if email.HTML != "" {
			mailId := getEmailFileName(email.Headers)

			s.mailId = mailId
			for _, userId := range s.to {
				go saveHtml(mailId, userId.chatId, s.email)
			}
		}
	}
	return nil
}

func (s *session) Reset() {}

func textContent(s *session, r rcpt, c *classificationResult) string {
	prefix := getPrefix(r, s)
	suffix := getSuffix(s, r)
	if c.SpamRating > -1.0 {
		prefix = fmt.Sprintf("%s\nSpam rating: %.2f\nSummary: %s", prefix, c.SpamRating, c.Summary)
	}
	validText := senderVerified(s)
	return fmt.Sprintf("From: %s (%s)\nSubject: %s%s\n\n%s%s", s.from, validText, s.email.Headers.Subject, prefix, s.email.Text, suffix)
}

func senderVerified(s *session) string {
	if s.hasValidDkim {
		return "verified"
	}
	return "no dkim signature"
}

func getSuffix(s *session, r rcpt) string {
	if s.mailId != "" && s.email.HTML != "" {
		hashQuery := s.backend.hash.createSimpleHash(fmt.Sprintf("%d%s", r.chatId, s.mailId))
		return fmt.Sprintf("\n\nRead original\n%s/mail/%d/%s.html?hash=%s", s.backend.config.BaseUrl, r.chatId, s.mailId, hashQuery)
	}
	return ""
}

func getPrefix(r rcpt, s *session) string {
	if r.extraInfo {
		return fmt.Sprintf("\nTo: %s\nIp: %s", r.address, s.client)
	}
	return ""
}

func (s *session) Logout() error {

	isSpam := s.backend.spam.IsSpamHtml(s.email.HTML) || s.backend.spam.IsSpamContent(s.email.Text)
	ip := getIpFromAddr(s.client)
	if isSpam {
		s.backend.spam.LogSpamIp(ip)
		log.Printf("Spam detected (%s) [%s]", s.from, ip)
		return nil
	}
	if len(s.to) > 0 {
		result := &classificationResult{
			SpamRating: -1,
			Summary:    "",
		}
		if s.email.HTML != "" && s.backend.aiClassifier != nil && len(s.email.HTML) < 1024*10 {
			if err := s.backend.aiClassifier.classify(s.email.HTML, result); err != nil {
				log.Printf("Error classifying email: %v", err)
			}
		}
		for _, r := range s.to {

			content := textContent(s, r, result)

			msg := botapi.NewMessage(r.chatId, content)

			s.backend.bot.Send(msg)
			log.Printf("Sent email to %d", r.chatId)

		}
	} else {
		s.backend.spam.LogSpamIp(ip)
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
		aiClassifier: newAiClassifier(&config.OpenAi),
		hash:         h,
		bot:          bot,
		config:       config,
		spam:         spm,
	})
	go WebServer(h)

	log.Printf("Bot authorized [%s]", bot.Self.UserName)

	s.Addr = config.Listen
	s.Domain = config.Domain
	s.AllowInsecureAuth = true

	if os.Getenv("DEBUG") == "true" {
		bot.Debug = true
		spm.Debug = true
		s.Debug = os.Stdout
	}

	go ListenForMessages(bot, &commandHandler{
		spam:   spm,
		config: config,
		bot:    bot,
	})

	log.Println("Starting SMTP server at", s.Addr)
	log.Fatal(s.ListenAndServe())
}
