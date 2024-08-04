package checker

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
)

func CheckAccess(resp *common.ServerResult) bool {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}

// check if the status code is between 300 and 400
func CheckRedirect(statusCode int) bool {
	if statusCode >= 300 && statusCode < 400 {
		return true
	}
	return false
}

// check if the status code is between 400 and 500
func CheckRequestError(statusCode int) bool {
	if statusCode >= 400 && statusCode < 500 && statusCode != 429 && statusCode != 404 {
		return true
	}
	return false
}

func Check405(resp *common.ServerResult) bool {
	return resp.StatusCode == 405
}

func Check429(resp *common.ServerResult) bool {
	return resp.StatusCode == 429
}

// Helper function to parse and check if MIME type is HTML
func CheckMimeType(contentType, mimeToCheck string) bool {
	// Split Content-Type header to get MIME type
	mimeType := contentType
	if idx := strings.Index(contentType, ";"); idx != -1 {
		mimeType = contentType[:idx]
	}

	mimeType = strings.TrimSpace(mimeType)
	// Check if MIME type is "text/html"
	return mimeType == mimeToCheck
}

// check if the url contains no path
func CheckHomePage(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	return parsedURL.Path == "/" || parsedURL.Path == ""
}

func CheckHttpRedirectToHttps(urlStr string) bool {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return false
	}
	resp, err := CheckServerCustom(req, clients.Clients.GetRandomClient("h0", false, true))
	if err != nil {
		return false
	}
	if CheckRedirect(resp.StatusCode) {
		locationHeader := resp.Headers.Get("Location")
		if strings.HasPrefix(locationHeader, "https://") {
			return true
		}
	}
	return false

}
