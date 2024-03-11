package test_domains_test

import (
	whoisparser "github.com/likexian/whois-parser"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestDomains(t *testing.T) {
	entries, err := os.ReadDir("./")
	if err != nil {
		panic(err)
	}

	for _, e := range entries {
		if e.Name() == "main_test.go" {
			continue
		}

		t.Run(e.Name(), func(t *testing.T) {
			nameSlices := strings.Split(e.Name(), ".")
			expirationUnix, err := strconv.Atoi(nameSlices[len(nameSlices)-1])
			if err != nil {
				panic(err)
			}

			b, err := os.ReadFile(e.Name())
			if err != nil {
				panic(err)
			}
			info, err := whoisparser.Parse(string(b))
			if err != nil {
				panic(err)
			}
			expiration := int64(0)
			if info.Domain.ExpirationDateInTime != nil {
				expiration = info.Domain.ExpirationDateInTime.Unix()
			}
			if expiration != int64(expirationUnix) {
				t.Errorf("Expected %d, but got %d", expirationUnix, expiration)
			}
		})
	}
}
