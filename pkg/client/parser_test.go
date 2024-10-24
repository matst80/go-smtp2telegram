package client

import "testing"

func TestParserOk(t *testing.T) {

	i := "To user@email.com\nSubject Test subject!\na\nb\nc"
	message, err := ParseMessage(i, "mats@tornberg.me", "default subject")

	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	if message.To != "user@email.com" {
		t.Errorf("To: %s", message.To)
	}

	if message.Subject != "Test subject!" {
		t.Errorf("Subject: %s", message.Subject)
	}

	if string(message.Body) != "a\r\nb\r\nc" {
		t.Errorf("Body: %v, %v", message.Body, []byte("a\r\nb\r\nc"))
	}

}

func TestParserMissingTo(t *testing.T) {

	i := "Subject Test subject!\na\nb\nc"
	_, err := ParseMessage(i, "mats@tornberg.me", "default subject")

	if err == nil {
		t.Errorf("Expected error")
		return
	}

}

func TestParserMissingEverything(t *testing.T) {

	i := ""
	_, err := ParseMessage(i, "mats@tornberg.me", "default subject")

	if err == nil {
		t.Errorf("Expected error")
		return
	}

}

func TestParserMissingSubject(t *testing.T) {

	i := "To user@email.com\na\nb\nc"
	message, err := ParseMessage(i, "mats@tornberg.me", "default subject")

	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	if message.To != "user@email.com" {
		t.Errorf("To: %s", message.To)
	}

	if message.Subject != "default subject" {
		t.Errorf("Subject: %s", message.Subject)
	}

	if string(message.Body) != "a\r\nb\r\nc" {
		t.Errorf("Body: %v, %v", message.Body, []byte("a\r\nb\r\nc"))
	}

}
