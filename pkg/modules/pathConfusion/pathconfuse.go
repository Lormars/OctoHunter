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
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

var encodings = []string{"/", "%0A", "%3B", "%23", "%3F"}
var buster = "vq8bo0zb3.css"
var cache = "bu90vqmpq.css"

func CheckPathConfusion(urlStr string) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Debugf("Error parsing url %v", err)
		return
	}

	//must have a path to work (yeah this is path confusion, how would it work without a path to confuse)
	//no need to check js files as they are not likely to contain private information
	if parsedURL.Path == "" || strings.HasSuffix(parsedURL.Path, ".js") {
		return
	}

	for _, encoding := range encodings {
		payload := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path + encoding + buster
		req, err := http.NewRequest("GET", payload, nil)
		if err != nil {
			logger.Debugf("Error creating request: %v", err)
			continue
		}

		resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
		if err != nil {
			logger.Debugf("Error getting response from %s: %v\n", payload, err)
			continue
		}
		if resp.StatusCode == 200 {
			cachePayload := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path + encoding + cache
			if checker.CheckCacheable(cachePayload) {
				msg := fmt.Sprintf("[WCD Suspect] Found using %s", cachePayload)
				color.Red(msg)
				common.OutputP.PublishMessage(msg)
				notify.SendMessage(msg)
				break
			}
		}
	}
}
