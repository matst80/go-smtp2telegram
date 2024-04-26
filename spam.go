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

func downloadLines(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	//We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//Convert the body to type string
	sb := string(body)
	return strings.Split(sb, "\n"), nil
}

func (s *spam) UpdateBlockedIpsFromUrl(url string) error {
	lines, err := downloadLines(url)
	if err != nil {
		return err
	}
	s.blockedIps = lines
	return nil
}

func (s *spam) UpdateWarningWordsFromUrl(url string) error {
	lines, err := downloadLines(url)
	if err != nil {
		return err
	}
	s.warningWords = lines
	return nil
}