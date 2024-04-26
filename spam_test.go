package main

import "testing"

var spamWords = []string{".ru/", "singleladies", "<script"}
var warningWords = []string{"dejt", "dating", "$$$", "cash", "money", "hacked", "password", "dick", "earn", "discount", "prince", "100%", "income", "fantastic", "bargain", "credit"}
var blockedIps = []string{"45.88.90.75", "45.88.90.115", "87.121.105.109"}

// test ip blocking
func TestAllowedAddress(t *testing.T) {
	s := spam{
		spamWords:    spamWords,
		warningWords: warningWords,
		blockedIps:   blockedIps,
	}
	err := s.AllowedAddress("1.1.1.1")
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestBlockedAddress(t *testing.T) {
	s := spam{
		spamWords:    spamWords,
		warningWords: warningWords,
		blockedIps:   blockedIps,
	}
	err := s.AllowedAddress("45.88.90.75")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestSpamContent(t *testing.T) {
	s := spam{
		spamWords:    spamWords,
		warningWords: warningWords,
		blockedIps:   blockedIps,
	}

	spam := "This is a spam message with a .ru/ link"
	if !s.IsSpamContent(spam) {
		t.Errorf("Expected to be spam")
	}

	warning := "This is a warning message with a dating word"
	if s.IsSpamContent(warning) {
		t.Errorf("Expected to be acceptable")
	}

	if !s.IsSpamContent("") {
		t.Errorf("Expected to be spam")
	}
}

func TestUpdateBlackList(t *testing.T) {
	s := spam{
		spamWords:    spamWords,
		warningWords: warningWords,
		blockedIps:   blockedIps,
	}

	err := s.UpdateBlockedIps("https://lists.blocklist.de/lists/mail.txt")
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	if len(s.blockedIps) < 100 {
		t.Errorf("Expected list to be updated with more than 100 ips, got %d", len(s.blockedIps))
	}
}
