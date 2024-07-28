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

func InitBrowser(headless bool) {
	l := launcher.New().Headless(headless).MustLaunch()
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

	if strings.HasPrefix(req.URL.String(), "http://") {
		return nil, fmt.Errorf("http request")
	}

	logger.Debugf("RequestWithBrowser: %s", req.URL.String())
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
				logger.Debugf("Setting header: %s: %s", key, v)
				ctx.Request.Req().Header.Set(key, v)
			}
		}
		c := ctx.Request.Req().Context()
		c = context.WithValue(c, "browser", true)
		ctx.Request.SetContext(c)

		logger.Debugf("Request: %s", ctx.Request.URL())
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

	err := page.Navigate(req.URL.String())
	if err != nil {
		mu.Lock()
		if !guard {
			guard = true
			wg.Done()
			close(done)
		}
		mu.Unlock()
		return nil, err
	}
	select {
	case <-done:
		// Request completed normally
	case <-timeout:
		// Timeout occurred
		mu.Lock()
		if !guard {
			guard = true
			wg.Done()
		}
		mu.Unlock()
		logger.Debugf("Timeout: %s", req.URL.String())
	}

	wg.Wait()
	if result == nil {
		return nil, fmt.Errorf("no result")
	}

	return result, nil
}
