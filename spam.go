package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type spam struct {
	spamWords    []string
	warningWords []string
	blockedIps   []string
}

// IsSpam checks if a given string contains any spam words
func (s *spam) IsSpamContent(text string) bool {
	if text == "" {
		return true
	}
	for _, word := range s.spamWords {
		if strings.Contains(text, word) {
			return true
		}
	}
	numberOfWords := len(strings.Fields(text))
	var warningCount = 0
	for _, word := range s.warningWords {
		if strings.Contains(text, word) {
			warningCount++
		}
	}
	return warningCount > numberOfWords/6 // 16.6%
}

var ErrBlocked = fmt.Errorf("address blocked")

func (s *spam) AllowedAddress(clientIp string) error {
	for _, ip := range s.blockedIps {
		if strings.Contains(clientIp, ip) {
			return ErrBlocked
		}
	}
	return nil
}

func (s *spam) UpdateBlockedIps(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	//We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//Convert the body to type string
	sb := string(body)
	s.blockedIps = strings.Split(sb, "\n")
	return nil
}
