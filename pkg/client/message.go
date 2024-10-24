package client

import (
	"bytes"
	"fmt"
	"io"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/rand"
)

type Message struct {
	To        string
	From      string
	Subject   string
	MessageID string
	Headers   map[string]string
	Body      []byte
}

func MakeMessageId(hostname string) string {
	now := time.Now()
	return fmt.Sprintf("<%d.%d.%d@%s>", now.Unix(), now.UnixNano(), rand.Int63(), hostname)
}

func NewMessage(from string) (*Message, error) {
	hostname, err := getDomainFromEmail(from)
	if err != nil {
		return nil, err
	}
	return &Message{
		From:      from,
		MessageID: MakeMessageId(hostname),
	}, nil
}

func (m *Message) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("To: %s\r\n", m.To))
	sb.WriteString(fmt.Sprintf("From: %s\r\n", m.From))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", m.Subject))
	sb.WriteString(fmt.Sprintf("Message-ID: %s\r\n", m.MessageID))
	if m.Headers != nil {
		for k, v := range m.Headers {
			sb.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}
	sb.WriteString("\r\n")
	sb.WriteString(string(m.Body))
	sb.WriteString("\r\n")
	return sb.String()
}

func (m *Message) Reader() io.Reader {
	return strings.NewReader(m.String())
}

func getDomainFromEmail(email string) (string, error) {
	r := regexp.MustCompile(`@([a-zA-Z0-9\.\-\_]+)`)
	matches := r.FindStringSubmatch(email)
	if len(matches) == 0 {
		return "", fmt.Errorf("could not extract domain from email address")
	}
	return matches[1], nil
}

func (m *Message) FromDomain() (string, error) {
	return getDomainFromEmail(m.From)
}

func (m *Message) LookupMX() ([]string, error) {
	domain, err := getDomainFromEmail(m.To)
	if err != nil {
		return nil, err
	}
	return getMxRecords(domain)
}

func (m *Message) Send(conn *smtp.Client, b bytes.Buffer) error {
	if err := conn.Mail(m.From); err != nil {
		return err
	}
	if err := conn.Rcpt(m.To); err != nil {
		return err
	}
	wc, err := conn.Data()
	if err != nil {
		return err
	}
	if _, err := b.WriteTo(wc); err != nil {
		return err
	}

	return wc.Close()
}
