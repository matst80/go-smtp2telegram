package server

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

	if s.IsSpamContent("") {
		t.Errorf("Expected to be ok")
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

	err := s.UpdateWarningWordsFromUrl("https://raw.githubusercontent.com/matst80/go-smtp2telegram/main/data/warning-words.txt")
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	if len(s.WarningWords) < 100 {
		t.Errorf("Expected list to be updated with more than 100 words, got %d", len(s.WarningWords))
	}
	isSpam := s.IsSpamContent(`To: slask@knatofs.se
Cc: 
Bcc: 
Date: Sat, 27 Apr 2024 15:33:14 +0200
Subject: Hej
Detta ska fungera bättre`)
	if isSpam {
		t.Errorf("Expected to not be spam")
	}
}

func TestSpamIdLogging(t *testing.T) {
	ip := "127.0.0.2"
	spm := &Spam{
		SpamWords:    []string{},
		WarningWords: []string{},
		BlockedIps:   []string{},
		MaxSpamCount: 3,
	}
	for i := 0; i < 3; i++ {
		spm.LogSpamIp(ip)
		if err := spm.AllowedAddress(ip); err != nil {
			t.Errorf("Expected ip to accepted")
		}
	}

	spm.LogSpamIp(ip)
	if err := spm.AllowedAddress(ip); err == nil {
		t.Errorf("Expected ip to blocked")
	}
}
