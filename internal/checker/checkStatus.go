package checker

import (
	"net/url"
	"strings"

	"github.com/lormars/octohunter/common"
)

func CheckAccess(resp *common.ServerResult) bool {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}

func CheckRedirect(statusCode int) bool {
	if statusCode >= 300 && statusCode < 400 {
		return true
	}
	return false
}

func CheckRequestError(statusCode int) bool {
	if statusCode >= 400 && statusCode < 500 {
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

func CheckHomePage(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	return parsedURL.Path == "/" || parsedURL.Path == ""
}
