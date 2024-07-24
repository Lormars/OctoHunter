package mime

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/notify"
)

// only accept input from 200-300 response for now
func CheckMime(result *common.ServerResult) {
	targetURL := result.Url
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return
	}
	parameterlessURL := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
	if !cacher.CheckCache(parameterlessURL, "mime") {
		return
	}

	go tryManipulate404Mime(parameterlessURL, parsedURL.Path)

}

func tryManipulate404Mime(urlStr, path string) {
	var payloads []string
	//need to check if there are paths in the url. If not, can only check www.abc.com/nonexistent.xml
	//if yes, can also check www.abc.com/path1nonexistent.xml
	if path == "" || path == "/" {
		payloads = []string{
			"/wbpbq.xml",
		}
	} else {
		payloads = []string{
			"/wbpbq.xml",
			"wbpbq.xml",
		}
	}

	for _, payload := range payloads {
		fuzzURL := strings.TrimRight(urlStr, "/") + payload
		req, err := http.NewRequest("GET", fuzzURL, nil)
		if err != nil {
			continue
		}
		resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
		if err != nil {
			continue
		}
		if resp.StatusCode != 404 {
			continue
		}
		if checker.CheckMimeType(resp.Headers.Get("Content-Type"), "application/xml") ||
			checker.CheckMimeType(resp.Headers.Get("Content-Type"), "text/xml") {
			msg := "[MIME] Possible MIME confusion: 404 page with XML mime: " + fuzzURL
			// common.AddToCrawlMap(fuzzURL, "mime", resp.StatusCode)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			notify.SendMessage(msg)
		}
	}
}
