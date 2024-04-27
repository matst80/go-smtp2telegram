package main

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/mnako/letters"
)

func getIpFromAddr(addr net.Addr) string {
	var clientIp string
	s := addr.String()
	if strings.ContainsRune(s, ':') {
		clientIp, _, _ = net.SplitHostPort(s)
	} else {
		clientIp = s
	}
	return clientIp
}

func getValidEmailAddresses(input string) []string {
	emails := strings.Split(input, " ")
	var validEmails []string
	for _, email := range emails {
		if strings.Contains(email, "@") {
			validEmails = append(validEmails, email)
		}
	}
	return validEmails
}

var m1 = regexp.MustCompile(`[^\w\-.]`)

func fileSaveName(emailFileName string) string {
	emailFileName = m1.ReplaceAllString(strings.ReplaceAll(emailFileName, " ", "-"), "")
	return strings.ToLower(emailFileName)
}

func getEmailFileName(headers letters.Headers) string {
	if headers.MessageID != "" {
		return fileSaveName(string(headers.MessageID))
	}
	if headers.Subject != "" {
		return fileSaveName(headers.Subject)
	}
	// generate unique id
	id := time.Now().UnixNano()
	return fileSaveName(fmt.Sprint(id))
}
