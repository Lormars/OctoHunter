package modules

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/getter"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/octohunter/internal/notify"
)

func CheckRedirect(ctx context.Context, wg *sync.WaitGroup, opts *common.Opts) {
	defer wg.Done()
	if opts.Target != "none" {
		SingleRedirectCheck(opts)
	} else {
		multiplex.Conscan(ctx, SingleRedirectCheck, opts, opts.RedirectFile, "redirect", 5)
	}
}

func SingleRedirectCheck(opts *common.Opts) {
	if !cacher.CheckCache(opts.Target, "redirect") {
		return
	}
	logger.Debugln("SingleRedirectCheck module running")
	finalURL, err := getFinalURL(opts.Target)

	if err != nil {
		logger.Warnf("Error getting final URL: %v\n", err)
		return
	}

	logger.Debugf("finalURL: %s for original url: %s", finalURL, opts.Target)

	req, err := http.NewRequest("GET", finalURL.String(), nil)
	if err == nil {
		resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
		if err == nil {
			common.DividerP.PublishMessage(resp) //send new-found finalURL to divider
		}
	}

	checkUnusualLength(finalURL, opts)
	checkOpenRedirect(finalURL, opts)

}

func checkUnusualLength(finalURL *url.URL, opts *common.Opts) {
	length, err := getLength(opts.Target)
	if err != nil {
		return
	}
	if length > 1000 {
		msg := fmt.Sprintf("[Redirect] from %s to %s with length %d\n", opts.Target, finalURL.String(), length)
		color.Red(msg)
		if opts.Module.Contains("broker") {
			notify.SendMessage(msg)
			common.OutputP.PublishMessage(msg)
		}
	}
}

func checkOpenRedirect(finalURL *url.URL, opts *common.Opts) {
	parsedURL, err := url.Parse(opts.Target)
	if err != nil {
		logger.Warnf("Error parsing URL: %v\n", err)
		return
	}

	queries := parsedURL.Query()
	for key, values := range queries {
		for _, value := range values {
			//first attempt to base64 decode the value as the value might be encoded
			attemptDecode, err := base64.URLEncoding.DecodeString(value)
			if err == nil {
				value = string(attemptDecode)
			}
			//first check whether the finalURL's hostname exists in the original URL's query
			//this is necessary to filter out false positive on query parameters
			if strings.Contains(value, finalURL.Hostname()) {

				msg := fmt.Sprintf("[OR Suspect] from %s to %s on param %s\n", opts.Target, finalURL.String(), key)
				color.Red(msg)
				if opts.Module.Contains("broker") {
					notify.SendMessage(msg)
					common.OutputP.PublishMessage(msg)
				}

				parsedOriginalURL, err := url.Parse(opts.Target)
				if err != nil {
					logger.Warnf("Error parsing URL: %v\n", err)
					continue
				}
				originalQueries := parsedOriginalURL.Query()
				var newValue string
				//replace the value with example.com based on the scheme
				if strings.HasPrefix(value, "http://") {
					newValue = "http://example.com"
				} else if strings.HasPrefix(value, "https://") {
					newValue = "https://example.com"
				} else {
					newValue = "example.com"
				}
				originalQueries.Set(key, newValue)
				parsedOriginalURL.RawQuery = originalQueries.Encode()
				req, err := http.NewRequest("GET", parsedOriginalURL.String(), nil)
				if err != nil {
					logger.Warnf("Error creating request: %v\n", err)
					continue
				}
				resp, err := checker.CheckServerCustom(req, clients.NormalClient)
				if err != nil {
					logger.Warnf("Error getting response from %s: %v\n", parsedOriginalURL.String(), err)
					continue
				}
				//since example.com contains "illustrative examples", we can check for that
				if strings.Contains(resp.Body, "illustrative examples") {
					msg := fmt.Sprintf("[OR Confirmed] from %s to %s on param %s\n", opts.Target, finalURL.String(), key)
					color.Red(msg)
					if opts.Module.Contains("broker") {
						notify.SendMessage(msg)
						common.OutputP.PublishMessage(msg)
					}
				}
			}
		}
	}
}

func getLength(url string) (int, error) {
	length, err := getter.GetHeader(url, "Content-Length")
	if err != nil {
		logger.Debugf("Error getting content length: %v\n", err)
		return 0, err
	}
	length_i, err := strconv.Atoi(length)
	if err != nil {
		logger.Warnf("Error converting length to int: %v\n", err)
		return 0, err
	}
	return length_i, nil
}

func getFinalURL(initialURL string) (*url.URL, error) {
	req, err := http.NewRequest("GET", initialURL, nil)
	if err != nil {
		logger.Warnf("Error creating request: %v", err)
		return nil, err
	}
	resp, err := checker.CheckServerCustom(req, clients.NormalClient)
	if err != nil {
		logger.Warnf("Error getting response from %s: %v\n", initialURL, err)
		return nil, err
	}

	finalURL := resp.FinalUrl

	return finalURL, nil

}
