package getter

import (
	"net/url"
	"strings"
)

func GetDomainAndFirstPath(urlStr string) (string, string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", err
	}

	hostParts := strings.Split(parsedURL.Hostname(), ".")
	domain := ""
	if len(hostParts) > 2 {
		domain = strings.Join(hostParts[len(hostParts)-2:], ".")
	} else {
		domain = parsedURL.Hostname()
	}

	pathSegments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	firstPathSegments := ""
	if len(pathSegments) > 0 && pathSegments[0] != "" && pathSegments[0] != "/" {
		firstPathSegments = "/" + pathSegments[0]
	}

	return domain, firstPathSegments, nil
}

func GetDomain(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	hostParts := strings.Split(parsedURL.Hostname(), ".")
	domain := ""
	if len(hostParts) > 2 {
		domain = strings.Join(hostParts[len(hostParts)-2:], ".")
	} else {
		domain = parsedURL.Hostname()
	}

	return domain, nil

}
