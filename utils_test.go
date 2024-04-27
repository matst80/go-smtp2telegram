package main

import (
	"testing"

	"github.com/mnako/letters"
)

func TestEmailAddressSplitting(t *testing.T) {
	emails := getValidEmailAddresses("a@b.se test hej test+1@a.se")
	if len(emails) != 2 {
		t.Errorf("Expected 2 emails, got %v", emails)
	}
	if emails[0] != "a@b.se" {
		t.Errorf("Expected first email to be a@b.se, got %v", emails[0])
	}
	if emails[1] != "test+1@a.se" {
		t.Errorf("Expected second email to be test+1@a.se, got %v", emails[1])
	}
}

func TestEmailFileName(t *testing.T) {
	headers := letters.Headers{
		MessageID: "123",
		Subject:   "test",
	}
	if f := getEmailFileName(headers); f != "123" {
		t.Errorf("Expected 123, got %v", f)
	}
	headers = letters.Headers{
		MessageID: "",
		Subject:   "Subject Test åäö/&%¤#",
	}
	if f := getEmailFileName(headers); f != "subject-test-" {
		t.Errorf("Expected a file name, got %v", f)
	}
	headers = letters.Headers{}
	if f := getEmailFileName(headers); f == "" {
		t.Errorf("Expected a file name")
	}
}
