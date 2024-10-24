package client

import (
	"io"
	"log"
	"os"
)

func TestMail() {

	file, err := os.Open("tornberg.pem")
	if err != nil {
		log.Fatalf("Fatal: %v\n", err)
	}

	pkey, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Fatal: %v\n", err)
	}
	client, err := MakeSmtpClient("tornberg.me", "mail", pkey)
	if err != nil {
		log.Fatalf("Fatal: %v\n", err)
	}
	message := Message{
		To:        "mats.tornberg@gmail.com",
		From:      "mats@tornberg.me",
		Subject:   "Ett nytt mail!",
		MessageID: MakeMessageId("tornberg.me"),
		Body:      []byte("This is a test message"),
	}

	client.Send(message)

}
