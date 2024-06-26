package smuggle

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"math/rand"

	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/logger"
)

// Two source of inputs:
// 1. static image file like svg or png or jpg or gif from crawler
// 2. static file from divider. (checked by cache signature)
func CheckCl0(urlstr string) {

	//Due to the nature of the check, we need to use a custom client for each goroutine
	client := &http.Client{
		Transport: clients.WrapTransport(clients.KeepAliveh1Transport()),
		Timeout:   120 * time.Second,
	}
	defer client.CloseIdleConnections()

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
		logger.Debugf("Cache hit for %s\n", urlstr)
		return
	}

	proxy := clients.Proxies[rand.Intn(len(clients.Proxies))]
	ctx := context.WithValue(context.Background(), "proxy", proxy)
	postBody := "GET /HopefullyMustBe404 HTTP/1.1\r\nFoo: x"
	postRequest, err := http.NewRequestWithContext(ctx, "POST", urlstr, strings.NewReader(postBody))
	if err != nil {
		logger.Warnf("Error creating POST request: %v\n", err)
		return
	}

	getRequest, err := http.NewRequestWithContext(ctx, "GET", urlstr, nil)
	if err != nil {
		logger.Warnf("Error creating GET request: %v\n", err)
		return
	}
	getRequest.Header.Set("Connection", "close")
	respChan := clients.AddToQueue(getRequest.Host, []*http.Request{postRequest, getRequest}, client)
	responses := <-respChan
	postResponse := responses[0]
	getResponse := responses[1]
	if postResponse.Err != nil {
		logger.Debugf("Error performing POST request: %v\n", err)
		return
	}

	defer postResponse.Resp.Body.Close()

	if getResponse.Err != nil {
		logger.Debugf("Error performing GET request: %v\n", err)
		return
	}
	defer getResponse.Resp.Body.Close()
	fmt.Println(getRequest.URL.String())
	fmt.Println(postResponse.Resp.StatusCode)
	fmt.Println(getResponse.Resp.StatusCode)
	if getResponse.Resp.StatusCode == 404 && postResponse.Resp.StatusCode != 404 {
		logger.Warnf("Potential CL0: %s\n", urlstr)
	}
}
