package whois

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

const (
	// defaultWhoisServer is iana whois server
	defaultWhoisServer = "whois.iana.org"
	// defaultWhoisPort is default whois port
	defaultWhoisPort = "43"
)

func WhoIs(ctx context.Context, domain string) ([]byte, []byte, error) {
	var server, port string
	result, err := rawQuery(ctx, domain, defaultWhoisServer, defaultWhoisPort)
	if err != nil {
		return nil, nil, fmt.Errorf("whois: query for whois server failed: %w- %s", err, domain)
	}
	server, port = getServer(string(result))
	if server == "" {
		return nil, nil, errors.New("whois: no whois server found for domain: " + domain)
	}

	round := 0
	for {
		data, err := rawQuery(ctx, domain, server, port)
		if err != nil {
			return nil, nil, err
		}
		result = append(result, data...)

		refServer, refPort := getServer(string(result))
		if refServer == "" || refServer == server {
			return result, data, nil
		}
		server, port = refServer, refPort

		round++
		if round > 10 {
			return result, nil, errors.New("whois: too many round")
		}
	}
}

// rawQuery do raw query to the server
func rawQuery(ctx context.Context, domain, server, port string) ([]byte, error) {
	// whois.nic.tirol
	if server == "None" {
		return nil, nil
	}

	if server == "whois.arin.net" {
		domain = "n + " + domain
	}

	// See: https://github.com/likexian/whois/issues/17
	if server == "whois.godaddy" {
		server = "whois.godaddy.com"
	}

	// ascio
	if server == "www.ascio.com/products/availability-check/whois" {
		server = "whois.ascio.com"
	}

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(server, port))
	if err != nil {
		return nil, fmt.Errorf("whois: connect to whois server failed: %w", err)
	}
	defer conn.Close()

	if _, err = conn.Write([]byte(domain + "\r\n")); err != nil {
		return nil, fmt.Errorf("whois: send to whois server failed: %w", err)
	}

	buffer, err := io.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("whois: read from whois server failed: %w", err)
	}

	return buffer, nil
}

// getServer returns server from whois data
func getServer(data string) (string, string) {
	tokens := []string{
		"Registrar WHOIS Server: ",
		"whois: ",
		"ReferralServer: ",
	}

	for _, token := range tokens {
		start := strings.Index(data, token)
		if start != -1 {
			start += len(token)
			end := strings.Index(data[start:], "\n")
			server := strings.TrimSpace(data[start : start+end])
			server = strings.TrimPrefix(server, "http:")
			server = strings.TrimPrefix(server, "https:")
			server = strings.TrimPrefix(server, "whois:")
			server = strings.TrimPrefix(server, "rwhois:")
			server = strings.Trim(server, "/")
			port := defaultWhoisPort
			if strings.Contains(server, ":") {
				v := strings.Split(server, ":")
				server, port = v[0], v[1]
			}
			return server, port
		}
	}

	return "", ""
}
