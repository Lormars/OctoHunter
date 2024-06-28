package common

import (
	"net/url"
	"strings"
)

// GetHostname extracts the hostname from a given string that can either be a domain or a URL.
func GetHostname(input string) string {
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		u, err := url.Parse(input)
		if err != nil {
			return input
		}
		return u.Hostname()
	}
	return input
}
