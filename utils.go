package main

import (
	"net"
	"strings"
)

func getIpFromAddr(addr net.Addr) string {
	return strings.Split(addr.String(), ":")[0]
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
