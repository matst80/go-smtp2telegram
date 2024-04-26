package main

import "testing"

var spamWords = []string{".ru/", "singleladies", "<script"}
var warningWords = []string{"dejt", "dating", "$$$", "cash", "money", "hacked", "password", "dick", "earn", "discount", "prince", "100%", "income", "fantastic", "bargain", "credit"}
var blockedIps = []string{"45.88.90.75", "45.88.90.115", "87.121.105.109"}

// test ip blocking
func TestAllowedAddress(t *testing.T) {
	s := Spam{
		SpamWords:    spamWords,
		WarningWords: warningWords,
		BlockedIps:   blockedIps,
	}
	err := s.AllowedAddress("1.1.1.1")
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestBlockedAddress(t *testing.T) {
	s := Spam{
		SpamWords:    spamWords,
		WarningWords: warningWords,
		BlockedIps:   blockedIps,
	}
	err := s.AllowedAddress("45.88.90.75")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestSpamContent(t *testing.T) {
	s := Spam{
		SpamWords:    spamWords,
		WarningWords: warningWords,
		BlockedIps:   blockedIps,
	}

	spam := "This is a spam message with a <a href=\"test.ru/t\">link</a>"
	if !s.IsSpamHtml(spam) {
		t.Errorf("Expected to be spam")
	}

	warning := "This is a warning message with a dating word"
	if s.IsSpamContent(warning) {
		t.Errorf("Expected to be acceptable")
	}

	warning = "This is a dating cash discount fantastic with a dating word 100% awesome"
	if !s.IsSpamContent(warning) {
		t.Errorf("Expected to be spam")
	}

	if !s.IsSpamContent("") {
		t.Errorf("Expected to be spam")
	}
}

func TestUpdateBlackList(t *testing.T) {
	s := Spam{
		SpamWords:    spamWords,
		WarningWords: warningWords,
		BlockedIps:   blockedIps,
	}

	err := s.UpdateBlockedIpsFromUrl("https://lists.blocklist.de/lists/mail.txt")
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	if len(s.BlockedIps) < 100 {
		t.Errorf("Expected list to be updated with more than 100 ips, got %d", len(s.BlockedIps))
	}
}

func TestUpdateWarningWords(t *testing.T) {
	s := Spam{
		SpamWords:    spamWords,
		WarningWords: warningWords,
		BlockedIps:   blockedIps,
	}

	err := s.UpdateWarningWordsFromUrl("https://gist.githubusercontent.com/prasidhda/13c9303be3cbc4228585a7f1a06040a3/raw/b8905a4012146212f7c7af37e379f90980310642/common%2520spam%2520words%25202020")
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	if len(s.WarningWords) < 100 {
		t.Errorf("Expected list to be updated with more than 100 words, got %d", len(s.WarningWords))
	}
}

func TestSpamIdLogging(t *testing.T) {
	config := GetConfig()
	ip := "127.0.0.2"
	spm := &Spam{
		SpamWords:    []string{},
		WarningWords: []string{},
		BlockedIps:   config.BlockedIps,
		MaxSpamCount: 3,
	}
	spm.LogSpamIp(ip)
	if err := spm.AllowedAddress(ip); err != nil {
		t.Errorf("Expected ip to accepted")
	}
	spm.LogSpamIp(ip)
	if err := spm.AllowedAddress(ip); err != nil {
		t.Errorf("Expected ip to accepted")
	}
	spm.LogSpamIp(ip)
	if err := spm.AllowedAddress(ip); err != nil {
		t.Errorf("Expected ip to accepted")
	}
	spm.LogSpamIp(ip)
	if err := spm.AllowedAddress(ip); err == nil {
		t.Errorf("Expected ip to blocked")
	}
}
