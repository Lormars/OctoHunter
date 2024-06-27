package modules

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	common.DividerP.PublishMessage(finalURL.String()) //send new-found finalURL to divider

	length, err := getLength(opts.Target)
	if err != nil {
		return
	}
	if length > 1000 {
		msg := fmt.Sprintf("[Redirect] from %s to %s\n", opts.Target, finalURL.String())
		color.Red(msg)
		if opts.Module.Contains("broker") {
			notify.SendMessage(msg)
			common.OutputP.PublishMessage(msg)
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
