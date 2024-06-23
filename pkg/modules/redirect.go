package modules

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/getter"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/octohunter/internal/notify"
)

var payload []string = []string{"admin", "dashboard", "user", "profile", "account", "portal", "home", "auth", "panel", "secure", "myaccount"}

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
		return
	}
	logger.Debugf("finalURL: %s for original url: %s", finalURL, opts.Target)
	common.DividerP.PublishMessage(finalURL.String()) //send new-found finalURL to divider
	if strings.Contains(finalURL.Path, "login") {
		for _, p := range payload {
			testPath := strings.Replace(finalURL.Path, "login", p, -1)
			newUrl := finalURL.Scheme + "://" + finalURL.Host + testPath
			newFinalURL, err := getFinalURL(newUrl)
			if err != nil {
				continue
			}
			logger.Debugln("newFinalURL.Path: ", newFinalURL.Path)
			if newFinalURL.Path == finalURL.Path {
				logger.Debugln("newFinalURL.Path == finalURL.Path for: ", newFinalURL.String())
				length, err := getter.GetHeader(newUrl, "Content-Length")
				if err != nil {
					logger.Debugf("Error getting content length: %v\n", err)
					continue
				}
				length_i, err := strconv.Atoi(length)
				if err != nil {
					logger.Debugf("Error converting length to int: %v\n", err)
					continue
				}
				if length_i > 1000 {
					msg := fmt.Sprintf("[Redirect] from %s to %s\n", newUrl, finalURL.String())
					color.Red(msg)
					if opts.Module.Contains("broker") {
						notify.SendMessage(msg)
						common.OutputP.PublishMessage(msg)
					}
				}
			} else {
				common.DividerP.PublishMessage(newFinalURL.String()) //send new-found finalURL to divider
			}
		}
		//if the redirected path is login page, but using oauth, try to bypass using different method
	} else if strings.Contains(finalURL.Path, "oauth") {
		common.MethodP.PublishMessage(opts.Target) //send new-found finalURL to method to try bypass
	}
}

func getFinalURL(initialURL string) (*url.URL, error) {

	resp, err := clients.NormalClient.Get(initialURL)
	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", initialURL, err)
		return nil, err
	}

	defer resp.Body.Close()

	finalURL := resp.Request.URL

	return finalURL, nil

}
