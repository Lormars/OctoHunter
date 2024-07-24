package common

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
	"github.com/lormars/octohunter/internal/logger"
)

func RequestWithBrowser(req *http.Request, client *http.Client) (*http.Response, error) {
	logger.Warnf("RequestWithBrowser: %s", req.URL.String())
	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := stealth.MustPage(browser)

	router := page.HijackRequests()
	defer router.MustStop()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	var result *http.Response

	pattern := req.URL.String() + "?"

	router.MustAdd(pattern, func(ctx *rod.Hijack) {

		for key, values := range req.Header {
			for _, v := range values {
				ctx.Request.Req().Header.Set(key, v)
			}
		}

		logger.Warnf("Request: %s", ctx.Request.URL())
		err := ctx.LoadResponse(client, true)
		if err != nil {
			logger.Warnf("Error loading response: %v", err)
		} else {
			result = &http.Response{
				Status:     ctx.Response.RawResponse.Status,
				StatusCode: ctx.Response.RawResponse.StatusCode,
				Body:       io.NopCloser(strings.NewReader(ctx.Response.Body())),
				Header:     ctx.Response.RawResponse.Header,
				Request:    req,
			}
		}
		wg.Done()
	})

	go router.Run()
	logger.Warnf("Navigating to: %s vs %s", req.URL.String(), pattern)
	page.MustNavigate(req.URL.String()).MustWaitIdle()
	wg.Wait()
	if result == nil {
		return nil, fmt.Errorf("no result")
	}

	return result, nil

}
