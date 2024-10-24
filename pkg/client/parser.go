package client

import (
	"fmt"
	"strings"
)

func ParseMessage(i, fromEmail, defaultSubject string) (*Message, error) {
	lines := strings.Split(i, "\n")
	fromDomain, err := getDomainFromEmail(fromEmail)
	if err != nil {
		return nil, err
	}
	message := &Message{
		Headers:   make(map[string]string),
		Subject:   defaultSubject,
		From:      fromEmail,
		MessageID: MakeMessageId(fromDomain),
	}
	headersDone := false
	bodyLines := []string{}
	for _, line := range lines {
		if !headersDone {
			if strings.HasPrefix(line, "To ") || strings.HasPrefix(line, "to ") {
				message.To = strings.TrimSpace(line[3:])
			} else if strings.HasPrefix(line, "Subject ") || strings.HasPrefix(line, "subject ") {
				message.Subject = strings.TrimSpace(line[8:])
			} else {
				headersDone = true
			}
		}
		if headersDone {
			bodyLines = append(bodyLines, line)

		}
	}
	message.Body = []byte(strings.Join(bodyLines, "\r\n"))
	if message.To == "" {
		return nil, fmt.Errorf("no to header")
	}
	if message.Subject == "" {
		return nil, fmt.Errorf("no subject header")
	}

	return message, nil
}
