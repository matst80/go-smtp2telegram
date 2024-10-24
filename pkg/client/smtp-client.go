package client

import (
	"bytes"
	"log"

	"github.com/emersion/go-msgauth/dkim"
)

type SmtpClient struct {
	options dkim.SignOptions
	Host    string
}

func MakeSmtpClient(host string, selector string, keyBytes []byte) (*SmtpClient, error) {
	key, err := getPrivateKeyFromBytes(keyBytes)
	if err != nil {
		return nil, err
	}
	options := dkim.SignOptions{
		Domain:   host,
		Selector: selector,
		Signer:   key,
	}
	return &SmtpClient{
		Host:    host,
		options: options,
	}, nil
}

func (s *SmtpClient) Send(msg Message) error {
	var b bytes.Buffer
	if err := dkim.Sign(&b, msg.Reader(), &s.options); err != nil {
		log.Fatal(err)
	}

	conn, err := GetClientFromMessage(&msg, true)
	if err != nil {
		log.Fatalf("Fatal: %v\n", err)
	}

	err = msg.Send(conn, b)

	conn.Quit()
	return err
}
