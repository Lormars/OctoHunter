package parser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/lormars/octohunter/internal/logger"
	"golang.org/x/net/html"
)

func resolveURL(baseURL, href string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	rel, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	resolved := base.ResolveReference(rel)

	if !strings.HasSuffix(resolved.Hostname(), base.Hostname()) {
		return "", fmt.Errorf("not within scope")
	}
	return resolved.String(), nil
}

func ExtractUrls(baseUrl, response string) []string {
	urls := []string{}

	resp := strings.NewReader(response)

	z := html.NewTokenizer(resp)

	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return urls
		case html.StartTagToken, html.SelfClosingTagToken:
			token := z.Token()
			isLinkTag := token.Data == "a" || token.Data == "link" || token.Data == "script"
			if isLinkTag {
				for _, attr := range token.Attr {
					if attr.Key == "href" || attr.Key == "src" {
						resolvedUrl, err := resolveURL(baseUrl, attr.Val)
						if err != nil {
							logger.Debugf("Error resolving URL %s for %s: %v\n", attr.Val, baseUrl, err)
							continue
						}
						logger.Debugf("Resolved URL: %s\n", resolvedUrl)
						if !strings.HasSuffix(resolvedUrl, ".css") && !strings.HasSuffix(resolvedUrl, ".png") {
							urls = append(urls, resolvedUrl)
						}
					}
				}
			}
		}

	}
}
