package pathconfusion

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/comparer"
	"github.com/lormars/octohunter/internal/generator"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/matcher"
	"github.com/lormars/octohunter/internal/notify"
)

var encodings = []string{"/", "%0A", "%3B", "%23", "%3Fname=val", "%2F", "%25%30%41", "25%30%30",
	"%25%33%46", "%25%33%42", "%25%32%33", "%25%32%46"}

func CheckPathConfusion(urlStr string) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Debugf("Error parsing url %v", err)
		return
	}

	//must have a path to work (yeah this is path confusion, how would it work without a path to confuse)
	//no need to check js files as they are not likely to contain private information
	if parsedURL.Path == "" ||
		strings.HasSuffix(parsedURL.Path, ".js") ||
		strings.HasSuffix(parsedURL.Path, ".css") ||
		strings.HasSuffix(parsedURL.Path, ".svg") ||
		strings.HasSuffix(parsedURL.Path, ".png") ||
		strings.HasSuffix(parsedURL.Path, ".jpg") ||
		strings.HasSuffix(parsedURL.Path, ".gif") ||
		strings.HasSuffix(parsedURL.Path, ".jpeg") {
		return
	}

	for _, encoding := range encodings {
		signature1, err := generator.GenerateSignature()
		if err != nil {
			logger.Debugf("Error generating signature: %v\n", err)
			continue
		}

		payload1 := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path + encoding + signature1 + ".css"
		req1, err := http.NewRequest("GET", payload1, nil)
		if err != nil {
			logger.Debugf("Error creating request: %v", err)
			continue
		}

		resp1, err := checker.CheckServerCustom(req1, clients.NoRedirectClient)
		if err != nil {
			logger.Debugf("Error getting response from %s: %v\n", payload1, err)
			continue
		}

		if resp1.StatusCode == 200 {
			signature2, err := generator.GenerateSignature()
			if err != nil {
				logger.Debugf("Error generating signature: %v\n", err)
				continue
			}

			payload2 := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path + encoding + signature2 + ".css"

			req2, err := http.NewRequest("GET", payload2, nil)
			if err != nil {
				logger.Debugf("Error creating request: %v", err)
				continue
			}

			resp2, err := checker.CheckServerCustom(req2, clients.NoRedirectClient)
			if err != nil {
				logger.Debugf("Error getting response from %s: %v\n", payload2, err)
				continue
			}

			same, _ := comparer.CompareResponse(resp1, resp2)
			if !same && matcher.HeaderKeyContainsSignature(resp1, "cache") && matcher.HeaderValueContainsSignature(resp1, "miss") {
				resp2, err = checker.CheckServerCustom(req1, clients.NoRedirectClient)
				if err != nil {
					logger.Debugf("Error getting response from %s: %v\n", payload2, err)
					continue
				}
				same, _ = comparer.CompareResponse(resp1, resp2)
				if same && matcher.HeaderKeyContainsSignature(resp2, "cache") && matcher.HeaderValueContainsSignature(resp2, "hit") {
					msg := fmt.Sprintf("[WCD Suspect] Found using %s", payload1)
					color.Red(msg)
					common.OutputP.PublishMessage(msg)
					notify.SendMessage(msg)
					break
				}
			}

		}
	}
}
