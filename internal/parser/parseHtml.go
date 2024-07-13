package parser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
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
			isLinkTag := token.Data == "a" || token.Data == "link" || token.Data == "script" || token.Data == "img"
			if isLinkTag {
				for _, attr := range token.Attr {
					if attr.Key == "href" || attr.Key == "src" {
						resolvedUrl, err := resolveURL(baseUrl, attr.Val)
						if err != nil {
							logger.Debugf("Error resolving URL %s for %s: %v\n", attr.Val, baseUrl, err)
							continue
						}
						logger.Debugf("Resolved URL: %s\n", resolvedUrl)
						if !strings.HasSuffix(resolvedUrl, ".css") {
							urls = append(urls, resolvedUrl)
						}
					}
				}
			}
		}

	}
}

func ExtractSignature(htmlBody, signature string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlBody))
	if err != nil {
		logger.Warnf("Error parsing HTML: %v\n", err)
		return
	}

	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		for _, attr := range s.Nodes[0].Attr {
			if strings.Contains(attr.Val, signature) {
				//found in attribute
			}
		}

		if strings.Contains(s.Text(), signature) {
			//found between tags
		}
	})
}
