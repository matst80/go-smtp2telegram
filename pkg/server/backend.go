package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/emersion/go-msgauth/dkim"
	"github.com/emersion/go-smtp"
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mnako/letters"
)

type Backend struct {
	HashGenerator  HashGenerator
	Bot            *botapi.BotAPI
	SpamClassifier SpamClassification
	Config         *Config
	SpamChecker    SpamChecker
}

type Session struct {
	backend      *Backend
	Client       net.Addr
	HasValidDkim bool
	From         string
	To           []Recipient
	Email        letters.Email
	MailId       string
	StoredData   map[int64]StorageResult
}

type Recipient struct {
	WantsDebugInfo bool
	Address        string
	User           *User
	ChatId         int64
}

func (bkd *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	client := c.Conn().RemoteAddr()

	ip := getIpFromAddr(client)
	err := bkd.SpamChecker.AllowedAddress(ip)
	if err != nil {
		if bkd.Config.AllowBlockedIps {
			log.Printf("Allowing blocked address %s", client)
		} else {
			log.Printf("Blocked address %s", client)
			return nil, &smtp.SMTPError{Code: 550, Message: "Blocked address"}
		}
	}

	return &Session{
		Client:     client,
		backend:    bkd,
		To:         []Recipient{},
		Email:      letters.Email{},
		StoredData: map[int64]StorageResult{},
	}, nil
}

func (s *Session) AuthPlain(username, password string) error {
	log.Printf("Someone is trying to login: %s, %s", username, password)
	return nil
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	if s.backend.Config.CustomFromMessage != nil {
		for _, i := range s.backend.Config.CustomFromMessage {
			if i.Email == from {
				log.Printf("Custom message for %s: %s", from, i.Message)
				return &smtp.SMTPError{Code: 550, Message: i.Message}
			}
		}
	}
	s.From = from
	return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	for _, u := range s.backend.Config.Users {
		if to == u.Email {
			s.To = append(s.To, Recipient{
				ChatId:         u.ChatId,
				WantsDebugInfo: u.DebugInfo,
				Address:        to,
				User:           &u,
			})
			return nil
		}
	}
	return &smtp.SMTPError{Code: 550, Message: "User not found"}
}

func checkDkim(r io.Reader, from string) bool {
	verifications, err := dkim.Verify(r)
	if err != nil {
		log.Printf("Error verifying DKIM: %v", err)
		return false
	}
	for _, v := range verifications {
		if v.Err == nil {
			log.Printf("Valid signature for: %s (from: %s)", v.Domain, from)
			return true
		}
	}
	return false
}

func (s *Session) Data(r io.Reader) error {

	if len(s.To) > 0 {
		var buf bytes.Buffer
		tee := io.TeeReader(r, &buf)

		email, err := letters.ParseEmail(tee)
		if err != nil {
			return err
		}
		s.HasValidDkim = checkDkim(&buf, s.From)
		s.Email = email

		mailId := getEmailFileName(email.Headers)

		s.MailId = mailId
		for _, userId := range s.To {
			data, err := saveMail(mailId, userId.ChatId, s.Email)
			if err != nil {
				log.Printf("Error saving email: %v", err)
			} else {
				s.StoredData[userId.ChatId] = data
			}
		}

	}
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return s.handleMail()
}

func (s *Session) handleMail() error {
	isSpam := s.backend.SpamChecker.IsSpamHtml(s.Email.HTML) || s.backend.SpamChecker.IsSpamContent(s.Email.Text)
	ip := getIpFromAddr(s.Client)
	if isSpam {
		s.backend.SpamChecker.LogSpamIp(ip)
		log.Printf("Spam detected (%s) [%s]", s.From, ip)
		return nil
	}
	if s.Email.Headers.Subject == "" && s.Email.Text == "" {
		s.backend.SpamChecker.LogSpamIp(ip)
		log.Printf("No subject, discarding email from %s (%s)", s.From, ip)
		return nil
	}
	var err error
	var result *ClassificationResult
	if len(s.To) > 0 {

		if s.Email.HTML != "" && s.backend.SpamClassifier != nil && len(s.Email.HTML) < (1024*20) {
			if result, err = s.backend.SpamClassifier.Classify(s.Email.HTML); err != nil {
				log.Printf("Error classifying email: %v", err)
			}
		}
		for _, r := range s.To {

			content := s.textContent(r, result)

			msg := botapi.NewMessage(r.ChatId, content)

			if r.WantsDebugInfo {
				msg.ReplyMarkup = botapi.NewReplyKeyboard(botapi.NewKeyboardButtonRow(
					botapi.NewKeyboardButton("/block "+ip),
					botapi.NewKeyboardButton(fmt.Sprintf("/send To %s\nSubject %s\n", s.From, s.Email.Headers.Subject)),
				))
			}

			_, err = s.backend.Bot.Send(msg)
			log.Printf("Sent email to %d (%s)", r.ChatId, r.Address)

		}
	} else {
		err = s.backend.SpamChecker.LogSpamIp(ip)
		log.Printf("Discarding email, no recipient, from: %s (%s)", s.From, s.Client)
	}

	return err
}
