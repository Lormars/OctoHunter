package pathconfusion

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

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

	common.AddToCrawlMap(urlStr, "pathconfusion", 200) //TODO: can be accurate

	var wg sync.WaitGroup

	for _, encoding := range encodings {
		wg.Add(1)
		go func(encoding string) {
			defer wg.Done()
			signature1, err := generator.GenerateSignature()
			if err != nil {
				logger.Debugf("Error generating signature: %v\n", err)
				return
			}

			payload1 := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path + encoding + signature1 + ".css"
			req1, err := http.NewRequest("GET", payload1, nil)
			if err != nil {
				logger.Debugf("Error creating request: %v", err)
				return
			}
			elapse1, resp1, err := checker.MeasureElapse(req1, clients.NoRedirectClient)
			if err != nil {
				logger.Debugf("Error getting response from %s: %v\n", payload1, err)
				return
			}

			if resp1.StatusCode == 200 {
				signature2, err := generator.GenerateSignature()
				if err != nil {
					logger.Debugf("Error generating signature: %v\n", err)
					return
				}

				payload2 := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path + encoding + signature2 + ".css"

				req2, err := http.NewRequest("GET", payload2, nil)
				if err != nil {
					logger.Debugf("Error creating request: %v", err)
					return
				}

				resp2, err := checker.CheckServerCustom(req2, clients.NoRedirectClient)
				if err != nil {
					logger.Debugf("Error getting response from %s: %v\n", payload2, err)
					return
				}

				same := resp1.Body == resp2.Body
				//if the response are different and the first request is not cached
				//Cache is checked either in the header (it has cache and miss) or if there is nothing in the header.
				if !same && ((matcher.HeaderKeyContainsSignature(resp1, "cache") && matcher.HeaderValueContainsSignature(resp1, "miss")) || !matcher.HeaderKeyContainsSignature(resp1, "cache")) {

					elapse2, resp2, err := checker.MeasureElapse(req1, clients.NoRedirectClient)
					if err != nil {
						logger.Debugf("Error getting response from %s: %v\n", payload2, err)
						return
					}
					same, _ = comparer.CompareResponse(resp1, resp2)
					//if the response are the same and the second request is cached.
					//Cache is measured either in the header (cache hit) or in the response time
					if same && ((matcher.HeaderKeyContainsSignature(resp2, "cache") && matcher.HeaderValueContainsSignature(resp2, "hit")) || elapse1 > elapse2*2) {
						msg := fmt.Sprintf("[WCD Suspect] Found using %s", payload1)
						color.Red(msg)
						common.OutputP.PublishMessage(msg)
						notify.SendMessage(msg)
					}
				}

			}
		}(encoding)
	}

	wg.Wait()
}
