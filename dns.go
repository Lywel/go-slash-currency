package main

import (
	"fmt"
	"net"
	"strings"
)

var (
	dnsSeeds = []string{
		"seed.slash-currency.tk",
		"test.google.com",
	}
)

func resolveDNSSeeds() (val []string, sync []string, err error) {
	for _, seed := range dnsSeeds {
		txts, err := net.LookupTXT(seed)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, txt := range txts {
			split := strings.Split(txt, "=")
			// Check for rong format
			if len(split) != 2 {
				continue
			}
			seedType, seedVal := split[0], split[1]
			// Check for empty strings
			if seedType == "" || seedVal == "" {
				continue
			}
			// Check for type
			switch seedType {
			case "s":
				sync = append(sync, seedVal)
				break
			case "v":
				val = append(val, seedVal)
				break
			default:
				continue
			}
		}
	}
	return
}
