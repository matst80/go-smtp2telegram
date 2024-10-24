package server

import (
	"fmt"
	"strings"
)

func (s *Session) textContent(r Recipient, c *ClassificationResult) string {
	var sb strings.Builder
	if s.HasValidDkim {
		sb.WriteString(fmt.Sprintf("Subject: %s\nFrom: %s\n", s.Email.Headers.Subject, s.From))

		if r.WantsDebugInfo {
			sb.WriteString(fmt.Sprintf("To: %s\nIp: %s\n", r.Address, s.Client))
		}

		if c != nil {
			sb.WriteString(fmt.Sprintf("Spam rating: %.2f\n\n%s\n\n", c.SpamRating, c.Summary))
		}

		sb.WriteString(s.Email.Text)
	} else {
		sb.WriteString(fmt.Sprintf("Spam from: %s, Subject: %s\n", s.From, s.Email.Headers.Subject))
		if c != nil {
			sb.WriteString(fmt.Sprintf("Spam rating: %.2f\n\n%s", c.SpamRating, c.Summary))
		}
	}

	userData, ok := s.StoredData[r.ChatId]
	if ok && s.backend.HashGenerator != nil {

		sb.WriteString(fmt.Sprintf("\n\nRead original: %s", userData.Html.WebUrl(s.backend.Config.BaseUrl, s.backend.HashGenerator)))
		for _, attachment := range userData.Attachments {
			sb.WriteString(fmt.Sprintf("\nAttachment: %s", attachment.WebUrl(s.backend.Config.BaseUrl, s.backend.HashGenerator)))
		}
	}

	return sb.String()
}
