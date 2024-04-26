package main

import (
	"fmt"
	"strings"
)

var spamWords = []string{".ru/", "singleladies", ""}
var warningWords = []string{"dejt", "dating", "$$$", "cash", "money", "hacked", "password", "dick", "earn", "discount", "prince", "100%", "income", "fantastic", "bargain", "credit"}
var blockedIps = []string{"45.88.90.75", "45.88.90.115"}

// IsSpam checks if a given string contains any spam words
func IsSpamContent(text string) bool {
	if text == "" {
		return true
	}
	for _, word := range spamWords {
		if strings.Contains(text, word) {
			return true
		}
	}
	numberOfWords := len(strings.Fields(text))
	var warningCount = 0
	for _, word := range warningWords {
		if strings.Contains(text, word) {
			warningCount++
		}
	}
	return warningCount > numberOfWords/6 // 16.6%
}

var ErrBlocked = fmt.Errorf("address blocked")

func AllowedAddress(clientIp string) error {
	for _, ip := range blockedIps {
		if strings.Contains(clientIp, ip) {
			return ErrBlocked
		}
	}
	return nil
}
