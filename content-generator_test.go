package main

import (
	"net"
	"testing"

	"github.com/mnako/letters"
)

func TestContentGenerator(t *testing.T) {
	s := &session{
		backend: &backend{
			HashGenerator: &hash{},
			Config: &Config{
				BaseUrl: "http://example.com",
			},
		},
		HasValidDkim: true,
		Email: letters.Email{
			Headers: letters.Headers{
				Subject: "Subject",
			},

			Text: "Message",
		},
		From: "from@example.com",
		Client: &net.TCPAddr{
			IP: net.ParseIP("127.0.0.1"), Port: 1,
		},
		StoredData: map[int64]StorageResult{
			1: {
				Html: StoredFile{
					UserId:   1,
					FileName: "test.html",
				},
				Attachments: []StoredFile{
					{
						UserId:   1,
						FileName: "test-0",
					},
				},
			},
		},
	}
	content := s.textContent(rcpt{
		extraInfo: true,
		address:   "test@example.com",
		chatId:    1,
	}, &ClassificationResult{
		SpamRating: 0.5,
		Summary:    "AI SUMMARY",
	},
	)
	expected := `From: from@example.com
Subject: Subject
To: test@example.com
Ip: 127.0.0.1:1
Spam rating: 0.50

AI SUMMARY

Message

Read original: http://example.com/mail/1/test.html?hash=c4ca4238a0b923820dcc509a6f75849b
Attachment: http://example.com/mail/1/test-0?hash=c4ca4238a0b923820dcc509a6f75849b`

	if content != expected {
		t.Errorf("Expected %s, got %s", expected, content)
	}
}

func TestContentGeneratorNoDkim(t *testing.T) {
	s := &session{
		backend: &backend{
			HashGenerator: &hash{},
			Config: &Config{
				BaseUrl: "http://example.com",
			},
		},
		HasValidDkim: false,
		Email: letters.Email{
			Headers: letters.Headers{
				Subject: "Subject",
			},

			Text: "Message",
		},
		From: "junk@email.com",
		Client: &net.TCPAddr{
			IP: net.ParseIP("127.0.0.1"), Port: 1,
		},
		StoredData: map[int64]StorageResult{
			1: {
				Html: StoredFile{
					UserId:   1,
					FileName: "test.html",
				},
				Attachments: []StoredFile{
					{
						UserId:   1,
						FileName: "test-0",
					},
				},
			},
		},
	}
	content := s.textContent(rcpt{
		extraInfo: true,
		address:   "test@example.com",
		chatId:    1,
	}, &ClassificationResult{
		SpamRating: 0.5,
		Summary:    "AI SUMMARY",
	},
	)
	expected := `Spam from: junk@email.com, Subject: Subject

Read original: http://example.com/mail/1/test.html?hash=c4ca4238a0b923820dcc509a6f75849b
Attachment: http://example.com/mail/1/test-0?hash=c4ca4238a0b923820dcc509a6f75849b`

	if content != expected {
		t.Errorf("Expected %s, got %s", expected, content)
	}
}
