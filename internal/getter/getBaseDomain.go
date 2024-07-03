package getter

import (
	"fmt"
	"net/url"

	"github.com/lormars/octohunter/internal/logger"
	"github.com/miekg/dns"
)

func GetBaseDomain(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Debugf("Error parsing URL: %v", err)
		return "", err
	}
	hostname := parsedURL.Hostname()
	labels := dns.SplitDomainName(hostname)
	if len(labels) < 2 {
		return "", fmt.Errorf("Invalid domain %s", hostname)
	}
	baseDomain := fmt.Sprintf("%s.%s", labels[len(labels)-2], labels[len(labels)-1])
	return baseDomain, nil
}
