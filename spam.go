package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Spam struct {
	SpamWords    []string
	WarningWords []string
	BlockedIps   []string
	MaxSpamCount int
	Debug        bool
	spamIps      map[string]int
}

func (s *Spam) IsSpamHtml(html string) bool {
	if html == "" {
		return false
	}
	for _, word := range s.SpamWords {
		if strings.Contains(html, word) {
			if s.Debug {
				log.Println("Spam word found: ", word)
			}
			return true
		}
	}
	return false
}

func (s *Spam) IsSpamContent(text string) bool {
	numberOfWords := len(strings.Fields(text))
	var warningCount = 0
	for _, word := range s.WarningWords {
		f := strings.Count(text, word)
		if (f > 0) && s.Debug {
			log.Println("Warning found %s %d times", f, word)
		}
		warningCount += f
	}
	return warningCount > numberOfWords/6 // 16.6%
}

var ErrBlocked = fmt.Errorf("address blocked")

func (s *Spam) AllowedAddress(clientIp string) error {
	for _, ip := range s.BlockedIps {
		if strings.Contains(clientIp, ip) {
			return ErrBlocked
		}
	}
	return nil
}

func (s *Spam) UpdateBlockedIpsFromUrl(url string) error {
	lines, err := downloadLines(url)
	if err != nil {
		return err
	}
	s.BlockedIps = lines
	return nil
}

func (s *Spam) UpdateWarningWordsFromUrl(url string) error {
	lines, err := downloadLines(url)
	if err != nil {
		return err
	}
	s.WarningWords = lines
	return nil
}

func (s *Spam) LogSpamIp(ip string) error {
	if s.spamIps == nil {
		s.spamIps = make(map[string]int)
	}
	s.spamIps[ip]++
	if s.spamIps[ip] > s.MaxSpamCount {
		s.BlockedIps = append(s.BlockedIps, ip)
		log.Print("Blocking ip", ip)
	}
	return nil
}

func downloadLines(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(body), "\n"), nil
}
