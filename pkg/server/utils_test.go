package server

import (
	"testing"

	"github.com/mnako/letters"
)

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
