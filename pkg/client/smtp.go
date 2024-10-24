package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/smtp"
	"time"
)

func getMxRecords(domain string) ([]string, error) {
	resolver := net.Resolver{
		PreferGo: true,
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
	mxs, err := resolver.LookupMX(ctx, domain)
	cancel()
	if err != nil {
		if dnserr, ok := err.(*net.DNSError); !ok || dnserr.Err != "no such host" {
			return nil, err
		}
	}
	if len(mxs) > 0 {
		hosts := make([]string, len(mxs))
		for i, mx := range mxs {
			hosts[i] = mx.Host
		}
		return hosts, nil
	}
	// fall back to a record.
	return []string{domain}, nil
}

func getClientFromMx(hosts []string) (*smtp.Client, string, error) {

	dialer := &net.Dialer{}
	var err error
	for _, host := range hosts {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		conn, err := dialer.DialContext(ctx, "tcp", host+":smtp")
		cancel()
		if err == nil {
			c, err := smtp.NewClient(conn, host)
			if err == nil {
				return c, host, nil
			}
		}
	}

	// fall back to 587 - mail submission port
	for _, host := range hosts {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		conn, err := dialer.DialContext(ctx, "tcp", host+":587")
		cancel()
		if err == nil {
			c, err := smtp.NewClient(conn, host)
			if err == nil {
				return c, host, nil
			}
		}
	}

	return nil, "", err
}

func GetClientFromMessage(message *Message, useTls bool) (*smtp.Client, error) {
	hosts, err := message.LookupMX()
	if err != nil {
		return nil, err
	}
	conn, host, err := getClientFromMx(hosts)
	if err != nil {
		return nil, err
	}
	fromDomain, err := message.FromDomain()
	if err != nil {
		return nil, err
	}
	if err := conn.Hello(fromDomain); err != nil {
		return nil, err
	}
	if !useTls {
		return conn, nil
	}
	conn.StartTLS(&tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
	})
	return conn, nil
}
