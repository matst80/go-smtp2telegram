package main

import (
	"fmt"
	"strings"
)

func (s *session) textContent(r rcpt, c *ClassificationResult) string {
	var sb strings.Builder
	if s.HasValidDkim {
		sb.WriteString(fmt.Sprintf("From: %s\nSubject: %s\n", s.From, s.Email.Headers.Subject))

		if r.extraInfo {
			sb.WriteString(fmt.Sprintf("To: %s\nIp: %s\n", r.address, s.Client))
		}

		if c != nil {
			sb.WriteString(fmt.Sprintf("Spam rating: %.2f\n\n%s\n\n", c.SpamRating, c.Summary))
		}

		sb.WriteString(s.Email.Text)
	} else {
		sb.WriteString(fmt.Sprintf("Spam from: %s, Subject: %s", s.From, s.Email.Headers.Subject))
	}

	userData, ok := s.StoredData[r.chatId]
	if ok && s.backend.HashGenerator != nil {
		hashQuery := s.backend.HashGenerator.CreateHash(fmt.Sprintf("%d%s", r.chatId, s.MailId))

		sb.WriteString(fmt.Sprintf("\n\nRead original: %s", userData.Html.WebUrl(s.backend.Config.BaseUrl, hashQuery)))
		for _, attachment := range userData.Attachments {
			sb.WriteString(fmt.Sprintf("\nAttachment: %s", attachment.WebUrl(s.backend.Config.BaseUrl, hashQuery)))
		}
	}

	return sb.String()
}
