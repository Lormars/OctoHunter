package common

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
	"github.com/lormars/octohunter/internal/logger"
)

var browser *rod.Browser
var pool rod.Pool[rod.Page]
var create func() *rod.Page

func init() {
	l := launcher.New().Headless(false).MustLaunch()
	browser = rod.New().ControlURL(l).MustConnect()

	// browser = rod.New().MustConnect()
	browser.IgnoreCertErrors(true)
	pool = rod.NewPagePool(10)
	create = func() *rod.Page {
		return stealth.MustPage(browser)
	}
}

func CloseBrowser() {
	browser.MustClose()
}

func RequestWithBrowser(req *http.Request, client *http.Client) (*http.Response, error) {

	logger.Warnf("RequestWithBrowser: %s", req.URL.String())
	page := pool.MustGet(create)
	defer pool.Put(page)
	timeout := time.After(10 * time.Second)
	done := make(chan struct{})
	// page = page.Timeout(15 * time.Second)
	router := page.HijackRequests()
	defer router.MustStop()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	mu := sync.Mutex{}
	var result *http.Response
	pattern := req.URL.String()
	guard := false
	router.MustAdd(pattern+"*", func(ctx *rod.Hijack) {
		for key, values := range req.Header {
			for _, v := range values {
				logger.Warnf("Setting header: %s: %s", key, v)
				ctx.Request.Req().Header.Set(key, v)
			}
		}
		c := ctx.Request.Req().Context()
		c = context.WithValue(c, "browser", true)
		ctx.Request.SetContext(c)

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
		mu.Lock()
		if !guard {
			guard = true
			wg.Done()
			close(done)
		}
		mu.Unlock()
	})

	go router.Run()

	page.MustNavigate(req.URL.String())
	select {
	case <-done:
		// Request completed normally
	case <-timeout:
		// Timeout occurred
		mu.Lock()
		wg.Done() // Ensure we call Done in case of timeout
		guard = true
		mu.Unlock()
		logger.Warnf("Timeout: %s", req.URL.String())
	}

	wg.Wait()
	if result == nil {
		return nil, fmt.Errorf("no result")
	}

	return result, nil
}
