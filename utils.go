package main

import (
	"net"
	"strings"
)

func getIpFromAddr(addr net.Addr) string {
	return strings.Split(addr.String(), ":")[0]
}
