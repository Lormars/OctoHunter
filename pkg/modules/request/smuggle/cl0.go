package smuggle

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/logger"
)

// Two source of inputs:
// 1. static image file like svg or png or jpg or gif from crawler
// 2. static file from divider. (checked by cache signature)
func CheckCl0(urlstr string) {
	//check cache on path

	logger.Debugf("Checking for CL0 on %s\n", urlstr)
	parsedURL, err := url.Parse(urlstr)
	if err != nil {
		logger.Warnf("Error parsing URL: %v\n", err)
		return
	}

	hostName := parsedURL.Hostname()
	fullPath := parsedURL.Path
	dir := path.Dir(fullPath)
	cachePath := path.Join(hostName, dir)

	if !cacher.CheckCache(cachePath, "cl0") {
		logger.Infof("Cache hit for %s\n", urlstr)
		return
	}

	GetHostMutex(hostName).Lock()
	defer GetHostMutex(hostName).Unlock()
	postBody := "GET /HopefullyMustBe404 HTTP/1.1\r\nFoo: x"
	postRequest, err := http.NewRequest("POST", urlstr, strings.NewReader(postBody))
	if err != nil {
		logger.Warnf("Error creating POST request: %v\n", err)
		return
	}
	postResponse, err := clients.KeepAliveh1Client.Do(postRequest)
	if err != nil {
		logger.Warnf("Error performing POST request: %v\n", err)
		return
	}

	defer postResponse.Body.Close()
	getRequest, err := http.NewRequest("GET", urlstr, nil)
	if err != nil {
		logger.Warnf("Error creating GET request: %v\n", err)
		return
	}
	getRequest.Header.Set("Connection", "close")

	getResponse, err := clients.KeepAliveh1Client.Do(getRequest)
	if err != nil {
		logger.Warnf("Error performing GET request: %v\n", err)
		return
	}
	defer getResponse.Body.Close()
	if getResponse.StatusCode == 404 {
		logger.Warnf("Potential CL0: %s\n", urlstr)
	}

}
